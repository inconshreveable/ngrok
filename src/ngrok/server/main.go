package server

import (
	"math/rand"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"ngrok/util"
	"os"
)

const (
	registryCacheSize uint64 = 1024 * 1024 // 1 MB
)

// GLOBALS
var (
	tunnelRegistry    *TunnelRegistry
	controlRegistry   *ControlRegistry

	// XXX: kill these global variables - they're only used in tunnel.go for constructing forwarding URLs
	opts              *Options
	listeners         map[string] *conn.Listener
)

func NewProxy(pxyConn conn.Conn, regPxy *msg.RegProxyMsg) {
	// fail gracefully if the proxy connection fails to register
	defer func() {
		if r := recover(); r != nil {
			pxyConn.Warn("Failed with error: %v", r)
			pxyConn.Close()
		}
	}()

	// set logging prefix
	pxyConn.SetType("pxy")

	// look up the control connection for this proxy
	pxyConn.Info("Registering new proxy for %s", regPxy.ClientId)
	ctl := controlRegistry.Get(regPxy.ClientId)

	if ctl == nil {
		panic("No client found for identifier: " + regPxy.ClientId)
	}

	ctl.RegisterProxy(pxyConn)
}

// Listen for incoming control and proxy connections
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port, we can bust through
// restrictive firewalls.
func tunnelListener(addr string) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "tun", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control and proxy connections on %d", listener.Addr.String())
	for c := range listener.Conns {
		var rawMsg msg.Message
		if rawMsg, err = msg.ReadMsg(c); err != nil {
			c.Error("Failed to read message: %v", err)
			c.Close()
		}

		switch m := rawMsg.(type) {
		case *msg.RegMsg:
			go NewControl(c, m)

		case *msg.RegProxyMsg:
			go NewProxy(c, m)
		}
	}
}

func Main() {
	// parse options
	opts = parseArgs()

	// init logging
	log.LogTo(opts.logto)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)

	// init tunnel/control registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnelRegistry = NewTunnelRegistry(registryCacheSize, registryCacheFile)
	controlRegistry = NewControlRegistry()

	// start listeners
	listeners = make(map[string] *conn.Listener)

	// listen for http
	if opts.httpAddr != "" {
		listeners["http"] = startHttpListener(opts.httpAddr, nil)
	}

	// listen for https
	if opts.httpsAddr != "" {
		listeners["https"] = startHttpListener(opts.httpsAddr, tlsConfig)
	}

	// ngrok clients
	tunnelListener(opts.tunnelAddr)
}
