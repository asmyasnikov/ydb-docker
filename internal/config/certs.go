package config

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"
)

func exists(files ...string) bool {
	for _, f := range files {
		if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}

type Certs struct {
	Path string
	CA   string
	Cert string
	Key  string
}

func newCerts(ctx context.Context, persist bool) *Certs {
	certs := &Certs{
		Path: envYdbGrpcTlsDataPath(),
		CA:   envYdbGrpcTlsDataPathCaPem(),
		Cert: envYdbGrpcTlsDataPathCertPem(),
		Key:  envYdbGrpcTlsDataPathKeyPem(),
	}

	// check certificates path exists and create it if necessary
	if _, err := os.Stat(certs.Path); errors.Is(err, os.ErrNotExist) {
		if persist {
			err = os.MkdirAll(certs.Path, 0777)
			if err != nil {
				panic(err)
			}
		}
	}

	// check certificate files exists
	if exists(certs.CA, certs.Cert, certs.Key) {
		if persist {
			return certs
		}
	}

	// generate certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject:      pkix.Name{Organization: []string{"localhost"}},
		NotBefore:    time.Now().AddDate(0, 0, 7),
		NotAfter:     time.Now().AddDate(1, 0, 7),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	crt, err := x509.CreateCertificate(rand.Reader,
		&template,
		&template,
		&key.PublicKey,
		key,
	)
	if err != nil {
		panic(err)
	}
	var publicKey, privateKey bytes.Buffer
	if err = pem.Encode(&publicKey, &pem.Block{Type: "CERTIFICATE", Bytes: crt}); err != nil {
		panic(err)
	}
	if err = pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		panic(err)
	}
	if persist {
		if err = os.WriteFile(certs.CA, publicKey.Bytes(), 0644); err != nil {
			panic(err)
		}
		if err = os.WriteFile(certs.Cert, publicKey.Bytes(), 0644); err != nil {
			panic(err)
		}
		if err = os.WriteFile(certs.Key, privateKey.Bytes(), 0644); err != nil {
			panic(err)
		}
	}
	return certs
}
