/* The controller in the MVC
 */

package ui

import (
	"ngrok/util"
	"sync"
)

type Controller struct {
	// the model sends updates through this broadcast channel
	Updates *util.Broadcast

	// all views put any commands into this channel
	Cmds chan Command

	// all threads may add themself to this to wait for clean shutdown
	Wait *sync.WaitGroup

	// channel to signal shutdown
	Shutdown chan int
}

func NewController() *Controller {
	ctl := &Controller{
		Updates:  util.NewBroadcast(),
		Cmds:     make(chan Command),
		Wait:     new(sync.WaitGroup),
		Shutdown: make(chan int),
	}

	return ctl
}

func (ctl *Controller) Update(state State) {
	ctl.Updates.In() <- state
}

func (ctl *Controller) DoShutdown() {
	close(ctl.Shutdown)
}

func (ctl *Controller) IsShuttingDown() bool {
	select {
	case <-ctl.Shutdown:
		return true
	default:
	}
	return false
}
