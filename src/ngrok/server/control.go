package server

import (
	"fmt"
	"io"
	"ngrok/conn"
	"ngrok/msg"
	"ngrok/util"
	"ngrok/version"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

const (
	pingTimeoutInterval = 30 * time.Second
	connReapInterval    = 10 * time.Second
)

type Control struct {
	// auth message
	auth *msg.Auth

	// actual connection
	conn conn.Conn

	// put a message in this channel to send it over
	// conn to the client
	out chan (msg.Message)

	// read from this channel to get the next message sent
	// to us over conn by the client
	in chan (msg.Message)

	// put a message in this channel to send it over
	// conn to the client and then terminate this
	// control connection and all of its tunnels
	stop chan (msg.Message)

	// the last time we received a ping from the client - for heartbeats
	lastPing time.Time

	// all of the tunnels this control connection handles
	tunnels []*Tunnel

	// proxy connections
	proxies chan conn.Conn

	// closing indicator
	closing int32

	// identifier
	id string
}

func NewControl(ctlConn conn.Conn, authMsg *msg.Auth) {
	var err error

	// create the object
	// channels are buffered because we read and write to them
	// from the same goroutine in managerThread()
	c := &Control{
		auth:     authMsg,
		conn:     ctlConn,
		out:      make(chan msg.Message, 5),
		in:       make(chan msg.Message, 5),
		stop:     make(chan msg.Message, 5),
		proxies:  make(chan conn.Conn, 10),
		lastPing: time.Now(),
	}

	failAuth := func(e error) {
		_ = msg.WriteMsg(ctlConn, &msg.AuthResp{Error: e.Error()})
		ctlConn.Close()
	}

	// register the clientid
	c.id = authMsg.ClientId
	if c.id == "" {
		// it's a new session, assign an ID
		if c.id, err = util.SecureRandId(16); err != nil {
			failAuth(err)
			return
		}
	}

	if authMsg.Version != version.Proto {
		failAuth(fmt.Errorf("Incompatible versions. Server %s, client %s. Download a new version at http://ngrok.com", version.MajorMinor(), authMsg.Version))
		return
	}

	// register the control
	controlRegistry.Add(c.id, c)

	c.out <- &msg.AuthResp{
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
		ClientId:  c.id,
	}

	// As a performance optimization, ask for a proxy connection up front
	c.out <- &msg.ReqProxy{}

	// set logging prefix
	ctlConn.SetType("ctl")

	// manage the connection
	go c.managerThread()
	go c.readThread()
}

// Register a new tunnel on this control connection
func (c *Control) registerTunnel(rawTunnelReq *msg.ReqTunnel) {
	for _, proto := range strings.Split(rawTunnelReq.Protocol, "+") {
		tunnelReq := *rawTunnelReq
		tunnelReq.Protocol = proto

		c.conn.Debug("Registering new tunnel")
		t, err := NewTunnel(&tunnelReq, c)
		if err != nil {
			ack := &msg.NewTunnel{Error: err.Error()}
			if len(c.tunnels) == 0 {
				// you can't fail your first tunnel registration
				// terminate the control connection
				c.stop <- ack
			} else {
				// inform client of failure
				c.out <- ack
			}

			// we're done
			return
		}

		// add it to the list of tunnels
		c.tunnels = append(c.tunnels, t)

		// acknowledge success
		c.out <- &msg.NewTunnel{
			Url:      t.url,
			Protocol: proto,
		}

		rawTunnelReq.Hostname = strings.Replace(t.url, proto+"://", "", 1)
	}
}

func (c *Control) managerThread() {
	reap := time.NewTicker(connReapInterval)

	// all shutdown functionality in here
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::managerThread failed with error %v: %s", err, debug.Stack())
		}

		// remove from the control registry
		controlRegistry.Del(c.id)

		// mark that we're shutting down
		atomic.StoreInt32(&c.closing, 1)

		// stop the reaping timer
		reap.Stop()

		// close the connection
		c.conn.Close()

		// shutdown all of the tunnels
		for _, t := range c.tunnels {
			t.Shutdown()
		}

		// we're safe to close(c.proxies) because c.closing
		// protects us inside of RegisterProxy
		close(c.proxies)

		// shut down all of the proxy connections
		for p := range c.proxies {
			p.Close()
		}

	}()

	for {
		select {
		case m := <-c.out:
			msg.WriteMsg(c.conn, m)

		case m := <-c.stop:
			if m != nil {
				msg.WriteMsg(c.conn, m)
			}
			return

		case <-reap.C:
			if time.Since(c.lastPing) > pingTimeoutInterval {
				c.conn.Info("Lost heartbeat")
				return
			}

		case mRaw := <-c.in:
			switch m := mRaw.(type) {
			case *msg.ReqTunnel:
				c.registerTunnel(m)

			case *msg.Ping:
				c.lastPing = time.Now()
				c.out <- &msg.Pong{}
			}
		}
	}
}

func (c *Control) readThread() {
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::readThread failed with error %v: %s", err, debug.Stack())
		}
		c.stop <- nil
	}()

	// read messages from the control channel
	for {
		if msg, err := msg.ReadMsg(c.conn); err != nil {
			if err == io.EOF {
				c.conn.Info("EOF")
				return
			} else {
				panic(err)
			}
		} else {
			c.in <- msg
		}
	}
}

func (c *Control) RegisterProxy(conn conn.Conn) {
	if atomic.LoadInt32(&c.closing) == 1 {
		c.conn.Debug("Can't register proxies for a control that is closing")
		conn.Close()
		return
	}

	select {
	case c.proxies <- conn:
		c.conn.Info("Registered proxy connection %s", conn.Id())
	default:
		// c.proxies buffer is full, discard this one
		conn.Close()
	}
}

// Remove a proxy connection from the pool and return it
// If not proxy connections are in the pool, request one
// and wait until it is available
// Returns an error if we couldn't get a proxy because it took too long
// or the tunnel is closing
func (c *Control) GetProxy() (proxyConn conn.Conn, err error) {
	// initial timeout is zero to try to get a proxy connection without asking for one
	timeout := time.NewTimer(0)

	// get a proxy connection. if we timeout, request one over the control channel
	for proxyConn == nil {
		var ok bool
		select {
		case proxyConn, ok = <-c.proxies:
			if !ok {
				err = fmt.Errorf("No proxy connections available, control is closing")
				return
			}
			continue
		case <-timeout.C:
			c.conn.Debug("Requesting new proxy connection")
			// request a proxy connection
			c.out <- &msg.ReqProxy{}
			// timeout after 1 second if we don't get one
			timeout.Reset(1 * time.Second)
		}
	}

	// To try to reduce latency hanndling tunnel connections, we employ
	// the following curde heuristic:
	// If the proxy connection pool is empty, request a new one.
	// The idea is to always have at least one proxy connection available for immediate use.
	// There are two major issues with this strategy: it's not thread safe and it's not predictive.
	// It should be a good start though.
	if len(c.proxies) == 0 {
		c.out <- &msg.ReqProxy{}
	}

	return
}
