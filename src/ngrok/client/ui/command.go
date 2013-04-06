package ui

type Command interface{}

type CmdQuit struct {
	// display this message after quit
	Message string
}

type CmdRequest struct {
	// the bytes of the request to issue
	Payload []byte
}
