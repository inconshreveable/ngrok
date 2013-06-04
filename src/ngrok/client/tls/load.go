package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
)

var (
	Config *tls.Config
)

func init() {
	pool := x509.NewCertPool()
	for _, b := range [][]byte{ngrokRootCrt(), snakeoilCaCrt()} {
		pemBlock, _ := pem.Decode(b)
		if pemBlock == nil {
			panic("Bad PEM data")
		}

		certs, err := x509.ParseCertificates(pemBlock.Bytes)
		if err != nil {
			panic(err)
		}

		pool.AddCert(certs[0])
	}

	Config = &tls.Config{
		RootCAs:    pool,
		ServerName: "tls.ngrok.com",
	}
}
