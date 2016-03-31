// +build !release,!autoupdate

package client

import (
	"surf/client/mvc"
)

// no auto-updating in debug mode
func autoUpdate(state mvc.State, token string) {
}
