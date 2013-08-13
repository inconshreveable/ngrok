package server

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"ngrok/conn"
	"ngrok/log"
	"ngrok/msg"
	"ngrok/version"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

/**
 * Tunnel: A control connection, metadata and proxy connections which
 *         route public traffic to a firewalled endpoint.
 */
type Tunnel struct {
	regMsg *msg.RegMsg

	// time when the tunnel was opened
	start time.Time

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

	// closing
	closing int32
}

func newTunnel(m *msg.RegMsg, ctl *Control) (t *Tunnel) {
	t = &Tunnel{
		regMsg:  m,
		start:   time.Now(),
		ctl:     ctl,
		proxies: make(chan conn.Conn),
		Logger:  log.NewPrefixLogger(),
	}

	failReg := func(err error) {
		t.ctl.stop <- &msg.RegAckMsg{Error: err.Error()}
	}

	var err error

	switch t.regMsg.Protocol {
	case "tcp":
		var port int = 0

		// try to return to you the same port you had before
		cachedUrl := tunnels.GetCachedRegistration(t)
		if cachedUrl != "" {
			parts := strings.Split(cachedUrl, ":")
			portPart := parts[len(parts)-1]
			port, err = strconv.Atoi(portPart)
			if err != nil {
				t.ctl.conn.Error("Failed to parse cached url port as integer: %s", portPart)
				// continue with zero
				port = 0
			}
		}

		// Bind for TCP connections
		t.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})

		// If we failed with a custom port, try with a random one
		if err != nil && port != 0 {
			t.ctl.conn.Warn("Failed to get custom port %d: %v, trying a random one", port, err)
			t.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0})
		}

		// we tried to bind with a random port and failed (no more ports available?)
		if err != nil {
			failReg(t.ctl.conn.Error("Error binding TCP listener: %v", err))
			return
		}

		// create the url
		addr := t.listener.Addr().(*net.TCPAddr)
		t.url = fmt.Sprintf("tcp://%s:%d", domain, addr.Port)

		// register it
		if err = tunnels.RegisterAndCache(t.url, t); err != nil {
			// This should never be possible because the OS will
			// only assign available ports to us.
			t.listener.Close()
			failReg(fmt.Errorf("TCP listener bound, but failed to register %s", t.url))
			return
		}

		go t.listenTcp(t.listener)

	case "http":
		vhost := os.Getenv("VHOST")
		if vhost == "" {
			vhost = fmt.Sprintf("%s:%d", domain, publicPort)
		}

		// Canonicalize virtual host on default port 80
		if strings.HasSuffix(vhost, ":80") {
			vhost = vhost[0 : len(vhost)-3]
		}

		if strings.TrimSpace(t.regMsg.Hostname) != "" {
			t.url = fmt.Sprintf("http://%s", t.regMsg.Hostname)
		} else if strings.TrimSpace(t.regMsg.Subdomain) != "" {
			t.url = fmt.Sprintf("http://%s.%s", t.regMsg.Subdomain, vhost)
		}

		vhost = strings.ToLower(vhost)
		t.url = strings.ToLower(t.url)

		if t.url != "" {
			if err := tunnels.Register(t.url, t); err != nil {
				failReg(err)
				return
			}
		} else {
			t.url, err = tunnels.RegisterRepeat(func() string {
				return fmt.Sprintf("http://%x.%s", rand.Int31(), vhost)
			}, t)

			if err != nil {
				failReg(err)
				return
			}
		}
	}

	if m.Version != version.Proto {
		failReg(fmt.Errorf("Incompatible versions. Server %s, client %s. Download a new version at http://ngrok.com", version.MajorMinor(), m.Version))
		return
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

	metrics.OpenTunnel(t)
	return
}

func (t *Tunnel) shutdown() {
	t.Info("Shutting down")

	// mark that we're shutting down
	atomic.StoreInt32(&t.closing, 1)

	// if we have a public listener (this is a raw TCP tunnel, shut it down
	if t.listener != nil {
		t.listener.Close()
	}

	// remove ourselves from the tunnel registry
	tunnels.Del(t.url)

	// XXX: shut down all of the proxy connections?

	metrics.CloseTunnel(t)
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
			// not an error, we're shutting down this tunnel
			if atomic.LoadInt32(&t.closing) == 1 {
				return
			}

			t.Error("Failed to accept new TCP connection: %v", err)
			continue
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

	startTime := time.Now()
	metrics.OpenConnection(t, publicConn)

	t.Debug("Requesting new proxy connection")
	t.ctl.out <- &msg.ReqProxyMsg{}

	proxyConn := <-t.proxies
	t.Info("Returning proxy connection %s", proxyConn.Id())

	defer proxyConn.Close()
	bytesIn, bytesOut := conn.Join(publicConn, proxyConn)

	metrics.CloseConnection(t, publicConn, startTime, bytesIn, bytesOut)
}

func (t *Tunnel) RegisterProxy(conn conn.Conn) {
	t.Info("Registered proxy connection %s", conn.Id())
	conn.AddLogPrefix(t.Id())
	t.proxies <- conn
}
