package server

import (
	"flag"
	"fmt"
	"net"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"os"
)

type Options struct {
	publicPort int
	proxyPort  int
	tunnelPort int
	domain     string
	logto      string
}

/* GLOBALS */
var (
	proxyAddr         string
	tunnels           *TunnelRegistry
	registryCacheSize uint64 = 1024 * 1024 // 1 MB
	domain            string
)

func parseArgs() *Options {
	publicPort := flag.Int("publicport", 80, "Public port")
	tunnelPort := flag.Int("tunnelport", 4443, "Tunnel port")
	proxyPort := flag.Int("proxyPort", 0, "Proxy port")
	domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")
	logto := flag.String(
		"log",
		"stdout",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	flag.Parse()

	return &Options{
		publicPort: *publicPort,
		tunnelPort: *tunnelPort,
		proxyPort:  *proxyPort,
		domain:     *domain,
		logto:      *logto,
	}
}

/**
 * Listens for new control connections from tunnel clients
 */
func controlListener(addr *net.TCPAddr, domain string) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "ctl", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control connections on %d", listener.Port)
	for c := range listener.Conns {
		NewControl(c)
	}
}

/**
 * Listens for new proxy connections from tunnel clients
 */
func proxyListener(addr *net.TCPAddr, domain string) {
	listener, err := conn.Listen(addr, "pxy", tlsConfig)
	if err != nil {
		panic(err)
	}

	// set global proxy addr variable
	proxyAddr = fmt.Sprintf("%s:%d", domain, listener.Port)
	log.Info("Listening for proxy connection on %d", listener.Port)
	for proxyConn := range listener.Conns {
		go func(conn conn.Conn) {
			// fail gracefully if the proxy connection dies
			defer func() {
				if r := recover(); r != nil {
					conn.Warn("Failed with error: %v", r)
					conn.Close()
				}
			}()

			// read the proxy register message
			var regPxy msg.RegProxyMsg
			if err = msg.ReadMsgInto(conn, &regPxy); err != nil {
				panic(err)
			}

			// look up the tunnel for this proxy
			conn.Info("Registering new proxy for %s", regPxy.Url)
			tunnel := tunnels.Get(regPxy.Url)
			if tunnel == nil {
				panic("No tunnel found for: " + regPxy.Url)
			}

			// register the proxy connection with the tunnel
			tunnel.RegisterProxy(conn)
		}(proxyConn)
	}
}

func Main() {
	// parse options
	opts := parseArgs()
	domain = opts.domain

	// init logging
	log.LogTo(opts.logto)

	// init tunnel registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnels = NewTunnelRegistry(registryCacheSize, registryCacheFile)

	go proxyListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.proxyPort}, opts.domain)
	go controlListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.tunnelPort}, opts.domain)
	go httpListener(&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: opts.publicPort})

	// wait forever
	done := make(chan int)
	<-done
}
