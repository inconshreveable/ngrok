package server

import (
	"fmt"
	"net"
	"ngrok/conn"
	"ngrok/log"
	"strings"
)

const (
	NotAuthorized = `HTTP/1.0 401 Not Authorized
WWW-Authenticate: Basic realm="ngrok"
Content-Length: 23

Authorization required
`

	NotFound = `HTTP/1.0 404 Not Found
Content-Length: %d

Tunnel %s not found
`

	BadRequest = `HTTP/1.0 400 Bad Request
Content-Length: 12

Bad Request
`
)

/**
 * Listens for new http connections from the public internet
 */
func httpListener(addr *net.TCPAddr) {
	// bind/listen for incoming connections
	listener, err := conn.Listen(addr, "pub", nil)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for public http connections on %v", listener.Port)
	for conn := range listener.Conns {
		go httpHandler(conn)
	}
}

/**
 * Handles a new http connection from the public internet
 */
func httpHandler(tcpConn net.Conn) {
	// wrap up the connection for logging
	conn := conn.NewHttp(tcpConn, "pub")

	defer conn.Close()
	defer func() {
		// recover from failures
		if r := recover(); r != nil {
			conn.Warn("httpHandler failed with error %v", r)
		}
	}()

	// read out the http request
	req, err := conn.ReadRequest()
	if err != nil {
		conn.Warn("Failed to read valid http request: %v", err)
		conn.Write([]byte(BadRequest))
		return
	}

	// read out the Host header from the request
	host := strings.ToLower(req.Host)
	conn.Debug("Found hostname %s in request", host)

	// multiplex to find the right backend host
	tunnel := tunnels.Get("http://" + host)
	if tunnel == nil {
		conn.Info("No tunnel found for hostname %s", host)
		conn.Write([]byte(fmt.Sprintf(NotFound, len(host)+18, host)))
		return
	}

	// If the client specified http auth and it doesn't match this request's auth
	// then fail the request with 401 Not Authorized and request the client reissue the
	// request with basic authdeny the request
	if tunnel.regMsg.HttpAuth != "" && req.Header.Get("Authorization") != tunnel.regMsg.HttpAuth {
		conn.Info("Authentication failed: %s", req.Header.Get("Authorization"))
		conn.Write([]byte(NotAuthorized))
		return
	}

	tunnel.HandlePublicConnection(conn)
}
