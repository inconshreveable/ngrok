package proto

import (
	"surf/conn"
)

type Tcp struct{}

func NewTcp() *Tcp {
	return new(Tcp)
}

func (h *Tcp) GetName() string { return "tcp" }

func (h *Tcp) WrapConn(c conn.Conn, ctx interface{}) conn.Conn {
	return c
}
