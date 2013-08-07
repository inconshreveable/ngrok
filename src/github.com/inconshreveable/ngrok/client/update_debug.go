// +build !release,!autoupdate

package client

import (
	"github.com/inconshreveable/ngrok/client/ui"
)

// no auto-updating in debug mode
func autoUpdate(s *State, ctl *ui.Controller, token string) {
}
