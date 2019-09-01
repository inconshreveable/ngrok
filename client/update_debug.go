// +build !release,!autoupdate

package client

import (
	"ngrok/client/mvc"
)

// no auto-updating in debug mode
func autoUpdate(state mvc.State, token string) {
}
