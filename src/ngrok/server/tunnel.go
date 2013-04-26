package server

import (
	"encoding/base64"
	"fmt"
	"net"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"ngrok/version"
)

/**
 * Tunnel: A control connection, metadata and proxy connections which
 *         route public traffic to a firewalled endpoint.
 */
type Tunnel struct {
	regMsg *msg.RegMsg

	// public url
	url string

	// tcp listener
	listener *net.TCPListener

	// control connection
	ctl *Control

	// proxy connections
	proxies chan conn.Conn

	// logger
	log.Logger
}

func newTunnel(m *msg.RegMsg, ctl *Control) (t *Tunnel) {
	t = &Tunnel{
		regMsg:  m,
		ctl:     ctl,
		proxies: make(chan conn.Conn),
		Logger:  log.NewPrefixLogger(),
	}

	switch t.regMsg.Protocol {
	case "tcp":
		var err error
		t.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0})

		if err != nil {
			panic(err)
		}

		go t.listenTcp(t.listener)

	default:
	}

	if err := tunnels.Add(t); err != nil {
		t.ctl.stop <- &msg.RegAckMsg{Error: fmt.Sprint(err)}
		return
	}

	if m.Version != version.Proto {
		t.ctl.stop <- &msg.RegAckMsg{Error: fmt.Sprintf("Incompatible versions. Server %s, client %s. Download a new version at http://ngrok.com", version.MajorMinor(), m.Version)}
	}

	// pre-encode the http basic auth for fast comparisons later
	if m.HttpAuth != "" {
		m.HttpAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(m.HttpAuth))
	}

	t.ctl.conn.AddLogPrefix(t.Id())
	t.AddLogPrefix(t.Id())
	t.Info("Registered new tunnel")
	t.ctl.out <- &msg.RegAckMsg{
		Url:       t.url,
		ProxyAddr: fmt.Sprintf("%s", proxyAddr),
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
	}

	return
}

func (t *Tunnel) shutdown() {
	t.Info("Shutting down")

	// if we have a public listener (this is a raw TCP tunnel, shut it down
	if t.listener != nil {
		t.listener.Close()
	}

	// remove ourselves from the tunnel registry
	tunnels.Del(t.url)

	// XXX: should we shut down all of the proxy connections?

	// XXX: will this block if this is being called from Control's shutdown code?
	t.ctl.stop <- nil
}

func (t *Tunnel) Id() string {
	return t.url
}

/**
 * Listens for new public tcp connections from the internet.
 */
func (t *Tunnel) listenTcp(listener *net.TCPListener) {
	for {
		defer func() {
			if r := recover(); r != nil {
				log.Warn("listenTcp failed with error %v", r)
			}
		}()

		// accept public connections
		tcpConn, err := listener.AcceptTCP()

		if err != nil {
			panic(err)
		}

		conn := conn.Wrap(tcpConn, "pub")
		conn.AddLogPrefix(t.Id())
		conn.Info("New connection from %v", conn.RemoteAddr())

		go t.HandlePublicConnection(conn)
	}
}

func (t *Tunnel) HandlePublicConnection(publicConn conn.Conn) {
	defer publicConn.Close()
	defer func() {
		if r := recover(); r != nil {
			publicConn.Warn("HandlePublicConnection failed with error %v", r)
		}
	}()

	metrics.requestTimer.Time(func() {
		metrics.requestMeter.Mark(1)

		t.Debug("Requesting new proxy connection")
		t.ctl.out <- &msg.ReqProxyMsg{}

		proxyConn := <-t.proxies
		t.Info("Returning proxy connection %s", proxyConn.Id())

		defer proxyConn.Close()
		conn.Join(publicConn, proxyConn)
	})
}

func (t *Tunnel) RegisterProxy(conn conn.Conn) {
	t.Info("Registered proxy connection %s", conn.Id())
	conn.AddLogPrefix(t.Id())
	t.proxies <- conn
}
