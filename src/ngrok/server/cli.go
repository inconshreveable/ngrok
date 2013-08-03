package server

import (
	"flag"
)

type Options struct {
	httpPort   int
	httpsPort  int
	tunnelPort int
	domain     string
	logto      string
}

func parseArgs() *Options {
	httpPort := flag.Int("httpPort", 80, "Public HTTP port, -1 to disable")
	httpsPort := flag.Int("httpsPort", 443, "Public HTTPS port, -1 to disable")
	tunnelPort := flag.Int("tunnelPort", 4443, "Port to which ngrok clients connect")
	domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")
	logto := flag.String(
		"log",
		"stdout",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	flag.Parse()

	return &Options{
		httpPort:   *httpPort,
		httpsPort:  *httpsPort,
		tunnelPort: *tunnelPort,
		domain:     *domain,
		logto:      *logto,
	}
}
