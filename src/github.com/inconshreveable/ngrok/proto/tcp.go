package proto

import (
	"github.com/inconshreveable/ngrok/conn"
)

type Tcp struct{}

func NewTcp() *Tcp {
	return new(Tcp)
}

func (h *Tcp) GetName() string { return "tcp" }

func (h *Tcp) WrapConn(c conn.Conn) conn.Conn {
	return c
}
