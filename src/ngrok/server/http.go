package server

import (
	"crypto/tls"
	"fmt"

	vhost "github.com/inconshreveable/go-vhost"
	//"net"
	"ngrok/conn"
	"ngrok/log"
	"ngrok/msg"
	"strings"
	"time"
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
	TunnelReqToClient = `HTTP/1.0 404 Not Found
Content-Length: %d

Tunnel Request for %s sent to client
`

	MessageError = `HTTP/1.0 400 Bad Request

%s
`

	OfflineMessage = `HTTP/1.0 400 Bad Request

%s
`	
)

// Listens for new http(s) connections from the public internet
func startHttpListener(addr string, tlsCfg *tls.Config) (listener *conn.Listener) {
	// bind/listen for incoming connections
	var err error
	if listener, err = conn.Listen(addr, "pub", tlsCfg); err != nil {
		panic(err)
	}

	proto := "http"
	if tlsCfg != nil {
		proto = "https"
	}

	log.Info("Listening for public %s connections on %v", proto, listener.Addr.String())
	go func() {
		for conn := range listener.Conns {
			go httpHandler(conn, proto)
		}
	}()

	return
}

// Handles a new http connection from the public internet
func httpHandler(c conn.Conn, proto string) {
	defer c.Close()
	defer func() {
		// recover from failures
		if r := recover(); r != nil {
			c.Warn("httpHandler failed with error %v", r)
		}
	}()

	// Make sure we detect dead connections while we decide how to multiplex
	c.SetDeadline(time.Now().Add(connReadTimeout))

	// multiplex by extracting the Host header, the vhost library
	vhostConn, err := vhost.HTTP(c)
	if err != nil {
		c.Warn("Failed to read valid %s request: %v", proto, err)
		c.Write([]byte(BadRequest))
		return
	}

	// read out the Host header and auth from the request
	host := strings.ToLower(vhostConn.Host())
	auth := vhostConn.Request.Header.Get("Authorization")

	//fmt.Println(vhostConn.Request.FormValue("test"))

	// handle AdminRequest
	if host == "ngrok.hasura.me" {
		err = adminRequest(vhostConn)
		if err != nil {
			c.Write([]byte(fmt.Sprintf(MessageError, err.Error())))
			return
		}
		c.Write([]byte(fmt.Sprintf(MessageError, "No Error")))
		return
	}

	// done reading mux data, free up the request memory
	vhostConn.Free()

	// We need to read from the vhost conn now since it mucked around reading the stream
	c = conn.Wrap(vhostConn, "pub")

	// multiplex to find the right backend host
	c.Debug("Found hostname %s in request", host)
	tunnel := tunnelRegistry.Get(fmt.Sprintf("%s://%s", proto, host))

	// If tunnel is not found, try to contact the client if project exists else show message
	if tunnel == nil {
		localDevDomain := strings.Split(host, ".")

		// Get projectName from request
		projectName := localDevDomain[len(localDevDomain)-3:][0]
		c.Info("Request Project: %s", projectName)
		tunnel = tunnelRegistry.Get(fmt.Sprintf("%s://%s", proto, "console."+projectName + ".hasura.me"))
		if tunnel == nil {
			c.Info("No tunnel found for hostname %s", host)
			message, _ := GetProjectMessage(c, projectName)
			if message == "" {
				c.Write([]byte(fmt.Sprintf(NotFound, len(host)+18, host)))
				return
			}
			c.Write([]byte(fmt.Sprintf(OfflineMessage, message)))
			return
		}

		// Send NewTunnelReq to client
		tunnel.ctl.out <- &msg.NewTunnelReq{
			CustomService: host,
		}
		c.Info("Tunnel Request sent to client for  %s", host)

		// Wait for the client to request for new tunnel and then try to proxy the new custom service
		c1 := make(chan bool, 1)
		go func() {
			time.Sleep(time.Second * 1)
			c1 <- false
		}()
		select {
		case res := <-c1:
			fmt.Println(res)
			tunnel = tunnelRegistry.Get(fmt.Sprintf("%s://%s", proto, host))
		case <-time.After(time.Second * 2):
			c.Info("%s timeout after 1 second", host)
			c.Write([]byte(fmt.Sprintf(TunnelReqToClient, len(host)+18, host)))
			return
		}
	}

	// If the client specified http auth and it doesn't match this request's auth
	// then fail the request with 401 Not Authorized and request the client reissue the
	// request with basic authdeny the request
	if tunnel.req.HttpAuth != "" && auth != tunnel.req.HttpAuth {
		c.Info("Authentication failed: %s", auth)
		c.Write([]byte(NotAuthorized))
		return
	}

	// dead connections will now be handled by tunnel heartbeating and the client
	c.SetDeadline(time.Time{})

	// let the tunnel handle the connection now
	tunnel.HandlePublicConnection(c)
}
