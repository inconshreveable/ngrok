package mvc

import (
	"ngrok/msg"
)

type Model interface {
	Run(serverAddr, authToken string, ctl Controller, reqTunnel *msg.ReqTunnel, localaddr string)

	Shutdown()

	PlayRequest(tunnel Tunnel, payload []byte)
}
