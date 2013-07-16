// +build release autoupdate

package client

import (
	"fmt"
	update "github.com/inconshreveable/go-update"
	"net/url"
	"ngrok/client/ui"
	"ngrok/log"
	"ngrok/version"
	"runtime"
	"time"
)

func autoUpdate(s *State, ctl *ui.Controller) {
	update := func() bool {
		params := make(url.Values)
		params.Add("version", version.MajorMinor())
		params.Add("platform", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))

		download := update.NewDownload()
		downloadComplete := make(chan int)
		go func() {
			for {
				select {
				case progress := <-download.Progress:
					if progress < 100 {
						s.update = ui.UpdateStatus(progress)
					} else {
						s.update = ui.UpdateInstalling
					}
					ctl.Update(s)
				case <-downloadComplete:
					close(downloadComplete)
					return
				}
			}
		}()

		err := download.UpdateFromUrl(updateEndpoint + "?" + params.Encode())
		downloadComplete <- 1
		<-downloadComplete
		if err != nil {
			if err != update.UpdateUnavailable {
				log.Error("Error while updating ngrok: %v", err)
				s.update = ui.UpdateError
			} else {
				s.update = ui.UpdateNone
			}
			ctl.Update(s)
			return false
		} else {
			log.Info("Marked update ready")
			s.update = ui.UpdateReady
			ctl.Update(s)
			return true
		}
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
