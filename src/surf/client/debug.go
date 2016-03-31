// +build !release

package client

var (
	rootCrtPaths = []string{"assets/client/tls/surfroot.crt", "assets/client/tls/snakeoilca.crt"}
)

// no server name in debug builds so that when you connect it will always work
func serverName(addr string) string {
	return ""
}
