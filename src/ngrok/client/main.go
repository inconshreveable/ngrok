package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"ngrok/client/ui"
	"ngrok/client/views/term"
	"ngrok/client/views/web"
	"ngrok/conn"
	"ngrok/log"
	"ngrok/msg"
	"ngrok/proto"
	"ngrok/util"
	"ngrok/version"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	pingInterval         = 20 * time.Second
	maxPongLatency       = 15 * time.Second
	versionCheckInterval = 6 * time.Hour
	versionEndpoint      = "http://dl.ngrok.com/versions"
	BadGateway           = `<html>
<body style="background-color: #97a8b9">
    <div style="margin:auto; width:400px;padding: 20px 60px; background-color: #D3D3D3; border: 5px solid maroon;">
        <h2>Tunnel %s unavailable</h2>
        <p>Unable to initiate connection to <strong>%s</strong>. A web server must be running on port <strong>%s</strong> to complete the tunnel.</p>
`
)

/**
 * Establishes and manages a tunnel proxy connection with the server
 */
func proxy(proxyAddr string, s *State, ctl *ui.Controller) {
	start := time.Now()
	remoteConn, err := conn.Dial(proxyAddr, "pxy", tlsConfig)
	if err != nil {
		// XXX: What is the proper response here?
		// display something to the user?
		// retry?
		// reset control connection?
		log.Error("Failed to establish proxy connection: %v", err)
		return
	}

	defer remoteConn.Close()
	err = msg.WriteMsg(remoteConn, &msg.RegProxyMsg{Url: s.publicUrl})
	if err != nil {
		// XXX: What is the proper response here?
		// display something to the user?
		// retry?
		// reset control connection?
		log.Error("Failed to write RegProxyMsg: %v", err)
		return
	}

	localConn, err := conn.Dial(s.opts.localaddr, "prv", nil)
	if err != nil {
		remoteConn.Warn("Failed to open private leg %s: %v", s.opts.localaddr, err)
		badGatewayBody := fmt.Sprintf(BadGateway, s.publicUrl, s.opts.localaddr, s.opts.localaddr)
		remoteConn.Write([]byte(fmt.Sprintf(`HTTP/1.0 502 Bad Gateway
Content-Type: text/html
Content-Length: %d

%s`, len(badGatewayBody), badGatewayBody)))
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

func versionCheck(s *State, ctl *ui.Controller) {
	check := func() {
		resp, err := http.Get(versionEndpoint)
		if err != nil {
			log.Warn("Failed to get version info %s: %v", versionEndpoint, err)
			return
		}
		defer resp.Body.Close()

		var payload struct {
			Client struct {
				Version string
			}
		}

		err = json.NewDecoder(resp.Body).Decode(&payload)
		if err != nil {
			log.Warn("Failed to read version info: %v", err)
			return
		}

		if payload.Client.Version != version.MajorMinor() {
			s.newVersion = payload.Client.Version
			ctl.Update(s)
		}
	}

	// check immediately and then at a set interval
	check()
	for _ = range time.Tick(versionCheckInterval) {
		check()
	}
}

/*
 * Hearbeating to ensure our connection ngrokd is still live
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

func reconnectingControl(s *State, ctl *ui.Controller) {
	// how long we should wait before we reconnect
	maxWait := 30 * time.Second
	wait := 1 * time.Second

	for {
		control(s, ctl)

		if s.status == "online" {
			wait = 1 * time.Second
		}

		log.Info("Waiting %d seconds before reconnecting", int(wait.Seconds()))
		time.Sleep(wait)
		// exponentially increase wait time
		wait = 2 * wait
		wait = time.Duration(math.Min(float64(wait), float64(maxWait)))
		s.status = "reconnecting"
		ctl.Update(s)
	}
}

/**
 * Establishes and manages a tunnel control connection with the server
 */
func control(s *State, ctl *ui.Controller) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("control recovering from failure %v", r)
		}
	}()

	// establish control channel
	conn, err := conn.Dial(s.opts.server, "ctl", tlsConfig)
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
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
		User:      s.opts.authtoken,
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
		ctl.Cmds <- ui.CmdQuit{Message: emsg}
		return
	}

	// update UI state
	conn.Info("Tunnel established at %v", regAck.Url)
	s.publicUrl = regAck.Url
	s.status = "online"
	s.serverVersion = regAck.MmVersion
	ctl.Update(s)

	SaveAuthToken(s.opts.authtoken)

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
	// parse options
	opts := parseArgs()

	// set up logging
	log.LogTo(opts.logto)

	// set up auth token
	if opts.authtoken == "" {
		opts.authtoken = LoadAuthToken()
	}

	// init client state
	s := &State{
		status: "connecting",

		// unique client id
		id: util.RandIdOrPanic(8),

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
	web.NewWebView(ctl, s, opts.webport)
	if opts.logto != "stdout" {
		term.New(ctl, s)
	}

	go reconnectingControl(s, ctl)
	go versionCheck(s, ctl)

	quitMessage := ""
	ctl.Wait.Add(1)
	go func() {
		defer ctl.Wait.Done()
		for {
			select {
			case obj := <-ctl.Cmds:
				switch cmd := obj.(type) {
				case ui.CmdQuit:
					quitMessage = cmd.Message
					ctl.DoShutdown()
					return
				case ui.CmdRequest:
					go func() {
						var localConn conn.Conn
						localConn, err := conn.Dial(s.opts.localaddr, "prv", nil)
						if err != nil {
							log.Warn("Failed to open private leg %s: %v", s.opts.localaddr, err)
							return
						}
						//defer localConn.Close()
						localConn = s.protocol.WrapConn(localConn)
						localConn.Write(cmd.Payload)
						ioutil.ReadAll(localConn)
					}()
				}
			}
		}
	}()

	ctl.Wait.Wait()
	fmt.Println(quitMessage)
}
