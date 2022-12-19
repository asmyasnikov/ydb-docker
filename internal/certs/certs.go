package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path"
	"time"
)

type Certs struct {
	Path string
	CA   string
	Cert string
	Key  string
}

func (certs *Certs) Persist() error {
	if err := os.RemoveAll(certs.Path); err != nil {
		return err
	}

	if err := os.MkdirAll(certs.Path, 0777); err != nil {
		panic(err)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return err
	}

	// generate certificate
	template := x509.Certificate{
		Version:      tls.VersionTLS12,
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

func caPem(ydbGrpcTlsDataPath string) string {
	return path.Join(ydbGrpcTlsDataPath, "ca.pem")
}

func certPem(ydbGrpcTlsDataPath string) string {
	return path.Join(ydbGrpcTlsDataPath, "cert.pem")
}

func keyPem(ydbGrpcTlsDataPath string) string {
	return path.Join(ydbGrpcTlsDataPath, "key.pem")
}

func New(ydbGrpcTlsDataPath string) *Certs {
	return &Certs{
		Path: ydbGrpcTlsDataPath,
		CA:   caPem(ydbGrpcTlsDataPath),
		Cert: certPem(ydbGrpcTlsDataPath),
		Key:  keyPem(ydbGrpcTlsDataPath),
	}
}
