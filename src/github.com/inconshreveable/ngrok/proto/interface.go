package proto

import (
	"github.com/inconshreveable/ngrok/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(conn.Conn) conn.Conn
}
