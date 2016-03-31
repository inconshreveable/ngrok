package util

import (
	"fmt"
	"runtime"
)

const crashMessage = `panic: %v

%s

Oh noes! surf crashed!

Please submit the stack trace and any relevant information to:
github.com/inconshreveable/surf/issues`

//TODO: Fix github link

func MakePanicTrace(err interface{}) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)
	return fmt.Sprintf(crashMessage, err, stackBuf[:n])
}

// Runs the given function and converts any panic encountered while doing so
// into an error. Useful for sending to channels that will close
func PanicToError(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic: %v", r)
		}
	}()
	fn()
	return
}
