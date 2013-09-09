package server

import (
	"flag"
)

type Options struct {
	httpAddr   string
	httpsAddr  string
	tunnelAddr string
	domain     string
	logto      string
}

func parseArgs() *Options {
	httpAddr := flag.String("httpAddr", ":80", "Public address for HTTP connections, empty string to disable")
	httpsAddr := flag.String("httpsAddr", ":443", "Public address listening for HTTPS connections, emptry string to disable")
	tunnelAddr := flag.String("tunnelAddr", ":4443", "Public address listening for ngrok client")
	domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")
	logto := flag.String(
		"log",
		"stdout",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	flag.Parse()

	return &Options{
		httpAddr:   *httpAddr,
		httpsAddr:  *httpsAddr,
		tunnelAddr: *tunnelAddr,
		domain:     *domain,
		logto:      *logto,
	}
}
