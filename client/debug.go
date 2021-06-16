// +build !release

package client

var (
	rootCrtPaths = []string{"assets/client/tls/pgrokroot.crt", "assets/client/tls/snakeoilca.crt"}
)

func useInsecureSkipVerify() bool {
	return true
}
