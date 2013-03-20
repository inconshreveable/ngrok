package server

import (
	"fmt"
	"net"
	"ngrok"
	"ngrok/conn"
	"ngrok/proto"
)

/**
 * Tunnel: A control connection, metadata and proxy connections which
 *         route public traffic to a firewalled endpoint.
 */
type Tunnel struct {
	regMsg *proto.RegMsg

	// public url
	url string

	// tcp listener
	listener *net.TCPListener

	// control connection
	ctl *Control

	// proxy connections
	proxies chan conn.Conn

	// logger
	ngrok.Logger
}

func newTunnel(msg *proto.RegMsg, ctl *Control) {
	t := &Tunnel{
		regMsg:  msg,
		ctl:     ctl,
		proxies: make(chan conn.Conn),
		Logger:  ngrok.NewPrefixLogger(),
	}

	switch t.regMsg.Protocol {
	case "tcp":
		var err error
		t.listener, err = net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), 0})

		if err != nil {
			panic(err)
		}

		go t.listenTcp(t.listener)

	default:
	}

	tunnels.Add(t)
	t.ctl.conn.AddLogPrefix(t.Id())
	t.AddLogPrefix(t.Id())
	t.Info("Registered new tunnel")
	t.ctl.out <- &proto.RegAckMsg{Url: t.url, ProxyAddr: fmt.Sprintf("%s", proxyAddr)}

	//go t.managerThread()
}

func (t *Tunnel) shutdown() {
	t.Info("Shutting down")
	// stop any go routines
	// close all proxy and public connections
	// stop any metrics
	t.ctl.stop <- 1
}

func (t *Tunnel) Id() string {
	return t.url
}

func (t *Tunnel) managerThread() {
}

/**
 * Listens for new public tcp connections from the internet.
 */
func (t *Tunnel) listenTcp(listener *net.TCPListener) {
	for {
		// accept public connections
		tcpConn, err := listener.AcceptTCP()

		if err != nil {
			panic(err)
		}

		conn := conn.NewLogged(tcpConn, "pub")
		conn.AddLogPrefix(t.Id())

		go func() {
			defer func() {
				if r := recover(); r != nil {
					conn.Warn("Failed with error %v", r)
				}
			}()
			defer conn.Close()

			t.HandlePublicConnection(conn)
		}()
	}
}

func (t *Tunnel) HandlePublicConnection(publicConn conn.Conn) {
	metrics.requestTimer.Time(func() {
		metrics.requestMeter.Mark(1)

		t.Debug("Requesting new proxy connection")
		t.ctl.out <- &proto.ReqProxyMsg{}

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
