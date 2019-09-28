// +build release

package client

var (
	rootCrtPaths = []string{"assets/client/tls/pgrokroot.crt"}
)

func useInsecureSkipVerify() bool {
	return false
}
