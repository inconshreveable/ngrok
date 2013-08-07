package client

/*
   Functions for reading and writing the auth token from the user's
   home directory.
*/

import (
	"github.com/inconshreveable/ngrok/log"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"sync"
)

var (
	once             sync.Once
	currentAuthToken string
	authTokenFile    string
)

func initAuth() {
	user, err := user.Current()

	// user.Current() does not work on linux when cross compilling because
	// it requires CGO; use os.Getenv("HOME") hack until we compile natively
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

// Load the auth token from file
func LoadAuthToken() string {
	once.Do(initAuth)
	return currentAuthToken
}

// Save the auth token to file
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
