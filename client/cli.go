package client

import (
	"flag"
	"fmt"
	"os"
	"pgrok/version"
)

const usage1 string = `Usage: %s [OPTIONS] <local port or address>
Options:
`

const usage2 string = `
Examples:
	pgrok 80
	pgrok -subdomain=example 8080
	pgrok -proto=tcp 22
	pgrok -hostname="example.com" -httpauth="user:password" 10.0.0.1


Advanced usage: pgrok [OPTIONS] <command> [command args] [...]
Commands:
	pgrok start [tunnel] [...]    Start tunnels by name from config file
	ngork start-all               Start all tunnels defined in config file
	pgrok list                    List tunnel names from config file
	pgrok help                    Print help
	pgrok version                 Print pgrok version

Examples:
	pgrok start www api blog pubsub
	pgrok -log=stdout -config=pgrok.yml start ssh
	pgrok start-all
	pgrok version

`

type Options struct {
	config        string
	logto         string
	loglevel      string
	authtoken     string
	httpauth      string
	hostname      string
	protocol      string
	subdomain     string
	command       string
	inspectaddr   string
	inspectpublic bool
	tls           bool
	tlsClientCrt  string
	tlsClientKey  string
	args          []string
}

func ParseArgs() (opts *Options, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage1, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, usage2)
	}

	config := flag.String(
		"config",
		"",
		"Path to pgrok configuration file. (default: $HOME/.pgrok)")

	logto := flag.String(
		"log",
		"none",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	loglevel := flag.String(
		"log-level",
		"DEBUG",
		"The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	authtoken := flag.String(
		"authtoken",
		"",
		"Authentication token for identifying an pgrok account")

	httpauth := flag.String(
		"httpauth",
		"",
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	subdomain := flag.String(
		"subdomain",
		"",
		"Request a custom subdomain from the pgrok server. (HTTP only)")

	hostname := flag.String(
		"hostname",
		"",
		"Request a custom hostname from the pgrok server. (HTTP only) (requires CNAME of your DNS)")

	protocol := flag.String(
		"proto",
		"http+https",
		"The protocol of the traffic over the tunnel (http+https|https|tcp)")

	tls := flag.Bool(
		"tls",
		false,
		"Use dial for tls port")

	tlsClientCrt := flag.String(
		"tlsClientCrt",
		"",
		"Path to a TLS Client CRT file if server requires")

	tlsClientKey := flag.String(
		"tlsClientKey",
		"",
		"Path to a TLS Client Key file if server requires")

	inspectaddr := flag.String(
		"inspectaddr",
		defaultInspectAddr,
		"The addr for inspect requests")

	inspectpublic := flag.Bool(
		"inspectpublic",
		false,
		"Should export inspector to public access")

	flag.Parse()

	opts = &Options{
		config:        *config,
		logto:         *logto,
		loglevel:      *loglevel,
		httpauth:      *httpauth,
		subdomain:     *subdomain,
		protocol:      *protocol,
		authtoken:     *authtoken,
		hostname:      *hostname,
		inspectaddr:   *inspectaddr,
		inspectpublic: *inspectpublic,
		tls:           *tls,
		tlsClientCrt:  *tlsClientCrt,
		tlsClientKey:  *tlsClientKey,
		command:       flag.Arg(0),
	}

	switch opts.command {
	case "list":
		opts.args = flag.Args()[1:]
	case "start":
		opts.args = flag.Args()[1:]
	case "start-all":
		opts.args = flag.Args()[1:]
	case "version":
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	case "help":
		flag.Usage()
		os.Exit(0)
	case "":
		err = fmt.Errorf("Error: Specify a local port to tunnel to, or " +
			"an pgrok command.\n\nExample: To expose port 80, run " +
			"'pgrok 80'")
		return

	default:
		if len(flag.Args()) > 1 {
			err = fmt.Errorf("You may only specify one port to tunnel to on the command line, got %d: %v",
				len(flag.Args()),
				flag.Args())
			return
		}

		opts.command = "default"
		opts.args = flag.Args()
	}

	return
}
