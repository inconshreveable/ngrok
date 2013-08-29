package mvc

import (
	"ngrok/msg"
)

type Model interface {
	Run(serverAddr, authToken string, ctl Controller, reg *msg.RegMsg, localaddr string)

	Shutdown()

	PlayRequest(tunnel Tunnel, payload []byte)
}
