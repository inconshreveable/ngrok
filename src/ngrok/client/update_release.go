// +build release autoupdate

package client

import (
	update "github.com/inconshreveable/go-update"
	"net/http"
	"net/url"
	"ngrok/client/mvc"
	"ngrok/log"
	"ngrok/version"
	"runtime"
	"time"
)

const (
	updateEndpoint = "https://dl.ngrok.com/update"
	checkEndpoint  = "https://dl.ngrok.com/update/check"
)

func progressWatcher(s mvc.State, progress chan int, complete chan int) {
	for {
		select {
		case pct, ok := <-progress:
			if !ok {
				close(complete)
				return
			} else if pct == 100 {
				s.SetUpdateStatus(mvc.UpdateInstalling)
				close(complete)
				return
			} else {
				if pct%25 == 0 {
					log.Info("Downloading update %d%% complete", pct)
				}
				s.SetUpdateStatus(mvc.UpdateStatus(pct))
			}
		}
	}
}

func autoUpdate(s mvc.State, token string) {
	tryAgain := true

	params := make(url.Values)
	params.Add("version", version.MajorMinor())
	params.Add("os", runtime.GOOS)
	params.Add("arch", runtime.GOARCH)
	params.Add("user", token)

	updateUrl := updateEndpoint + "?" + params.Encode()
	checkUrl := checkEndpoint + "?" + params.Encode()

	update := func() {
		log.Info("Checking for update")
		available, err := update.NewDownload(checkUrl).Check()
		if err != nil {
			log.Error("Error while checking for update: %v", err)
			return
		}

		if !available {
			log.Info("No update available")
			return
		}

		// stop trying after a single download attempt
		// XXX: improve this so the we can:
		// 1. safely update multiple times
		// 2. only retry after a network connection failure
		tryAgain = false

		download := update.NewDownload(updateUrl)
		downloadComplete := make(chan int)

		go progressWatcher(s, download.Progress, downloadComplete)

		log.Info("Trying to update . . .")
		err, errRecover := download.GetAndUpdate()
		<-downloadComplete

		if err != nil {
			// log error to console
			log.Error("Error while updating ngrok: %v", err)
			if errRecover != nil {
				log.Error("Error while recovering from failed ngrok update, your binary may be missing: %v", errRecover.Error())
				params.Add("errorRecover", errRecover.Error())
			}

			// log error to ngrok.com's servers for debugging purposes
			params.Add("error", err.Error())
			resp, reportErr := http.PostForm("https://dl.ngrok.com/update/error", params)
			if err != nil {
				log.Error("Error while reporting update error: %v, %v", err, reportErr)
			}
			resp.Body.Close()

			// tell the user to update manually
			s.SetUpdateStatus(mvc.UpdateAvailable)
		} else {
			if !download.Available {
				// this is the way the server tells us to update manually
				log.Info("Server wants us to update manually")
				s.SetUpdateStatus(mvc.UpdateAvailable)
			} else {
				// tell the user the update is ready
				log.Info("Update ready!")
				s.SetUpdateStatus(mvc.UpdateReady)
			}
		}

		return
	}

	// try to update immediately and then at a set interval
	update()
	for _ = range time.Tick(updateCheckInterval) {
		if !tryAgain {
			break
		}
		update()
	}
}
