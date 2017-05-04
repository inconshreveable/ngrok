package server

import (
	"net/http"
	"ngrok/log"

	vhost "github.com/inconshreveable/go-vhost"
	"strings"
	"errors"
)

func messageAPI(req *http.Request) (err error) {
	var message string
	log.Info("MessageAPI Method: %s", req.Method)
	projectName := req.FormValue("projectName")
	if projectName == "" {
		log.Info("ProjectName is empty")
		return
	}
	log.Info("Message API ProjectName: %s", projectName)
	switch method := req.Method; method {
	case "PUT":
		log.Info("Message Update")
		message, err = fetchMessage(projectName)
		if err != nil {
			log.Warn("Error fetching message %s", err.Error())
			return
		}
		ProjectsMessage[projectName] = message
		message, _ = ProjectsMessage[projectName]
		log.Info("Message: %s", message)
	case "DELETE":
		log.Info("Message Delete")
		delete(ProjectsMessage, projectName)
		message, _ := ProjectsMessage[projectName]
		log.Info("Message: %s", message)
		return
	default:
		err = errors.New("Method Not Allowed")
		return
	}
	return
}

func adminRequest(vhostConn *vhost.HTTPConn) (err error) {
	hasuraURL := vhostConn.Request.RequestURI
	if strings.Contains(hasuraURL, "/message") {
		err = messageAPI(vhostConn.Request)
	}
	vhostConn.Free()
	return
}