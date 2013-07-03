package client

import (
	"io/ioutil"
	"ngrok/log"
	"os"
	"os/user"
	"path"
	"sync"
)

/*
   Functions for reading and writing the auth token from the user's
   home directory.
*/
var (
	once             sync.Once
	currentAuthToken string
	authTokenFile    string
)

func Init() {
	user, err := user.Current()

	// os.Getenv("HOME") hack is here to support osx -> linux cross-compilation
	// because user.Current() only cross compiles correctly from osx -> windows
	homeDir := os.Getenv("HOME")
	if err != nil {
		log.Warn("Failed to get user's home directory: %s", err.Error())
	} else {
		homeDir = user.HomeDir
	}

	authTokenFile = path.Join(homeDir, ".ngrok")

	log.Debug("Reading auth token from file %s", authTokenFile)
	tokenBytes, err := ioutil.ReadFile(authTokenFile)

	if err == nil {
		currentAuthToken = string(tokenBytes)
	} else {
		log.Warn("Failed to read ~/.ngrok for auth token: %s", err.Error())
	}
}

func LoadAuthToken() string {
	once.Do(func() { Init() })
	return currentAuthToken
}

func SaveAuthToken(token string) {
	if token == "" || token == LoadAuthToken() || authTokenFile == "" {
		return
	}

	perms := os.FileMode(0644)
	err := ioutil.WriteFile(authTokenFile, []byte(token), perms)
	if err != nil {
		log.Warn("Failed to write auth token to file %s: %v", authTokenFile, err.Error())
	}
}
