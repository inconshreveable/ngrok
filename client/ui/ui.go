package ui

import (
	"sync"
)

type Command int

const (
	QUIT = iota
)

type View interface {
	SetUi(*Ui)
}

type Ui struct {
	// the model always updates
	Updates *Broadcast

	// all views put their commands into this channel
	Cmds chan Command

	// all threads may add themself to this to wait for clean shutdown
	Wait *sync.WaitGroup
}

func NewUi(views ...View) *Ui {
	ui := &Ui{
		Updates: NewBroadcast(),
		Cmds:    make(chan Command),
		Wait:    new(sync.WaitGroup),
	}

	for _, v := range views {
		v.SetUi(ui)
	}

	return ui
}
