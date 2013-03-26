package client

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	PORT_OUT_OF_RANGE error = errors.New("Port number must be between 1 and 65535")
)

type Options struct {
	server      string
	auth        string
	hostname    string
	localaddr   string
	protocol    string
	url         string
	subdomain   string
	historySize int
}

func fail(msg string, args ...interface{}) {
	//log.Error(msg, args..)
	fmt.Printf(msg+"\n", args...)
	flag.PrintDefaults()
	os.Exit(1)
}

func parsePort(portString string) (err error) {
	var port int
	if port, err = strconv.Atoi(portString); err != nil {
		return err
	}

	if port < 1 || port > 65535 {
		return PORT_OUT_OF_RANGE
	}

	return
}

// Local address could be a port of a host:port string
// we always return a host:port string from this function or fail
func parseLocalAddr() string {
	if flag.NArg() == 0 {
		fail("LOCAL not specified, specify a port number or host:port connection string")
	}

	if flag.NArg() > 1 {
		fail("Only one LOCAL may be specified, not %d", flag.NArg())
	}

	addr := flag.Arg(0)

	// try to parse as a port number
	if err := parsePort(addr); err == nil {
		return fmt.Sprintf("127.0.0.1:%s", addr)
	} else if err == PORT_OUT_OF_RANGE {
		fail("%s is not in the valid port range 1-65535")
	}

	// try to parse as a connection string
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		fail("%s is not a port number of a host:port connection string", addr)
	}

	if parsePort(parts[1]) != nil {
		fail("The port of the connection string '%s' is not a valid port number (1-65535)",
			parts[1])
	}

	return addr
}

func parseProtocol(proto string) string {
	switch proto {
	case "http":
		fallthrough
	case "tcp":
		return proto
	default:
		fail("%s is not a valid protocol", proto)
	}
	panic("unreachable")
}

func parseArgs() *Options {
	server := flag.String(
		"server",
		"ngrok.com:2280",
		"The remote ngrok server")

	auth := flag.String(
		"auth",
		"",
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	hostname := flag.String(
		"hostname",
		"",
		"A full DNS hostname to identify public tunnel endpoint. (Advanced, requires you CNAME your DNS)")

	subdomain := flag.String(
		"subdomain",
		"",
		"Request a custom subdomain from the ngrok server. (HTTP mode only)")

	protocol := flag.String(
		"proto",
		"http",
		"The protocol of the traffic over the tunnel {'http', 'tcp'} (default: 'http')")

	historySize := flag.Int(
		"history",
		20,
		"The number of previous requests to keep in your history")

	flag.Parse()

	return &Options{
		server:      *server,
		auth:        *auth,
		hostname:    *hostname,
		subdomain:   *subdomain,
		localaddr:   parseLocalAddr(),
		protocol:    parseProtocol(*protocol),
		historySize: *historySize,
	}
}
