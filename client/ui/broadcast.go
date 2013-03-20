package ui

type Broadcast struct {
	listeners []chan State
	reg       chan (chan State)
	in        chan State
}

func NewBroadcast() *Broadcast {
	b := &Broadcast{
		listeners: make([]chan State, 0),
		reg:       make(chan (chan State)),
		in:        make(chan State),
	}

	go func() {
		for {
			select {
			case l := <-b.reg:
				b.listeners = append(b.listeners, l)

			case item := <-b.in:
				for _, l := range b.listeners {
					l <- item
				}
			}
		}
	}()

	return b
}

func (b *Broadcast) In() chan State {
	return b.in
}

func (b *Broadcast) Reg() chan State {
	listener := make(chan State)
	b.reg <- listener
	return listener
}
