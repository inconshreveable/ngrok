package client

import (
	log "code.google.com/p/log4go"
	"crypto/rand"
	"fmt"
	"net"
	"ngrok/client/ui"
	"ngrok/conn"
	nlog "ngrok/log"
	"ngrok/proto"
	"runtime"
	"time"
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
func proxy(proxyAddr string, s *State) {
	start := time.Now()
	remoteConn, err := connect(proxyAddr, "pxy")
	if err != nil {
		panic(err)
	}

	defer remoteConn.Close()
	err = proto.WriteMsg(remoteConn, &proto.RegProxyMsg{Url: s.publicUrl})

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
	s.Update()
	m.connTimer.Time(func() {
		if s.opts.protocol == "http" {
			teeConn := conn.NewTee(remoteConn)
			remoteConn = teeConn
			go conn.ParseHttp(teeConn, s.history.reqs, s.history.resps)
		}
		bytesIn, bytesOut := conn.Join(localConn, remoteConn)
		m.bytesIn.Update(bytesIn)
		m.bytesOut.Update(bytesOut)
		m.bytesInCount.Inc(bytesIn)
		m.bytesOutCount.Inc(bytesOut)
	})
	s.Update()
}

/**
 * Establishes and manages a tunnel control connection with the server
 */
func control(s *State) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovering from failure %v, attempting to reconnect to server after 10 seconds . . .", r)
			s.status = "reconnecting"
			s.Update()
			time.Sleep(10 * time.Second)
			go control(s)
		}
	}()

	// establish control channel
	conn, err := connect(s.opts.server, "ctl")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// register with the server
	err = proto.WriteMsg(conn, &proto.RegMsg{
		Protocol:  s.opts.protocol,
		OS:        runtime.GOOS,
		Hostname:  s.opts.hostname,
		Subdomain: s.opts.subdomain,
		ClientId:  s.id,
	})

	if err != nil {
		panic(err)
	}

	// wait for the server to ack our register
	var regAck proto.RegAckMsg
	if err = proto.ReadMsgInto(conn, &regAck); err != nil {
		panic(err)
	}

	if regAck.Error != "" {
		emsg := fmt.Sprintf("Server failed to allocate tunnel: %s", regAck.Error)
		s.ui.Cmds <- ui.Command{ui.QUIT, emsg}
		return
	}

	// update UI state
	conn.Info("Tunnel established at %v", regAck.Url)
	//state.version = regAck.Version
	s.publicUrl = regAck.Url
	s.status = "online"
	s.Update()

	// main control loop
	for {
		var msg proto.Message
		if msg, err = proto.ReadMsg(conn); err != nil {
			panic(err)
		}

		switch msg.GetType() {
		case "ReqProxyMsg":
			go proxy(regAck.ProxyAddr, s)

		case "PingMsg":
			proto.WriteMsg(conn, &proto.PongMsg{})
		}
	}
}

// create a random identifier for this client
func mkid() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create random client identifier, %v", err))
	}
	return fmt.Sprintf("%x", b)

}

func Main() {
	// XXX: should do this only if they ask us too
	nlog.LogToFile()

	// parse options
	opts := parseArgs()

	// init terminal, http UI
	termView := ui.NewTerm()
	httpView := ui.NewHttp(9999)

	// init client state
	s := &State{
		// unique client id
		id: mkid(),

		// ui communication channels
		ui: ui.NewUi(termView, httpView),
		//ui: ui.NewUi(httpView),

		// command-line options
		opts: opts,

		// metrics
		metrics: NewClientMetrics(),
	}

	// request history
	// XXX: don't use a callback, use a channel
	// and define it inline in the struct
	s.history = NewRequestHistory(opts.historySize, s.metrics, func(history []*RequestHistoryEntry) {
		s.historyEntries = history
		s.Update()
	})

	// set initial ui state
	s.status = "connecting"
	s.Update()

	go control(s)

	quitMessage := ""
	s.ui.Wait.Add(1)
	go func() {
		defer s.ui.Wait.Done()
		for {
			select {
			case cmd := <-s.ui.Cmds:
				switch cmd.Code {
				case ui.QUIT:
					quitMessage = cmd.Payload.(string)
					s.stopping = true
					s.Update()
					return
				}
			}
		}
	}()

	s.ui.Wait.Wait()
	fmt.Println(quitMessage)
}
