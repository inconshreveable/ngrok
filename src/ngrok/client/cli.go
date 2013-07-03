package client

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"ngrok/version"
	"os"
	"strconv"
)

var (
	PORT_OUT_OF_RANGE error = errors.New("Port number must be between 1 and 65535")
)

type Options struct {
	server    string
	httpAuth  string
	hostname  string
	localaddr string
	protocol  string
	url       string
	subdomain string
	webport   int
	logto     string
	authtoken string
}

func fail(msg string, args ...interface{}) {
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
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		fail("%v", err)
	}

	if parsePort(port) != nil {
		fail("'%s' is not a valid port number (1-65535)", port)
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
	authtoken := flag.String(
		"authtoken",
		"",
		"Authentication token for identifying a premium ngrok.com account")

	server := flag.String(
		"server",
		"ngrok.com:4443",
		"The remote ngrok server")

	httpAuth := flag.String(
		"httpauth",
		"",
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	subdomain := flag.String(
		"subdomain",
		"",
		"Request a custom subdomain from the ngrok server. (HTTP mode only)")

	hostname := flag.String(
		"hostname",
		"",
		"Request a custom hostname from the ngrok server. (HTTP only) (requires CNAME of your DNS)")

	protocol := flag.String(
		"proto",
		"http",
		"The protocol of the traffic over the tunnel {'http', 'tcp'} (default: 'http')")

	webport := flag.Int(
		"webport",
		4040,
		"The port on which the web interface is served")

	logto := flag.String(
		"log",
		"none",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	v := flag.Bool(
		"version",
		false,
		"Print ngrok version and exit")

	flag.Parse()

	if *v {
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	}

	return &Options{
		server:    *server,
		httpAuth:  *httpAuth,
		subdomain: *subdomain,
		localaddr: parseLocalAddr(),
		protocol:  parseProtocol(*protocol),
		webport:   *webport,
		logto:     *logto,
		authtoken: *authtoken,
		hostname:  *hostname,
	}
}
