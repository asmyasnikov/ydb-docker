package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"github.com/asmyasnikov/ydb-docker/internal/env"
	"math/big"
	"os"
	"time"
)

type Certs struct {
	Path string
	CA   string
	Cert string
	Key  string
}

func (certs *Certs) Persist() error {
	// check certificates path exists and create it if necessary
	if _, err := os.Stat(certs.Path); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(certs.Path, 0777)
		if err != nil {
			panic(err)
		}
	}

	hostName, err := os.Hostname()
	if err != nil {
		return err
	}

	// generate certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject:      pkix.Name{Organization: []string{"localhost", hostName}},
		NotBefore:    time.Now().AddDate(-1, 0, 0),
		NotAfter:     time.Now().AddDate(100, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", hostName},
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
	if err = os.WriteFile(certs.CA, publicKey.Bytes(), 0644); err != nil {
		panic(err)
	}
	if err = os.WriteFile(certs.Cert, publicKey.Bytes(), 0644); err != nil {
		panic(err)
	}
	if err = os.WriteFile(certs.Key, privateKey.Bytes(), 0644); err != nil {
		panic(err)
	}
	return nil
}

func New() *Certs {
	return &Certs{
		Path: env.YdbGrpcTlsDataPath(),
		CA:   env.YdbGrpcTlsDataPathCaPem(),
		Cert: env.YdbGrpcTlsDataPathCertPem(),
		Key:  env.YdbGrpcTlsDataPathKeyPem(),
	}
}
