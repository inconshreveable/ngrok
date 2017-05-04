package server

import (
	"ngrok/log"
	hasura "ngrok/hasura"
	"errors"
	"encoding/json"
	"ngrok/conn"
)

var ProjectsMessage = map[string]string{}

func fetchMessage(projectName string) (message string, err error) {
	var userData []NgrokUsers
	m := hasura.HasuraQuery{
		Type: "select",
		Args: hasura.HasuraArgument{
			Table:   "ngrok_users",
			Columns: []string{"message"},
			Where: map[string]string{
				"project_name": projectName,
			},
		},
	}
	/*
	request := gorequest.New()
	resp, body, errs := request.Post("https://data.beta.hasura.io/v1/query").
		Send(m).
		EndBytes()
	if errs != nil {
		err = errors.New(errs[0].Error())
		return
	}
	if resp.StatusCode != 200 {
		log.Info("Status %d", resp.StatusCode)
		json.Unmarshal(body, &hError)
		log.Info("Hasura Error %s", hError.Error)
		err = errors.New(hError.Error)
		return
	}
	*/
	_, body, err := hasura.SendQuery(m)
	if err != nil {
		log.Warn("Error in fetching all messages %s", err.Error())
		return
	}
	json.Unmarshal(body, &userData)
	if len(userData) == 0 {
		log.Info("No Projects Found")
		err = errors.New("No Projects Found")
		return
	}
	message = userData[0].Message
	return
}


func GetProjectMessage(c conn.Conn, ProjectName string) (message string, err error) {
	c.Info("Getting Message for Project %s", ProjectName)
	message, ok := ProjectsMessage[ProjectName]
	if !ok {
		c.Info("Project Not present internally")
		message, err = fetchMessage(ProjectName)
		if err != nil {
			return
		}
		ProjectsMessage[ProjectName] = message
		return
	}
	return
}

func fetchAllProjects(){
	var userData []NgrokUsers
	m := hasura.HasuraQuery{
		Type: "select",
		Args: hasura.HasuraArgument{
			Table:   "ngrok_users",
			Columns: []string{"message", "project_name"},
		},
	}
	_, body, err := hasura.SendQuery(m)
	if err != nil {
		log.Warn("Error in fetching all messages %s", err.Error())
		return
	}
	json.Unmarshal(body, &userData)
	if len(userData) == 0 {
		log.Warn("No Projects Found")
		return
	}
	for _, project := range userData{
		ProjectsMessage[project.ProjectName] = project.Message
	}
	log.Info("Messages Updated")
}