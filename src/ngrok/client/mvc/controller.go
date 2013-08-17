package mvc

type Controller interface {
	// how the model communicates that it has changed state
	Update(State)

	// instructs the controller to shut the app down
	Shutdown(message string)

	// PlayRequest instructs the model to play requests
	PlayRequest(tunnel *Tunnel, payload []byte)
}
