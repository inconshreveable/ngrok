package proto

import (
	"ngrok/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(conn.Conn) conn.Conn
}
