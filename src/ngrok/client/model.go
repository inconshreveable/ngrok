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
func (c ClientModel) SetUpdateStatus(updateStatus mvc.UpdateStatus) {
	c.updateStatus = updateStatus
	c.update()
}

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

func (c *ClientModel) Run(serverAddr, authToken string, ctl mvc.Controller, reqTunnel *msg.ReqTunnel, localaddr string) {
	c.serverAddr = serverAddr
	c.authToken = authToken
	c.ctl = ctl
	c.reconnectingControl(reqTunnel, localaddr)
}

func (c *ClientModel) reconnectingControl(reqTunnel *msg.ReqTunnel, localaddr string) {
	// how long we should wait before we reconnect
	maxWait := 30 * time.Second
	wait := 1 * time.Second

	for {
		c.control(reqTunnel, localaddr)

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
func (c *ClientModel) control(reqTunnel *msg.ReqTunnel, localaddr string) {
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

	// authenticate with the server
	auth := &msg.Auth{
		ClientId:  c.id,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
		User:      c.authToken,
	}

	if err = msg.WriteMsg(conn, auth); err != nil {
		panic(err)
	}

	// wait for the server to authenticate us
	var authResp msg.AuthResp
	if err = msg.ReadMsgInto(conn, &authResp); err != nil {
		panic(err)
	}

	if authResp.Error != "" {
		emsg := fmt.Sprintf("Failed to authenticate to server: %s", authResp.Error)
		c.ctl.Shutdown(emsg)
		return
	}

	c.id = authResp.ClientId
	c.serverVersion = authResp.MmVersion
	c.Info("Authenticated with server, client id: %v", c.id)
	c.update()
	SaveAuthToken(c.authToken)

	// register the tunnel
	if err = msg.WriteMsg(conn, reqTunnel); err != nil {
		panic(err)
	}

	// register an https tunnel as well for http tunnels
	if reqTunnel.Protocol == "http" {
		httpsReqTunnel := *reqTunnel
		httpsReqTunnel.Protocol = "https"
		// httpsReqTunnel.ReqId =

		if err = msg.WriteMsg(conn, &httpsReqTunnel); err != nil {
			panic(err)
		}
	}

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
		case *msg.ReqProxy:
			c.ctl.Go(c.proxy)

		case *msg.Pong:
			atomic.StoreInt64(&lastPong, time.Now().UnixNano())

		case *msg.NewTunnel:
			if m.Error != "" {
				emsg := fmt.Sprintf("Server failed to allocate tunnel: %s", m.Error)
				c.Error(emsg)
				c.ctl.Shutdown(emsg)
				continue
			}

			tunnel := mvc.Tunnel{
				PublicUrl: m.Url,
				LocalAddr: localaddr,
				Protocol:  c.protoMap[m.Protocol],
			}

			c.tunnels[tunnel.PublicUrl] = tunnel
			c.connStatus = mvc.ConnOnline
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
	err = msg.WriteMsg(remoteConn, &msg.RegProxy{ClientId: c.id})
	if err != nil {
		log.Error("Failed to write RegProxy: %v", err)
		return
	}

	// wait for the server to ack our register
	var startPxy msg.StartProxy
	if err = msg.ReadMsgInto(remoteConn, &startPxy); err != nil {
		log.Error("Server failed to write StartProxy: %v", err)
		return
	}

	tunnel, ok := c.tunnels[startPxy.Url]
	if !ok {
		c.Error("Couldn't find tunnel for proxy: %s", startPxy.Url)
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
		localConn := tunnel.Protocol.WrapConn(localConn, mvc.ConnectionContext{Tunnel: tunnel, ClientAddr: startPxy.ClientAddr})
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
			err := msg.WriteMsg(conn, &msg.Ping{})
			if err != nil {
				conn.Debug("Got error %v when writing PingMsg", err)
				return
			}
			lastPing = time.Now()
		}
	}
}
