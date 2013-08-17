package client

import (
	"ngrok/client/controller"
)

func Main() {
	controller := controller.NewController(newClientModel())
	controller.Run()
}
