package server

import (
	"fmt"
	"io"
	"ngrok/conn"
	"ngrok/msg"
	"ngrok/version"
	"runtime/debug"
	"time"
)

const (
	pingTimeoutInterval = 30 * time.Second
	connReapInterval    = 10 * time.Second
)

type Control struct {
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
}

func NewControl(conn conn.Conn, regMsg *msg.RegMsg) {
	// create the object
	// channels are buffered because we read and write to them
	// from the same goroutine in managerThread()
	c := &Control{
		conn:     conn,
		out:      make(chan msg.Message, 5),
		in:       make(chan msg.Message, 5),
		stop:     make(chan msg.Message, 5),
		lastPing: time.Now(),
	}

	// register the first tunnel
	c.in <- regMsg

	// manage the connection
	go c.managerThread()
	go c.readThread()
}

// Register a new tunnel on this control connection
func (c *Control) registerTunnel(regMsg *msg.RegMsg) {
	c.conn.Debug("Registering new tunnel")
	t, err := NewTunnel(regMsg, c)
	if err != nil {
		ack := &msg.RegAckMsg{Error: err.Error()}
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
	c.out <- &msg.RegAckMsg{
		Url:       t.url,
		ProxyAddr: fmt.Sprintf("%s:%d", opts.domain, opts.tunnelPort),
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
	}

	if regMsg.Protocol == "http" {
		httpsRegMsg := *regMsg
		httpsRegMsg.Protocol = "https"
		c.in <- &httpsRegMsg
	}
}

func (c *Control) managerThread() {
	reap := time.NewTicker(connReapInterval)

	// all shutdown functionality in here
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::managerThread failed with error %v: %s", err, debug.Stack())
		}

		reap.Stop()
		c.conn.Close()

		// shutdown all of the tunnels
		for _, t := range c.tunnels {
			t.shutdown()
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
			case *msg.RegMsg:
				c.registerTunnel(m)

			case *msg.PingMsg:
				c.lastPing = time.Now()
				c.out <- &msg.PongMsg{}

			case *msg.VersionMsg:
				c.out <- &msg.VersionRespMsg{
					Version:   version.Proto,
					MmVersion: version.MajorMinor(),
				}
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
