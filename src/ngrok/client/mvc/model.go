package mvc

import (
	"sync"
)

type Model interface {
	Run(opts *Options, ctl Controller)

	Shutdown(wg *sync.WaitGroup)

	PlayRequest(tunnel *Tunnel, payload []byte)
}
