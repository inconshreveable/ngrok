package version

import (
	"fmt"
)

const (
	Proto = "2"
	Major = "3"
	Minor = "0"
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
