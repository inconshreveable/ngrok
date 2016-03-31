package client

import (
	"flag"
	"fmt"
	"os"
	"surf/version"
)

const usage1 string = `Usage: %s [OPTIONS] <local port or address>
Options:
`

const usage2 string = `
Examples:
	surf 80
	surf -proto=tcp 22


Advanced usage: surf [OPTIONS] <command> [command args] [...]
Commands:
	surf start [tunnel] [...]    Start tunnels by name from config file
	surf start-all               Start all tunnels defined in config file
	surf list                    List tunnel names from config file
	surf help                    Print help
	surf version                 Print surf version

Examples:
	surf start www api blog pubsub
	surf -log=stdout -config=surf.yml start ssh
	surf start-all
	surf version

`

type Options struct {
	config    string
	logto     string
	loglevel  string
	authtoken string
	httpauth  string
	hostname  string
	protocol  string
	subdomain string
	command   string
	args      []string
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
		"Path to surf configuration file. (default: $HOME/.surf)")

	logto := flag.String(
		"log",
		"none",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	loglevel := flag.String(
		"log-level",
		"DEBUG",
		"The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	httpauth := flag.String(
		"httpauth",
		"",
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	hostname := flag.String(
		"hostname",
		"",
		"Request a custom hostname from the surf server. (HTTP only) (requires CNAME of your DNS)")

	protocol := flag.String(
		"proto",
		"http+https",
		"The protocol of the traffic over the tunnel {'http', 'https', 'tcp'} (default: 'http+https')")

	flag.Parse()

	opts = &Options{
		config:    *config,
		logto:     *logto,
		loglevel:  *loglevel,
		httpauth:  *httpauth,
		subdomain: Subdomain,
		protocol:  *protocol,
		authtoken: AuthToken,
		hostname:  *hostname,
		command:   flag.Arg(0),
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
			"an surf command.\n\nExample: To expose port 80, run " +
			"'surf 80'")
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
