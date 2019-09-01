package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"ngrok/server/assets"
)

var cyphers = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
}

func LoadTLSConfig(crtPath string, keyPath string) (tlsConfig *tls.Config, err error) {
	fileOrAsset := func(path string, default_path string) ([]byte, error) {
		loadFn := ioutil.ReadFile
		if path == "" {
			loadFn = assets.Asset
			path = default_path
		}

		return loadFn(path)
	}

	var (
		crt  []byte
		key  []byte
		cert tls.Certificate
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

	tlsConfig = &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		CipherSuites:             cyphers,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
	}

	return
}

func LoadTLSConfigServer(crtPath string, keyPath string, caPath string) (tlsConfig *tls.Config, err error) {
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

	if caPath != "" {
		caCert, err = ioutil.ReadFile(caPath)
		if err != nil {
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = &tls.Config{
			RootCAs:                  caCertPool,
			ClientCAs:                caCertPool,
			ClientAuth:               tls.RequireAndVerifyClientCert,
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			CipherSuites:             cyphers,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		}
	} else {

		tlsConfig = &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			CipherSuites:             cyphers,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		}

	}
	return
}
