package server

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	"net"
	"ngrok/conn"
	nlog "ngrok/log"
	"ngrok/msg"
	"regexp"
)

type Options struct {
	publicPort int
	proxyPort  int
	tunnelPort int
	domain     string
}

/* GLOBALS */
var (
	hostRegex *regexp.Regexp
	version   string = "0.1"
	proxyAddr string
	tunnels   *TunnelManager
)

func init() {
	hostRegex = regexp.MustCompile("[H|h]ost: ([^\\(\\);:,<>]+)\n")
}

func parseArgs() *Options {
	publicPort := flag.Int("publicport", 80, "Public port")
	tunnelPort := flag.Int("tunnelport", 2280, "Tunnel port")
	proxyPort := flag.Int("proxyPort", 0, "Proxy port")
	domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")

	flag.Parse()

	return &Options{
		publicPort: *publicPort,
		tunnelPort: *tunnelPort,
		proxyPort:  *proxyPort,
		domain:     *domain,
	}
}

/**
 * Listens for new control connections from tunnel clients
 */
func controlListener(addr *net.TCPAddr, domain string) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "ctl")
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
	listener, err := conn.Listen(addr, "pxy")
	if err != nil {
		panic(err)
	}

	// set global proxy addr variable
	proxyAddr = fmt.Sprintf("%s:%d", domain, listener.Port)
	log.Info("Listening for proxy connection on %d", listener.Port)
	for conn := range listener.Conns {
		go func() {
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
		}()
	}
}

func Main() {
	nlog.LogToConsole()
	done := make(chan int)
	// parse options
	opts := parseArgs()

	tunnels = NewTunnelManager(opts.domain)

	go proxyListener(&net.TCPAddr{net.ParseIP("0.0.0.0"), opts.proxyPort}, opts.domain)
	go controlListener(&net.TCPAddr{net.ParseIP("0.0.0.0"), opts.tunnelPort}, opts.domain)
	go httpListener(&net.TCPAddr{net.ParseIP("0.0.0.0"), opts.publicPort})

	<-done
}
