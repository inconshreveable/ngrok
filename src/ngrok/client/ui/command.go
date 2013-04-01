package ui

type Command struct {
	Code    int
	Payload interface{}
}

const (
	QUIT = iota
	REPLAY
)
