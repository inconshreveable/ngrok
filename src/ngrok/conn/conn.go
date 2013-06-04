package conn

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"ngrok/log"
)

type Conn interface {
	net.Conn
	log.Logger
	Id() string
}

type tcpConn struct {
	net.Conn
	log.Logger
	id  int32
	typ string
}

type Listener struct {
	*net.TCPAddr
	Conns chan Conn
}

func wrapTcpConn(conn net.Conn, typ string) *tcpConn {
	c := &tcpConn{conn, log.NewPrefixLogger(), rand.Int31(), typ}
	c.AddLogPrefix(c.Id())
	return c
}

func Listen(addr *net.TCPAddr, typ string, tlsCfg *tls.Config) (l *Listener, err error) {
	// listen for incoming connections
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	l = &Listener{
		TCPAddr: listener.Addr().(*net.TCPAddr),
		Conns:   make(chan Conn),
	}

	go func() {
		for {
			tcpConn, err := listener.AcceptTCP()
			if err != nil {
				log.Error("Failed to accept new TCP connection of type %s: %v", typ, err)
				continue
			}

			c := wrapTcpConn(tcpConn, typ)
			if tlsCfg != nil {
				c.Conn = tls.Server(c.Conn, tlsCfg)
			}
			c.Info("New connection from %v", tcpConn.RemoteAddr())
			l.Conns <- c
		}
	}()
	return
}

func Wrap(conn net.Conn, typ string) *tcpConn {
	return wrapTcpConn(conn, typ)
}

func Dial(addr, typ string, tlsCfg *tls.Config) (conn *tcpConn, err error) {
	var (
		tcpAddr *net.TCPAddr
		tcpConn net.Conn
	)

	if tcpAddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return
	}

	if tcpConn, err = net.DialTCP("tcp", nil, tcpAddr); err != nil {
		return
	}

	if tlsCfg != nil {
		tcpConn = tls.Client(tcpConn, tlsCfg)
	}

	conn = wrapTcpConn(tcpConn, typ)
	conn.Debug("New connection to: %v", tcpAddr)
	return conn, nil
}

func (c *tcpConn) Close() error {
	c.Debug("Closing")
	return c.Conn.Close()
}

func (c *tcpConn) Id() string {
	return fmt.Sprintf("%s:%x", c.typ, c.id)
}

func Join(c Conn, c2 Conn) (int64, int64) {
	done := make(chan error)
	pipe := func(to Conn, from Conn, bytesCopied *int64) {
		var err error
		*bytesCopied, err = io.Copy(to, from)
		if err != nil {
			from.Warn("Copied %d bytes to %s before failing with error %v", *bytesCopied, to.Id(), err)
			done <- err
		} else {
			from.Debug("Copied %d bytes from to %s", *bytesCopied, to.Id())
			done <- nil
		}
	}

	var fromBytes, toBytes int64
	go pipe(c, c2, &fromBytes)
	go pipe(c2, c, &toBytes)
	c.Info("Joined with connection %s", c2.Id())
	<-done
	c.Close()
	c2.Close()
	<-done
	return fromBytes, toBytes
}

type httpConn struct {
	*tcpConn
	reqBuf *bytes.Buffer
}

func NewHttp(conn net.Conn, typ string) *httpConn {
	return &httpConn{
		wrapTcpConn(conn, typ),
		bytes.NewBuffer(make([]byte, 0, 1024)),
	}
}

func (c *httpConn) ReadRequest() (*http.Request, error) {
	r := io.TeeReader(c.tcpConn, c.reqBuf)
	return http.ReadRequest(bufio.NewReader(r))
}

func (c *tcpConn) ReadFrom(r io.Reader) (n int64, err error) {
	// special case when we're reading from an http request where
	// we had to parse the request and consume bytes from the socket
	// and store them in a temporary request buffer
	if httpConn, ok := r.(*httpConn); ok {
		if n, err = httpConn.reqBuf.WriteTo(c); err != nil {
			return
		}
	}

	nCopied, err := io.Copy(c.Conn, r)
	n += nCopied
	return
}
