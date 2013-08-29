package util

import (
	"fmt"
	"runtime"
)

func MakePanicTrace(err interface{}) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)
	return fmt.Sprintf("panic: %v\n\n%s", err, stackBuf[:n])
}
