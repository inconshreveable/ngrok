// +build release autoupdate

package client

import (
	update "github.com/inconshreveable/go-update"
	"net/http"
	"net/url"
	"ngrok/log"
	"ngrok/version"
	"ngrok/client/mvc"
	"runtime"
	"time"
)

const (
	updateEndpoint = "https://dl.ngrok.com/update"
)

func autoUpdate(ctl mvc.Controller, token string) {
	update := func() (updateSuccessful bool) {
		params := make(url.Values)
		params.Add("version", version.MajorMinor())
		params.Add("os", runtime.GOOS)
		params.Add("arch", runtime.GOARCH)

		download := update.NewDownload()
		downloadComplete := make(chan int)
		go func() {
			for {
				select {
				case progress, ok := <-download.Progress:
					if !ok {
						close(downloadComplete)
						return
					} else if progress == 100 {
						s.update = mvc.UpdateInstalling
						ctl.Update(s)
						close(downloadComplete)
						return
					} else {
						if progress%25 == 0 {
							log.Info("Downloading update %d%% complete", progress)
						}
						s.update = mvc.UpdateStatus(progress)
						ctl.Update(s)
					}
				}
			}
		}()

		log.Info("Checking for update")
		err := download.UpdateFromUrl(updateEndpoint + "?" + params.Encode())
		<-downloadComplete
		if err != nil {
			log.Error("Error while updating ngrok: %v", err)
			if download.Available {
				s.update = mvc.UpdateError
			} else {
				s.update = mvc.UpdateNone
			}

			// record the error to ngrok.com's servers for debugging purposes
			params.Add("error", err.Error())
			params.Add("user", token)
			resp, err := http.PostForm("https://dl.ngrok.com/update/error", params)
			if err != nil {
				log.Error("Error while reporting update error")
			}
			resp.Body.Close()
		} else {
			if download.Available {
				log.Info("Marked update ready")
				s.update = mvc.UpdateReady
				updateSuccessful = true
			} else {
				log.Info("No update available at this time")
				s.update = mvc.UpdateNone
			}
		}

		ctl.Update(s)
		return
	}

	// try to update immediately and then at a set interval
	update()
	for _ = range time.Tick(updateCheckInterval) {
		if update() {
			// stop trying to update if the update function is successful
			// XXX: improve this by trying to download versions newer than
			// the last one we downloaded
			return
		}
	}
}
