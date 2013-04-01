package client

import (
	log "code.google.com/p/log4go"
	"fmt"
	"io/ioutil"
	"net"
	"ngrok/client/ui"
	"ngrok/client/views/term"
	"ngrok/client/views/web"
	"ngrok/conn"
	nlog "ngrok/log"
	"ngrok/msg"
	"ngrok/proto"
	"ngrok/util"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	pingInterval   = 20 * time.Second
	maxPongLatency = 15 * time.Second
)

/** 
 * Connect to the ngrok server
 */
func connect(addr string, typ string) (c conn.Conn, err error) {
	var (
		tcpAddr *net.TCPAddr
		tcpConn *net.TCPConn
	)

	if tcpAddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return
	}

	log.Debug("Dialing %v", addr)
	if tcpConn, err = net.DialTCP("tcp", nil, tcpAddr); err != nil {
		return
	}

	c = conn.NewTCP(tcpConn, typ)
	c.Debug("Connected to: %v", tcpAddr)
	return c, nil
}

/**
 * Establishes and manages a tunnel proxy connection with the server
 */
func proxy(proxyAddr string, s *State, ctl *ui.Controller) {
	start := time.Now()
	remoteConn, err := connect(proxyAddr, "pxy")
	if err != nil {
		panic(err)
	}

	defer remoteConn.Close()
	err = msg.WriteMsg(remoteConn, &msg.RegProxyMsg{Url: s.publicUrl})

	if err != nil {
		panic(err)
	}

	localConn, err := connect(s.opts.localaddr, "prv")
	if err != nil {
		remoteConn.Warn("Failed to open private leg %s: %v", s.opts.localaddr, err)
		return
	}
	defer localConn.Close()

	m := s.metrics
	m.proxySetupTimer.Update(time.Since(start))
	m.connMeter.Mark(1)
	ctl.Update(s)
	m.connTimer.Time(func() {
		localConn := s.protocol.WrapConn(localConn)
		bytesIn, bytesOut := conn.Join(localConn, remoteConn)
		m.bytesIn.Update(bytesIn)
		m.bytesOut.Update(bytesOut)
		m.bytesInCount.Inc(bytesIn)
		m.bytesOutCount.Inc(bytesOut)
	})
	ctl.Update(s)
}

/*
 * Hearbeating ensure our connection ngrokd is still live
 */
func heartbeat(lastPongAddr *int64, c conn.Conn) {
	lastPing := time.Unix(atomic.LoadInt64(lastPongAddr)-1, 0)
	ping := time.NewTicker(pingInterval)
	pongCheck := time.NewTicker(time.Second)

	defer func() {
		c.Close()
		ping.Stop()
		pongCheck.Stop()
	}()

	for {
		select {
		case <-pongCheck.C:
			lastPong := time.Unix(0, atomic.LoadInt64(lastPongAddr))
			needPong := lastPong.Sub(lastPing) < 0
			pongLatency := time.Since(lastPing)

			if needPong && pongLatency > maxPongLatency {
				c.Info("Last ping: %v, Last pong: %v", lastPing, lastPong)
				c.Info("Connection stale, haven't gotten PongMsg in %d seconds", int(pongLatency.Seconds()))
				return
			}

		case <-ping.C:
			err := msg.WriteMsg(c, &msg.PingMsg{})
			if err != nil {
				c.Debug("Got error %v when writing PingMsg", err)
				return
			}
			lastPing = time.Now()
		}
	}
}

/**
 * Establishes and manages a tunnel control connection with the server
 */
func control(s *State, ctl *ui.Controller) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovering from failure %v, attempting to reconnect to server after 10 seconds . . .", r)
			s.status = "reconnecting"
			ctl.Update(s)
			time.Sleep(10 * time.Second)
			go control(s, ctl)
		}
	}()

	// establish control channel
	conn, err := connect(s.opts.server, "ctl")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// register with the server
	err = msg.WriteMsg(conn, &msg.RegMsg{
		Protocol:  s.opts.protocol,
		OS:        runtime.GOOS,
		HttpAuth:  s.opts.httpAuth,
		Hostname:  s.opts.hostname,
		Subdomain: s.opts.subdomain,
		ClientId:  s.id,
		Version:   msg.Version,
	})

	if err != nil {
		panic(err)
	}

	// wait for the server to ack our register
	var regAck msg.RegAckMsg
	if err = msg.ReadMsgInto(conn, &regAck); err != nil {
		panic(err)
	}

	if regAck.Error != "" {
		emsg := fmt.Sprintf("Server failed to allocate tunnel: %s", regAck.Error)
		ctl.Cmds <- ui.Command{ui.QUIT, emsg}
		return
	}

	// update UI state
	conn.Info("Tunnel established at %v", regAck.Url)
	s.publicUrl = regAck.Url
	s.status = "online"
	s.serverVersion = regAck.Version
	ctl.Update(s)

	// start the heartbeat
	lastPong := time.Now().UnixNano()
	go heartbeat(&lastPong, conn)

	// main control loop
	for {
		var m msg.Message
		if m, err = msg.ReadMsg(conn); err != nil {
			panic(err)
		}

		switch m.(type) {
		case *msg.ReqProxyMsg:
			go proxy(regAck.ProxyAddr, s, ctl)

		case *msg.PongMsg:
			atomic.StoreInt64(&lastPong, time.Now().UnixNano())
		}
	}
}

func Main() {
	// XXX: should do this only if they ask us too
	nlog.LogToFile()

	// parse options
	opts := parseArgs()

	// init client state
	s := &State{
		status: "connecting",

		// unique client id
		id: util.RandId(),

		// command-line options
		opts: opts,

		// metrics
		metrics: NewClientMetrics(),
	}

	switch opts.protocol {
	case "http":
		s.protocol = proto.NewHttp()
	case "tcp":
		s.protocol = proto.NewTcp()
	}

	// init ui
	ctl := ui.NewController()
	term.New(ctl, s)
	web.NewWebView(ctl, s, opts.webport)

	go control(s, ctl)

	quitMessage := ""
	ctl.Wait.Add(1)
	go func() {
		defer ctl.Wait.Done()
		for {
			select {
			case cmd := <-ctl.Cmds:
				switch cmd.Code {
				case ui.QUIT:
					quitMessage = cmd.Payload.(string)
					ctl.DoShutdown()
					return
				case ui.REPLAY:
					go func() {
						payload := cmd.Payload.([]byte)
						localConn, err := connect(s.opts.localaddr, "prv")
						if err != nil {
							log.Warn("Failed to open private leg %s: %v", s.opts.localaddr, err)
							return
						}
						//defer localConn.Close()
						localConn = s.protocol.WrapConn(localConn)
						localConn.Write(payload)
						ioutil.ReadAll(localConn)
					}()
				}
			}
		}
	}()

	ctl.Wait.Wait()
	fmt.Println(quitMessage)
}
