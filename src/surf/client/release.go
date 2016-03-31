// +build release

package client

import "net"

var (
	rootCrtPaths = []string{"assets/client/tls/surfroot.crt"}
)

// server name in release builds is the host part of the server address
func serverName(addr string) string {
	host, _, err := net.SplitHostPort(addr)

	// should never panic because the config parser calls SplitHostPort first
	if err != nil {
		panic(err)
	}

	return host
}
