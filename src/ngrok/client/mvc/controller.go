package mvc

import (
	"ngrok/util"
)

type Controller interface {
	// how the model communicates that it has changed state
	Update(State)

	// instructs the controller to shut the app down
	Shutdown(message string)

	// PlayRequest instructs the model to play requests
	PlayRequest(tunnel Tunnel, payload []byte)

	// A channel of updates
	Updates() *util.Broadcast

	// returns the current state
	State() State

	// safe wrapper for running go-routines
	Go(fn func())

	// the address where the web inspection interface is running
	GetWebInspectAddr() string
}
