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

func getTCPPort(addr net.Addr) int {
	return addr.(*net.TCPAddr).Port
}

/**
 * Listens for new control connections from tunnel clients
 */
func controlListener(addr *net.TCPAddr, domain string) {
	// listen for incoming connections
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control connections on %d", getTCPPort(addr))
	for {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		NewControl(tcpConn)
	}
}

/**
 * Listens for new proxy connections from tunnel clients
 */
func proxyListener(addr *net.TCPAddr, domain string) {
	listener, err := net.ListenTCP("tcp", addr)
	proxyAddr = fmt.Sprintf("%s:%d", domain, getTCPPort(listener.Addr()))

	if err != nil {
		panic(err)
	}

	log.Info("Listening for proxy connection on %d", getTCPPort(listener.Addr()))
	for {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		conn := conn.NewTCP(tcpConn, "pxy")

		go func() {
			defer func() {
				if r := recover(); r != nil {
					conn.Warn("Failed with error: %v", r)
					conn.Close()
				}
			}()

			var regPxy msg.RegProxyMsg
			if err = msg.ReadMsgInto(conn, &regPxy); err != nil {
				panic(err)
			}

			conn.Info("Registering new proxy for %s", regPxy.Url)

			tunnel := tunnels.Get(regPxy.Url)
			if tunnel == nil {
				panic("No tunnel found for: " + regPxy.Url)
			}

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
