package version

import (
	"fmt"
)

const (
	Proto = "1"
	Major = "0"
<<<<<<< HEAD
	Minor = "16"
=======
	Minor = "17"
>>>>>>> master
)

func MajorMinor() string {
	return fmt.Sprintf("%s.%s", Major, Minor)
}

func Full() string {
	return fmt.Sprintf("%s-%s.%s", Proto, Major, Minor)
}

func Compat(client string, server string) bool {
	return client == server
}
