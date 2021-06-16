package server

import (
	"flag"
)

type Options struct {
	httpAddr          string
	httpsAddr         string
	tunnelAddr        string
	tunnelTLSClientCA string
	domain            string
	tlsCrt            string
	tlsKey            string
	tlsClientCA       string
	logto             string
	loglevel          string
}

func parseArgs() *Options {
	httpAddr := flag.String("httpAddr", ":80", "Public address for HTTP connections, empty string to disable")
	httpsAddr := flag.String("httpsAddr", ":443", "Public address listening for HTTPS connections, emptry string to disable")
	tunnelAddr := flag.String("tunnelAddr", ":4443", "Public address listening for pgrok client")
	tunnelTLSClientCA := flag.String("tunnelTLSClientCA", "", "Path to a TLS Client CA file if you want enable mutual auth for tunnel")
	domain := flag.String("domain", "ejemplo.me", "Domain where the tunnels are hosted")
	tlsCrt := flag.String("tlsCrt", "", "Path to a TLS certificate file")
	tlsKey := flag.String("tlsKey", "", "Path to a TLS key file")
	tlsClientCA := flag.String("tlsClientCA", "", "Path to a TLS Client CA file if you want enable mutual auth for subdomains")
	logto := flag.String("log", "stdout", "Write log messages to this file. 'stdout' and 'none' have special meanings")
	loglevel := flag.String("log-level", "DEBUG", "The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")
	flag.Parse()

	return &Options{
		httpAddr:          *httpAddr,
		httpsAddr:         *httpsAddr,
		tunnelAddr:        *tunnelAddr,
		tunnelTLSClientCA: *tunnelTLSClientCA,
		domain:            *domain,
		tlsCrt:            *tlsCrt,
		tlsKey:            *tlsKey,
		tlsClientCA:       *tlsClientCA,
		logto:             *logto,
		loglevel:          *loglevel,
	}
}
