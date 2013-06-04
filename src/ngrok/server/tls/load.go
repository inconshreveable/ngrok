package tls

import (
	"crypto/tls"
	"io/ioutil"
	"os"
)

var (
	Config *tls.Config
)

func init() {
	readOrBytes := func(envVar string, defaultFile func() []byte) []byte {
		f := os.Getenv(envVar)
		if f == "" {
			return defaultFile()
		} else {
			if b, err := ioutil.ReadFile(f); err != nil {
				panic(err)
			} else {
				return b
			}
		}
	}

	crt := readOrBytes("TLS_CRT_FILE", snakeoilCrt)
	key := readOrBytes("TLS_KEY_FILE", snakeoilKey)
	cert, err := tls.X509KeyPair(crt, key)

	if err != nil {
		panic(err)
	}

	Config = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
