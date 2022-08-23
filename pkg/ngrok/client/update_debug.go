//go:build !release && !autoupdate
// +build !release,!autoupdate

package client

import (
	"github.com/dolfly/ngrok/pkg/ngrok/client/mvc"
)

// no auto-updating in debug mode
func autoUpdate(state mvc.State, token string) {
}
