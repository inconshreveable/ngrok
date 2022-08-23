package proto

import (
	"github.com/dolfly/ngrok/pkg/ngrok/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(conn.Conn, interface{}) conn.Conn
}
