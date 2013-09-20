package util

import (
	"fmt"
	"runtime"
)

const crashMessage = `Oh noes! ngrok crashed!

Please submit the stack trace and any relevant information to:
github.com/inconshreveable/ngrok/issues

panic: %v

%s`

func MakePanicTrace(err interface{}) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)
	return fmt.Sprintf(crashMessage, err, stackBuf[:n])
}
