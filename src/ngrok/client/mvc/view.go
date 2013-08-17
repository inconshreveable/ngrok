package mvc

import (
	"sync"
)

type View interface {
	Shutdown(*sync.WaitGroup)
}
