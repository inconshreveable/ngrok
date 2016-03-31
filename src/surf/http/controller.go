package http

import (
	"encoding/json"
	"errors"
	"github.com/parnurzeal/gorequest"
	"os"
	"strings"
	"surf/msg"
)

type Credentials struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	ApplicationId  string `json:"application_id"`
	ApplicationKey string `json:"application_key"`
}

type Token struct {
	Token string
	User  User
}

type User struct {
	Id                 string
	Username           string
	Name               string
	PrimaryPhoneNumber string
	CreationDate       string
}

type TunnelReqPayload struct {
	ReqId      string `json:"request_id,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Subdomain  string `json:"subdomain,omitempty"`
	HttpAuth   string `json:"http_auth,omitempty"`
	RemotePort uint16 `json:"remote_port,omitempty"`
	Token      string `json:"token,omitempty"`
}

type Request struct {
	Error   string
	Message string
}

func Logout() (bool, error) {
	os.Setenv("DASHBOARD_TOKEN", "")
	url := GetLogoutURL()

	resp, body, requestErr := gorequest.New().
		Get(url).
		End()

	// 422 server error
	if resp.StatusCode == 422 {
		return false, errors.New("Could not login: " + body)
	}

	// Request failed
	if requestErr != nil {
		return false, errors.New("Could not login: request failed")
	}

	return true, nil
}

func Login() (bool, error) {
	os.Setenv("DASHBOARD_TOKEN", "")
	credentials := getCredentials()
	url := GetLoginURL()

	resp, body, requestErr := gorequest.New().
		Post(url).
		Send(credentials).
		End()

	// 422 server error
	if resp.StatusCode == 422 {
		return false, errors.New("Could not login: " + body)
	}

	// Request failed
	if requestErr != nil {
		return false, errors.New("Could not login: request failed")
	}

	// Now parse response and get token
	var token Token
	tokenErr := json.Unmarshal([]byte(body), &token)
	if tokenErr != nil {
		return false, errors.New("Could not login: " + tokenErr.Error())
	}

	// Make sure we have a token
	if len(token.Token) == 0 {
		return false, errors.New("Could not login: " + body)
	}

	// Set token and finish
	os.Setenv("DASHBOARD_TOKEN", token.Token)
	return true, nil
}

func ValidateConnToken(connToken string, reAttempt bool) (bool, error) {
	dashboardToken := os.Getenv("DASHBOARD_TOKEN")

	// We are not logged in; so login first, and then try again
	if len(dashboardToken) == 0 {
		logged, loggingErr := Login()
		if logged == true {
			return ValidateConnToken(connToken, reAttempt)
		} else {
			return false, loggingErr
		}
	} else {
		url := GetValidateTunnelTokenURL()
		payload := `{"token":"` + connToken + `"}`

		resp, body, requestErr := gorequest.New().
			Post(url).
			Set("x-session-token", dashboardToken).
			Send(payload).
			End()

		var jsonBody Request
		json.Unmarshal([]byte(body), &jsonBody)

		// 422 server error
		if resp.StatusCode == 422 {
			// Token was invalid, re-attempt
			if isTokenInvalid(jsonBody) {
				os.Setenv("DASHBOARD_TOKEN", "")
				if reAttempt == true {
					return ValidateConnToken(connToken, false)
				}
			}

			return false, errors.New("Could not validate token: " + body)
		}

		// Request failed
		if requestErr != nil {
			return false, errors.New("Could not validate token: request failed")
		}

		// If message is success, then we are good to go
		if body == `{"message":"success"}` {
			return true, nil
		} else {
			return false, errors.New("Could not validate token: " + body)
		}
	}
}

func ValidateTunnelRequest(reqTunnel *msg.ReqTunnel, connToken string, reAttempt bool) (bool, error) {
	dashboardToken := os.Getenv("DASHBOARD_TOKEN")

	// We are not logged in; so login first, and then try again
	if len(dashboardToken) == 0 {
		logged, loggingErr := Login()
		if logged == true {
			return ValidateTunnelRequest(reqTunnel, connToken, true)
		} else {
			return false, loggingErr
		}
	} else {
		url := GetValidateTunnelRequestURL()
		payload := TunnelReqPayload{
			ReqId:      reqTunnel.ReqId,
			Protocol:   reqTunnel.Protocol,
			Hostname:   reqTunnel.Hostname,
			Subdomain:  reqTunnel.Subdomain,
			HttpAuth:   reqTunnel.HttpAuth,
			RemotePort: reqTunnel.RemotePort,
			Token:      connToken,
		}

		resp, body, requestErr := gorequest.New().
			Post(url).
			Set("x-session-token", dashboardToken).
			Send(payload).
			End()

		var jsonBody Request
		json.Unmarshal([]byte(body), &jsonBody)

		// 422 server error
		if resp.StatusCode == 422 {
			// Token was invalid, re-attempt
			if isTokenInvalid(jsonBody) {
				os.Setenv("DASHBOARD_TOKEN", "")
				if reAttempt == true {
					return ValidateTunnelRequest(reqTunnel, connToken, false)
				}
			}

			return false, errors.New("Could not validate requested tunnel: " + body)
		}

		// Request failed
		if requestErr != nil {
			return false, errors.New("Could not validate requested tunnel: request failed")
		}

		// If message is success, then we are good to go
		if body == `{"message":"success"}` {
			return true, nil
		} else {
			return false, errors.New("Could not validate requested tunnel: " + body)
		}
	}
}

func isTokenInvalid(response Request) bool {
	err := strings.TrimSpace(response.Error)
	if err == "bad session token" || err == "missing session token" {
		return true
	} else {
		return false
	}
}

func getCredentials() Credentials {
	var credentials Credentials
	credentials.Username = Username
	credentials.Password = Password
	credentials.ApplicationId = ApplicationId
	credentials.ApplicationKey = ApplicationKey
	return credentials
}
