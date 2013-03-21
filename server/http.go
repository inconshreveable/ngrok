package server

import (
	log "code.google.com/p/log4go"
	"net"
	"ngrok/conn"
)

/**
 * Listens for new http connections from the public internet
 */
func httpListener(addr *net.TCPAddr) {
	// bind/listen for incoming connections
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for public http connections on %v", getTCPPort(listener.Addr()))
	for {
		// accept new public connections
		tcpConn, err := listener.AcceptTCP()

		if err != nil {
			panic(err)
		}

		// handle the new connection asynchronously
		go httpHandler(tcpConn)
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
			conn.Warn("Failed with error %v", r)
		}
	}()

	// read out the http request
	req, err := conn.ReadRequest()
	if err != nil {
		panic(err)
	}
	conn.Debug("Found hostname %s in request", req.Host)

	tunnel := tunnels.Get("http://" + req.Host)

	if tunnel == nil {
		conn.Info("Not tunnel found for hostname %s", req.Host)
		return
	}

	tunnel.HandlePublicConnection(conn)
}
