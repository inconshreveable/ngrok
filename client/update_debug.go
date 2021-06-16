// +build !release,!autoupdate

package client

import (
	"pgrok/client/mvc"
)

// no auto-updating in debug mode
func autoUpdate(state mvc.State, token string) {
}
