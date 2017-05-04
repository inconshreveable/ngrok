package server

import (
	"encoding/json"
	"errors"
	"fmt"
	hasura "ngrok/hasura"

	"github.com/parnurzeal/gorequest"
)

type NgrokUsers struct {
	ProjectName string `json:"project_name" binding:"required"`
	UserID      int64  `json:"user_id" binding:"required"`
	Message     string `json:"message"`
}

func ValidateToken(token string, projectName string) error {
	var hError hasura.HasuraError
	var userData []NgrokUsers
	// Create the object
	m := hasura.HasuraQuery{
		Type: "select",
		Args: hasura.HasuraArgument{
			Table:   "ngrok_users",
			Columns: []string{"project_name", "user_id"},
			Where: map[string]string{
				"project_name": projectName,
			},
		},
	}
	// Send the request to data api
	request := gorequest.New()
	resp, body, errs := request.Post("https://data.beta.hasura.io/v1/query").
		Send(m).
		Set("Authorization", "Bearer "+token).
		Set("X-Hasura-Role", "user").
		EndBytes()
	fmt.Println(resp)
	if errs != nil {
		fmt.Println(errs)
		return errors.New(errs[0].Error())
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.Status)
		json.Unmarshal(body, &hError)
		return errors.New(hError.Message)
	}
	json.Unmarshal(body, &userData)
	if len(userData) == 0 {
		return errors.New("Project is not valid for your token")
	}
	return nil
}
