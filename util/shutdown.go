package util

import (
	"sync"
)

// A small utility class for managing controlled shutdowns
type Shutdown struct {
	sync.Mutex
	inProgress bool
	begin      chan int // closed when the shutdown begins
	complete   chan int // closed when the shutdown completes
}

func NewShutdown() *Shutdown {
	return &Shutdown{
		begin:    make(chan int),
		complete: make(chan int),
	}
}

func (s *Shutdown) Begin() {
	s.Lock()
	defer s.Unlock()
	if s.inProgress == true {
		return
	} else {
		s.inProgress = true
		close(s.begin)
	}
}

func (s *Shutdown) WaitBegin() {
	<-s.begin
}

func (s *Shutdown) Complete() {
	close(s.complete)
}

func (s *Shutdown) WaitComplete() {
	<-s.complete
}
