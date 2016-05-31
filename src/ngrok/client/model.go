package client

import (
	"crypto/tls"
	"fmt"
	metrics "github.com/rcrowley/go-metrics"
	"io/ioutil"
	"math"
	"net"
	"ngrok/client/mvc"
	"ngrok/conn"
	"ngrok/log"
	"ngrok/msg"
	"ngrok/proto"
	"ngrok/util"
	"ngrok/version"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultServerAddr   = "ngrokd.ngrok.com:443"
	defaultInspectAddr  = "127.0.0.1:4040"
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
	proxyUrl      string
	authToken     string
	tlsConfig     *tls.Config
	tunnelConfig  map[string]*TunnelConfiguration
	configPath    string
}

func newClientModel(config *Configuration, ctl mvc.Controller) *ClientModel {
	protoMap := make(map[string]proto.Protocol)
	protoMap["http"] = proto.NewHttp()
	protoMap["https"] = protoMap["http"]
	protoMap["tcp"] = proto.NewTcp()
	protocols := []proto.Protocol{protoMap["http"], protoMap["tcp"]}

	m := &ClientModel{
		Logger: log.NewPrefixLogger("client"),

		// server address
		serverAddr: config.ServerAddr,

		// proxy address
		proxyUrl: config.HttpProxy,

		// auth token
		authToken: config.AuthToken,

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

		// tunnel configuration
		tunnelConfig: config.Tunnels,

		// config path
		configPath: config.Path,
	}

	// configure TLS
	if config.TrustHostRootCerts {
		m.Info("Trusting host's root certificates")
		m.tlsConfig = &tls.Config{}
	} else {
		m.Info("Trusting root CAs: %v", rootCrtPaths)
		var err error
		if m.tlsConfig, err = LoadTLSConfig(rootCrtPaths); err != nil {
			panic(err)
		}
	}

	// configure TLS SNI
	m.tlsConfig.ServerName = serverName(m.serverAddr)
	m.tlsConfig.InsecureSkipVerify = useInsecureSkipVerify()

	return m
}

// server name in release builds is the host part of the server address
func serverName(addr string) string {
	host, _, err := net.SplitHostPort(addr)

	// should never panic because the config parser calls SplitHostPort first
	if err != nil {
		panic(err)
	}

	return host
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

func (c *ClientModel) Run() {
	// how long we should wait before we reconnect
	maxWait := 30 * time.Second
	wait := 1 * time.Second

	for {
		// run the control channel
		c.control()

		// control only returns when a failure has occurred, so we're going to try to reconnect
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
func (c *ClientModel) control() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("control recovering from failure %v", r)
		}
	}()

	// establish control channel
	var (
		ctlConn conn.Conn
		err     error
	)
	if c.proxyUrl == "" {
		// simple non-proxied case, just connect to the server
		ctlConn, err = conn.Dial(c.serverAddr, "ctl", c.tlsConfig)
	} else {
		ctlConn, err = conn.DialHttpProxy(c.proxyUrl, c.serverAddr, "ctl", c.tlsConfig)
	}
	if err != nil {
		panic(err)
	}
	defer ctlConn.Close()

	// authenticate with the server
	auth := &msg.Auth{
		ClientId:  c.id,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
		User:      c.authToken,
	}

	if err = msg.WriteMsg(ctlConn, auth); err != nil {
		panic(err)
	}

	// wait for the server to authenticate us
	var authResp msg.AuthResp
	if err = msg.ReadMsgInto(ctlConn, &authResp); err != nil {
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
	if err = SaveAuthToken(c.configPath, c.authToken); err != nil {
		c.Error("Failed to save auth token: %v", err)
	}

	// request tunnels
	reqIdToTunnelConfig := make(map[string]*TunnelConfiguration)
	for _, config := range c.tunnelConfig {
		// create the protocol list to ask for
		var protocols []string
		for proto, _ := range config.Protocols {
			protocols = append(protocols, proto)
		}

		reqTunnel := &msg.ReqTunnel{
			ReqId:      util.RandId(8),
			Protocol:   strings.Join(protocols, "+"),
			Hostname:   config.Hostname,
			Subdomain:  config.Subdomain,
			HttpAuth:   config.HttpAuth,
			RemotePort: config.RemotePort,
		}

		// send the tunnel request
		if err = msg.WriteMsg(ctlConn, reqTunnel); err != nil {
			panic(err)
		}

		// save request id association so we know which local address
		// to proxy to later
		reqIdToTunnelConfig[reqTunnel.ReqId] = config
	}

	// start the heartbeat
	lastPong := time.Now().UnixNano()
	c.ctl.Go(func() { c.heartbeat(&lastPong, ctlConn) })

	// main control loop
	for {
		var rawMsg msg.Message
		if rawMsg, err = msg.ReadMsg(ctlConn); err != nil {
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
				LocalAddr: reqIdToTunnelConfig[m.ReqId].Protocols[m.Protocol],
				Protocol:  c.protoMap[m.Protocol],
			}

			c.tunnels[tunnel.PublicUrl] = tunnel
			c.connStatus = mvc.ConnOnline
			c.Info("Tunnel established at %v", tunnel.PublicUrl)
			c.update()

		default:
			ctlConn.Warn("Ignoring unknown control message %v ", m)
		}
	}
}

// Establishes and manages a tunnel proxy connection with the server
func (c *ClientModel) proxy() {
	var (
		remoteConn conn.Conn
		err        error
	)

	if c.proxyUrl == "" {
		remoteConn, err = conn.Dial(c.serverAddr, "pxy", c.tlsConfig)
	} else {
		remoteConn, err = conn.DialHttpProxy(c.proxyUrl, c.serverAddr, "pxy", c.tlsConfig)
	}

	if err != nil {
		log.Error("Failed to establish proxy connection: %v", err)
		return
	}
	defer remoteConn.Close()

	err = msg.WriteMsg(remoteConn, &msg.RegProxy{ClientId: c.id})
	if err != nil {
		remoteConn.Error("Failed to write RegProxy: %v", err)
		return
	}

	// wait for the server to ack our register
	var startPxy msg.StartProxy
	if err = msg.ReadMsgInto(remoteConn, &startPxy); err != nil {
		remoteConn.Error("Server failed to write StartProxy: %v", err)
		return
	}

	tunnel, ok := c.tunnels[startPxy.Url]
	if !ok {
		remoteConn.Error("Couldn't find tunnel for proxy: %s", startPxy.Url)
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
