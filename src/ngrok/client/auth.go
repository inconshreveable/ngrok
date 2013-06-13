package client

import (
	"io/ioutil"
	"ngrok/log"
	"os"
	"os/user"
	"path"
)

/*
   Functions for reading and writing the auth token from the user's
   home directory.
*/
var (
	currentAuthToken string
	authTokenFile    string
)

func init() {
	user, err := user.Current()
	if err != nil {
		log.Warn("Failed to get user's home directory: %s", err.Error())
		return
	}

	authTokenFile = path.Join(user.HomeDir, ".ngrok")
	tokenBytes, err := ioutil.ReadFile(authTokenFile)

	if err == nil {
		currentAuthToken = string(tokenBytes)
	}
}

func SaveAuthToken(token string) {
	if token == "" || token == currentAuthToken || authTokenFile == "" {
		return
	}

	perms := os.FileMode(0644)
	err := ioutil.WriteFile(authTokenFile, []byte(token), perms)
	if err != nil {
		log.Warn("Failed to write auth token to file %s: %v", authTokenFile, err.Error())
	}
}

func LoadAuthToken() string {
	return currentAuthToken
}
