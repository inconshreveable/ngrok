package util

type Broadcast struct {
	listeners []chan interface{}
	reg       chan (chan interface{})
	in        chan interface{}
}

func NewBroadcast() *Broadcast {
	b := &Broadcast{
		listeners: make([]chan interface{}, 0),
		reg:       make(chan (chan interface{})),
		in:        make(chan interface{}),
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

func (b *Broadcast) In() chan interface{} {
	return b.in
}

func (b *Broadcast) Reg() chan interface{} {
	listener := make(chan interface{})
	b.reg <- listener
	return listener
}
