package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"ngrok/server/assets"
)

func LoadTLSConfig(crtPath string, keyPath string, caPath string) (tlsConfig *tls.Config, err error) {
	fileOrAsset := func(path string, default_path string) ([]byte, error) {
		loadFn := ioutil.ReadFile
		if path == "" {
			loadFn = assets.Asset
			path = default_path
		}

		return loadFn(path)
	}

	var (
		crt    []byte
		key    []byte
		caCert []byte
		cert   tls.Certificate
	)

	if crt, err = fileOrAsset(crtPath, "assets/server/tls/snakeoil.crt"); err != nil {
		return
	}

	if key, err = fileOrAsset(keyPath, "assets/server/tls/snakeoil.key"); err != nil {
		return
	}

	if cert, err = tls.X509KeyPair(crt, key); err != nil {
		return
	}

	if caPath == "" {

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

	} else {

		caCert, err = ioutil.ReadFile(caPath)
		if err != nil {
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = &tls.Config{
			// check if this works
			// RootCAs:   caCertPool,
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{cert},
		}
	}
	return
}
