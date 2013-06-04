package server

import (
	"fmt"
	"net"
	"ngrok/conn"
	"ngrok/log"
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

	// multiplex to find the right backend host
	conn.Debug("Found hostname %s in request", req.Host)
	tunnel := tunnels.Get("http://" + req.Host)
	if tunnel == nil {
		conn.Info("No tunnel found for hostname %s", req.Host)
		conn.Write([]byte(fmt.Sprintf(NotFound, len(req.Host)+18, req.Host)))
		return
	}

	// satisfy auth, if necessary
	conn.Debug("From client: %s", req.Header.Get("Authorization"))
	conn.Debug("To match: %s", tunnel.regMsg.HttpAuth)
	if req.Header.Get("Authorization") != tunnel.regMsg.HttpAuth {
		conn.Info("Authentication failed")
		conn.Write([]byte(NotAuthorized))
		return
	}

	tunnel.HandlePublicConnection(conn)
}
