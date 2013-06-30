package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"ngrok/client/assets"
)

var (
	tlsConfig *tls.Config
)

func init() {
	pool := x509.NewCertPool()

	ngrokRootCrt, err := assets.ReadAsset("assets/client/tls/ngrokroot.crt")
	if err != nil {
		panic(err)
	}

	snakeoilCaCrt, err := assets.ReadAsset("assets/client/tls/snakeoilca.crt")
	if err != nil {
		panic(err)
	}

	for _, b := range [][]byte{ngrokRootCrt, snakeoilCaCrt} {
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

	tlsConfig = &tls.Config{
		RootCAs:    pool,
		ServerName: "tls.ngrok.com",
	}
}
