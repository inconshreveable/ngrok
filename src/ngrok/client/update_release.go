// +build release autoupdate

package client

import (
	"ngrok/client/mvc"
	"ngrok/log"
	"ngrok/version"
	"time"

	"gopkg.in/inconshreveable/go-update.v0"
	"gopkg.in/inconshreveable/go-update.v0/check"
)

const (
	appId          = "ap_pJSFC5wQYkAyI0FIVwKYs9h1hW"
	updateEndpoint = "https://api.equinox.io/1/Updates"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Gx8r9no1QBtCruJW2tu
082MJJ5ZA7k803GisR2c6WglPOD1b/+kUg+dx5Y0TKXz+uNlR3GrCxLh8WkoA95M
T38CQldIjoVN/bWP6jzFxL+6BRoKy5L1TcaIf3xb9B8OhwEq60cvFy7BBrLKEHJN
ua/D1S5axgNOAJ8tQ2w8gISICd84ng+U9tNMqIcEjUN89h3Z4zablfNIfVkbqbSR
fnkR9boUaMr6S1w8OeInjWdiab9sUr87GmEo/3tVxrHVCzHB8pzzoZceCkjgI551
d/hHfAl567YhlkQMNz8dawxBjQwCHHekgC8gAvTO7kmXkAm6YAbpa9kjwgnorPEP
ywIDAQAB
-----END PUBLIC KEY-----`

func autoUpdate(s mvc.State, token string) {
	up, err := update.New().VerifySignatureWithPEM([]byte(publicKey))
	if err != nil {
		log.Error("Failed to create update with signature: %v", err)
		return
	}

	update := func() (tryAgain bool) {
		log.Info("Checking for update")
		params := check.Params{
			AppId:      appId,
			AppVersion: version.MajorMinor(),
			UserId:     token,
		}

		result, err := params.CheckForUpdate(updateEndpoint, up)
		if err == check.NoUpdateAvailable {
			log.Info("No update available")
			return true
		} else if err != nil {
			log.Error("Error while checking for update: %v", err)
			return true
		}

		if result.Initiative == check.INITIATIVE_AUTO {
			if err := up.CanUpdate(); err != nil {
				log.Error("Can't update: insufficient permissions: %v", err)
				// tell the user to update manually
				s.SetUpdateStatus(mvc.UpdateAvailable)
			} else {
				applyUpdate(s, result)
			}
		} else if result.Initiative == check.INITIATIVE_MANUAL {
			// this is the way the server tells us to update manually
			log.Info("Server wants us to update manually")
			s.SetUpdateStatus(mvc.UpdateAvailable)
		} else {
			log.Info("Update available, but ignoring")
		}

		// stop trying after a single download attempt
		// XXX: improve this so the we can:
		// 1. safely update multiple times
		// 2. only retry after temporary errors
		return false
	}

	// try to update immediately and then at a set interval
	for {
		if tryAgain := update(); !tryAgain {
			break
		}

		time.Sleep(updateCheckInterval)
	}
}

func applyUpdate(s mvc.State, result *check.Result) {
	err, errRecover := result.Update()
	if err == nil {
		log.Info("Update ready!")
		s.SetUpdateStatus(mvc.UpdateReady)
		return
	}

	log.Error("Error while updating ngrok: %v", err)
	if errRecover != nil {
		log.Error("Error while recovering from failed ngrok update, your binary may be missing: %v", errRecover.Error())
	}

	// tell the user to update manually
	s.SetUpdateStatus(mvc.UpdateAvailable)
}
