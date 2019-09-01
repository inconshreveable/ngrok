package conn

import (
	"bufio"
	"io"
)

// conn.Tee is a wraps a conn.Conn
// causing all writes/reads to be tee'd just
// like the unix command such that all data that
// is read and written to the connection through its
// interfaces will also be copied into two dedicated pipes
// used for consuming a copy of the data stream
//
// this is useful for introspecting the traffic flowing
// over a connection without having to tamper with the actual
// code that reads and writes over the connection
//
// NB: the data is Tee'd into a shared-memory io.Pipe which
// has a limited (and small) buffer. If you are not consuming from
// the ReadBuffer() and WriteBuffer(), you are going to block
// your application's real traffic from flowing over the connection

type Tee struct {
	rd       io.Reader
	wr       io.Writer
	readPipe struct {
		rd *io.PipeReader
		wr *io.PipeWriter
	}
	writePipe struct {
		rd *io.PipeReader
		wr *io.PipeWriter
	}
	Conn
}

func NewTee(conn Conn) *Tee {
	c := &Tee{
		rd:   nil,
		wr:   nil,
		Conn: conn,
	}

	c.readPipe.rd, c.readPipe.wr = io.Pipe()
	c.writePipe.rd, c.writePipe.wr = io.Pipe()

	c.rd = io.TeeReader(c.Conn, c.readPipe.wr)
	c.wr = io.MultiWriter(c.Conn, c.writePipe.wr)
	return c
}

func (c *Tee) ReadBuffer() *bufio.Reader {
	return bufio.NewReader(c.readPipe.rd)
}

func (c *Tee) WriteBuffer() *bufio.Reader {
	return bufio.NewReader(c.writePipe.rd)
}

func (c *Tee) Read(b []byte) (n int, err error) {
	n, err = c.rd.Read(b)
	if err != nil {
		c.readPipe.wr.Close()
	}
	return
}

func (c *Tee) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = io.Copy(c.wr, r)
	if err != nil {
		c.writePipe.wr.Close()
	}
	return
}

func (c *Tee) Write(b []byte) (n int, err error) {
	n, err = c.wr.Write(b)
	if err != nil {
		c.writePipe.wr.Close()
	}
	return
}
