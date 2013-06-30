package server

import (
	"crypto/tls"
	"io/ioutil"
	"ngrok/server/assets"
	"os"
)

var (
	tlsConfig *tls.Config
)

func init() {
	readOrBytes := func(envVar string, default_path string) []byte {
		f := os.Getenv(envVar)
		if f == "" {
			b, err := assets.ReadAsset(default_path)
			if err != nil {
				panic(err)
			}
			return b
		} else {
			if b, err := ioutil.ReadFile(f); err != nil {
				panic(err)
			} else {
				return b
			}
		}
	}

	crt := readOrBytes("TLS_CRT_FILE", "assets/server/tls/snakeoil.crt")
	key := readOrBytes("TLS_KEY_FILE", "assets/server/tls/snakeoil.key")
	cert, err := tls.X509KeyPair(crt, key)

	if err != nil {
		panic(err)
	}

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
