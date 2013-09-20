package mvc

type Model interface {
	Run()

	Shutdown()

	PlayRequest(tunnel Tunnel, payload []byte)
}
