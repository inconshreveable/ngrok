package client

import (
	"fmt"
	metrics "github.com/inconshreveable/go-metrics"
	"io/ioutil"
	"math"
	"ngrok/client/mvc"
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
	pingInterval        = 20 * time.Second
	maxPongLatency      = 15 * time.Second
	updateCheckInterval = 6 * time.Hour
	BadGateway          = `<html>
<body style="background-color: #97a8b9">
    <div style="margin:auto; width:400px;padding: 20px 60px; background-color: #D3D3D3; border: 5px solid maroon;">
        <h2>Tunnel %s unavailable</h2>
        <p>Unable to initiate connection to <strong>%s</strong>. A web server must be running on port <strong>%s</strong> to complete the tunnel.</p>
`
)

type ClientModel struct {
	log.Logger

	id            string
	tunnels       map[string]mvc.Tunnel
	serverVersion string
	metrics       *ClientMetrics
	updateStatus  mvc.UpdateStatus
	connStatus    mvc.ConnStatus
	protoMap      map[string]proto.Protocol
	protocols     []proto.Protocol
	ctl           mvc.Controller
	serverAddr    string
	authToken     string
}

func newClientModel(ctl mvc.Controller) *ClientModel {
	protoMap := make(map[string]proto.Protocol)
	protoMap["http"] = proto.NewHttp()
	protoMap["https"] = protoMap["http"]
	protoMap["tcp"] = proto.NewTcp()
	protocols := []proto.Protocol{protoMap["http"], protoMap["tcp"]}

	return &ClientModel{
		Logger: log.NewPrefixLogger("client"),

		// unique client id
		id: util.RandIdOrPanic(8),

		// connection status
		connStatus: mvc.ConnConnecting,

		// update status
		updateStatus: mvc.UpdateNone,

		// metrics
		metrics: NewClientMetrics(),

		// protocols
		protoMap: protoMap,

		// protocol list
		protocols: protocols,

		// open tunnels
		tunnels: make(map[string]mvc.Tunnel),

		// controller
		ctl: ctl,
	}
}

// mvc.State interface
func (c ClientModel) GetProtocols() []proto.Protocol { return c.protocols }
func (c ClientModel) GetClientVersion() string       { return version.MajorMinor() }
func (c ClientModel) GetServerVersion() string       { return c.serverVersion }
func (c ClientModel) GetTunnels() []mvc.Tunnel {
	tunnels := make([]mvc.Tunnel, 0)
	for _, t := range c.tunnels {
		tunnels = append(tunnels, t)
	}
	return tunnels
}
func (c ClientModel) GetConnStatus() mvc.ConnStatus     { return c.connStatus }
func (c ClientModel) GetUpdateStatus() mvc.UpdateStatus { return c.updateStatus }

func (c ClientModel) GetConnectionMetrics() (metrics.Meter, metrics.Timer) {
	return c.metrics.connMeter, c.metrics.connTimer
}

func (c ClientModel) GetBytesInMetrics() (metrics.Counter, metrics.Histogram) {
	return c.metrics.bytesInCount, c.metrics.bytesIn
}

func (c ClientModel) GetBytesOutMetrics() (metrics.Counter, metrics.Histogram) {
	return c.metrics.bytesOutCount, c.metrics.bytesOut
}
func (c ClientModel) SetUpdateStatus(updateStatus mvc.UpdateStatus) { c.updateStatus = updateStatus; c.update() }


// mvc.Model interface
func (c *ClientModel) PlayRequest(tunnel mvc.Tunnel, payload []byte) {
	var localConn conn.Conn
	localConn, err := conn.Dial(tunnel.LocalAddr, "prv", nil)
	if err != nil {
		c.Warn("Failed to open private leg to %s: %v", tunnel.LocalAddr, err)
		return
	}

	defer localConn.Close()
	localConn = tunnel.Protocol.WrapConn(localConn, mvc.ConnectionContext{Tunnel: tunnel, ClientAddr: "127.0.0.1"})
	localConn.Write(payload)
	ioutil.ReadAll(localConn)
}

func (c *ClientModel) Shutdown() {
}

func (c *ClientModel) update() {
	c.ctl.Update(c)
}

func (c *ClientModel) Run(serverAddr, authToken string, ctl mvc.Controller, reg *msg.RegMsg, localaddr string) {
	c.serverAddr = serverAddr
	c.authToken = authToken
	c.ctl = ctl
	c.reconnectingControl(reg, localaddr)
}

func (c *ClientModel) reconnectingControl(reg *msg.RegMsg, localaddr string) {
	// how long we should wait before we reconnect
	maxWait := 30 * time.Second
	wait := 1 * time.Second

	for {
		c.control(reg, localaddr)

		if c.connStatus == mvc.ConnOnline {
			wait = 1 * time.Second
		}

		log.Info("Waiting %d seconds before reconnecting", int(wait.Seconds()))
		time.Sleep(wait)
		// exponentially increase wait time
		wait = 2 * wait
		wait = time.Duration(math.Min(float64(wait), float64(maxWait)))
		c.connStatus = mvc.ConnReconnecting
		c.update()
	}
}

// Establishes and manages a tunnel control connection with the server
func (c *ClientModel) control(reg *msg.RegMsg, localaddr string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("control recovering from failure %v", r)
		}
	}()

	// establish control channel
	conn, err := conn.Dial(c.serverAddr, "ctl", tlsConfig)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// register with the server
	reg.OS = runtime.GOOS
	reg.ClientId = c.id
	reg.Version = version.Proto
	reg.MmVersion = version.MajorMinor()
	reg.User = c.authToken

	if err = msg.WriteMsg(conn, reg); err != nil {
		panic(err)
	}

	// wait for the server to ack our register
	var regAck msg.RegAckMsg
	if err = msg.ReadMsgInto(conn, &regAck); err != nil {
		panic(err)
	}

	if regAck.Error != "" {
		emsg := fmt.Sprintf("Server failed to allocate tunnel: %s", regAck.Error)
		c.ctl.Shutdown(emsg)
		return
	}

	tunnel := mvc.Tunnel{
		PublicUrl: regAck.Url,
		LocalAddr: localaddr,
		Protocol:  c.protoMap[reg.Protocol],
	}

	c.tunnels[tunnel.PublicUrl] = tunnel

	// update UI state
	c.id = regAck.ClientId
	c.Info("Tunnel established at %v", tunnel.PublicUrl)
	c.connStatus = mvc.ConnOnline
	c.serverVersion = regAck.MmVersion
	c.update()

	SaveAuthToken(c.authToken)

	// start the heartbeat
	lastPong := time.Now().UnixNano()
	c.ctl.Go(func() { c.heartbeat(&lastPong, conn) })

	// main control loop
	for {
		var rawMsg msg.Message
		if rawMsg, err = msg.ReadMsg(conn); err != nil {
			panic(err)
		}

		switch m := rawMsg.(type) {
		case *msg.ReqProxyMsg:
			c.ctl.Go(c.proxy)

		case *msg.PongMsg:
			atomic.StoreInt64(&lastPong, time.Now().UnixNano())

		case *msg.RegAckMsg:
			if m.Error != "" {
				c.Error("Server failed to allocate tunnel: %s", regAck.Error)
				continue
			}

			tunnel := mvc.Tunnel{
				PublicUrl: m.Url,
				LocalAddr: localaddr,
				Protocol:  c.protoMap[m.Protocol],
			}

			c.tunnels[tunnel.PublicUrl] = tunnel
			c.Info("Tunnel established at %v", tunnel.PublicUrl)
			c.update()

		default:
			conn.Warn("Ignoring unknown control message %v ", m)
		}
	}
}

// Establishes and manages a tunnel proxy connection with the server
func (c *ClientModel) proxy() {
	remoteConn, err := conn.Dial(c.serverAddr, "pxy", tlsConfig)
	if err != nil {
		log.Error("Failed to establish proxy connection: %v", err)
		return
	}

	defer remoteConn.Close()
	err = msg.WriteMsg(remoteConn, &msg.RegProxyMsg{ClientId: c.id})
	if err != nil {
		log.Error("Failed to write RegProxyMsg: %v", err)
		return
	}

	// wait for the server to ack our register
	var startPxyMsg msg.StartProxyMsg
	if err = msg.ReadMsgInto(remoteConn, &startPxyMsg); err != nil {
		log.Error("Server failed to write StartProxyMsg: %v", err)
		return
	}

	tunnel, ok := c.tunnels[startPxyMsg.Url]
	if !ok {
		c.Error("Couldn't find tunnel for proxy: %s", startPxyMsg.Url)
		return
	}

	// start up the private connection
	start := time.Now()
	localConn, err := conn.Dial(tunnel.LocalAddr, "prv", nil)
	if err != nil {
		remoteConn.Warn("Failed to open private leg %s: %v", tunnel.LocalAddr, err)

		if tunnel.Protocol.GetName() == "http" {
			// try to be helpful when you're in HTTP mode and a human might see the output
			badGatewayBody := fmt.Sprintf(BadGateway, tunnel.PublicUrl, tunnel.LocalAddr, tunnel.LocalAddr)
			remoteConn.Write([]byte(fmt.Sprintf(`HTTP/1.0 502 Bad Gateway
Content-Type: text/html
Content-Length: %d

%s`, len(badGatewayBody), badGatewayBody)))
		}
		return
	}
	defer localConn.Close()

	m := c.metrics
	m.proxySetupTimer.Update(time.Since(start))
	m.connMeter.Mark(1)
	c.update()
	m.connTimer.Time(func() {
		localConn := tunnel.Protocol.WrapConn(localConn, mvc.ConnectionContext{Tunnel: tunnel, ClientAddr: startPxyMsg.ClientAddr})
		bytesIn, bytesOut := conn.Join(localConn, remoteConn)
		m.bytesIn.Update(bytesIn)
		m.bytesOut.Update(bytesOut)
		m.bytesInCount.Inc(bytesIn)
		m.bytesOutCount.Inc(bytesOut)
	})
	c.update()
}

// Hearbeating to ensure our connection ngrokd is still live
func (c *ClientModel) heartbeat(lastPongAddr *int64, conn conn.Conn) {
	lastPing := time.Unix(atomic.LoadInt64(lastPongAddr)-1, 0)
	ping := time.NewTicker(pingInterval)
	pongCheck := time.NewTicker(time.Second)

	defer func() {
		conn.Close()
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
			err := msg.WriteMsg(conn, &msg.PingMsg{})
			if err != nil {
				conn.Debug("Got error %v when writing PingMsg", err)
				return
			}
			lastPing = time.Now()
		}
	}
}
