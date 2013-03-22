package server

import (
	"io"
	"net"
	"ngrok/conn"
	"ngrok/proto"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const (
	pingInterval     = 30 * time.Second
	connReapInterval = pingInterval * 5
)

type Control struct {
	// actual connection
	conn conn.Conn

	// channels for communicating messages over the connection
	out  chan (interface{})
	in   chan (proto.Message)
	stop chan (proto.Message)

	// heartbeat
	lastPong int64

	// tunnel
	tun *Tunnel
}

func NewControl(tcpConn *net.TCPConn) {
	c := &Control{
		conn:     conn.NewTCP(tcpConn, "ctl"),
		out:      make(chan (interface{}), 1),
		in:       make(chan (proto.Message), 1),
		stop:     make(chan (proto.Message), 1),
		lastPong: time.Now().Unix(),
	}

	go c.managerThread()
	go c.readThread()
}

func (c *Control) managerThread() {
	ping := time.NewTicker(pingInterval)
	reap := time.NewTicker(connReapInterval)

	// all shutdown functionality in here
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::managerThread failed with error %v: %s", err, debug.Stack())
		}
		ping.Stop()
		reap.Stop()
		c.conn.Close()

		// shutdown the tunnel if it's open
		if c.tun != nil {
			c.tun.shutdown()
		}
	}()

	for {
		select {
		case m := <-c.out:
			proto.WriteMsg(c.conn, m)

		case <-ping.C:
			proto.WriteMsg(c.conn, &proto.PingMsg{})

		case <-reap.C:
			if (time.Now().Unix() - c.lastPong) > 60 {
				c.conn.Info("Lost heartbeat")
				metrics.lostHeartbeatMeter.Mark(1)
				return
			}

		case m := <-c.stop:
			if m != nil {
				proto.WriteMsg(c.conn, m)
			}
			return

		case msg := <-c.in:
			switch msg.GetType() {
			case "RegMsg":
				c.conn.Info("Registering new tunnel")
				c.tun = newTunnel(msg.(*proto.RegMsg), c)

			case "PongMsg":
				atomic.StoreInt64(&c.lastPong, time.Now().Unix())

			case "VersionReqMsg":
				c.out <- &proto.VersionRespMsg{Version: version}
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
		if msg, err := proto.ReadMsg(c.conn); err != nil {
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
