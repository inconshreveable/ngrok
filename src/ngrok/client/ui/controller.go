/* The controller in the MVC
 */

package ui

import (
	"ngrok/util"
	"sync"
)

type Command struct {
	Code    int
	Payload interface{}
}

const (
	QUIT = iota
	REPLAY
)

type Controller struct {
	// the model sends updates through this broadcast channel
	Updates *util.Broadcast

	// all views put any commands into this channel
	Cmds chan Command

	// all threads may add themself to this to wait for clean shutdown
	Wait *sync.WaitGroup
}

func NewController() *Controller {
	ctl := &Controller{
		Updates: util.NewBroadcast(),
		Cmds:    make(chan Command),
		Wait:    new(sync.WaitGroup),
	}

	return ctl
}

func (ctl *Controller) Update(state State) {
	ctl.Updates.In() <- state
}
