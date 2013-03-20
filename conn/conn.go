package conn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"ngrok"
)

type Conn interface {
	net.Conn
	ngrok.Logger
	Id() string
}

type loggedConn struct {
	net.Conn
	ngrok.Logger
	id  int32
	typ string
}

func NewLogged(conn net.Conn, typ string) *loggedConn {
	c := &loggedConn{conn, ngrok.NewPrefixLogger(), rand.Int31(), typ}
	c.AddLogPrefix(c.Id())
	c.Info("New connection from %v", conn.RemoteAddr())
	return c
}

func (c *loggedConn) Close() error {
	c.Debug("Closing")
	return c.Conn.Close()
}

func (c *loggedConn) Id() string {
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

type loggedHttpConn struct {
	*loggedConn
	reqBuf *bytes.Buffer
}

func NewHttp(conn net.Conn, typ string) *loggedHttpConn {
	return &loggedHttpConn{
		NewLogged(conn, typ),
		bytes.NewBuffer(make([]byte, 0, 1024)),
	}
}

func (c *loggedHttpConn) ReadRequest() (*http.Request, error) {
	r := io.TeeReader(c.loggedConn, c.reqBuf)
	return http.ReadRequest(bufio.NewReader(r))
}

func (c *loggedConn) ReadFrom(r io.Reader) (n int64, err error) {
	// special case when we're reading from an http request where
	// we had to parse the request and consume bytes from the socket
	// and store them in a temporary request buffer
	if httpConn, ok := r.(*loggedHttpConn); ok {
		if n, err = httpConn.reqBuf.WriteTo(c); err != nil {
			return
		}
	}

	nCopied, err := io.Copy(c.Conn, r)
	n += nCopied
	return
}
