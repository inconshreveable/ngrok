package client

import (
	_ "crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"pgrok/client/assets"
)

func LoadTLSRootCAs(rootCertPaths []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, certPath := range rootCertPaths {
		rootCrt, err := assets.Asset(certPath)
		if err != nil {
			return nil, err
		}

		pemBlock, _ := pem.Decode(rootCrt)
		if pemBlock == nil {
			return nil, fmt.Errorf("Bad PEM data")
		}

		certs, err := x509.ParseCertificates(pemBlock.Bytes)
		if err != nil {
			return nil, err
		}

		pool.AddCert(certs[0])
	}

	return pool, nil
}

func LoadTLSCertificate(certPath, certKey string) ([]tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, certKey)
	if err != nil {
		return nil, err
	}
	return []tls.Certificate{cert}, nil
}
