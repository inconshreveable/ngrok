package server

import (
	"fmt"
	"net"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"os"
)

// GLOBALS
var (
	opts              *Options
	tunnels           *TunnelRegistry
	registryCacheSize uint64 = 1024 * 1024 // 1 MB
	domain            string
	publicPort        int
)

func NewProxy(pxyConn conn.Conn, regPxy *msg.RegProxyMsg) {
	// fail gracefully if the proxy connection fails to register
	defer func() {
		if r := recover(); r != nil {
			pxyConn.Warn("Failed with error: %v", r)
			pxyConn.Close()
		}
	}()

	// add log prefix
	pxyConn.AddLogPrefix("pxy")

	// look up the tunnel for this proxy
	pxyConn.Info("Registering new proxy for %s", regPxy.Url)
	tunnel := tunnels.Get(regPxy.Url)
	if tunnel == nil {
		panic("No tunnel found for: " + regPxy.Url)
	}

	if regPxy.ClientId != tunnel.regMsg.ClientId {
		panic(fmt.Sprintf("Client identifier %s does not match tunnel's %s", regPxy.ClientId, tunnel.regMsg.ClientId))
	}

	// register the proxy connection with the tunnel
	tunnel.RegisterProxy(pxyConn)
}

// Listen for incoming control and proxy connections
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port, we can bust through
// restrictive firewalls.
func tunnelListener(addr *net.TCPAddr, domain string) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "ctl", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control and proxy connections on %d", listener.Port)
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

	// init tunnel registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnels = NewTunnelRegistry(registryCacheSize, registryCacheFile)

	// ngrok clients
	go tunnelListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.tunnelPort}, opts.domain)

	// listen for http
	if opts.httpPort != -1 {
		go httpListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.httpPort}, nil)
	}

	// listen for https
	if opts.httpsPort != -1 {
		go httpListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.httpsPort}, tlsConfig)
	}

	// wait forever
	done := make(chan int)
	<-done
}
