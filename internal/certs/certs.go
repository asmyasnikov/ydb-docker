package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path"
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

func New(persist bool) (_ *Certs, err error) {
	certs := &Certs{
		Path: "/ydb_certs/",
		CA:   "/ydb_certs/ca.pem",
		Cert: "/ydb_certs/cert.pem",
		Key:  "/ydb_certs/key.pem",
	}
	if env, has := os.LookupEnv("YDB_GRPC_TLS_DATA_PATH"); has {
		certs.Path = env
		certs.CA = path.Join(env, "ca.pem")
		certs.Cert = path.Join(env, "cert.pem")
		certs.Key = path.Join(env, "key.pem")
	}

	// check certificates path exists and create it if necessary
	if _, err = os.Stat(certs.Path); errors.Is(err, os.ErrNotExist) {
		if persist {
			err = os.MkdirAll(certs.Path, 0777)
			if err != nil {
				return
			}
		}
	}

	// check certificate files exists
	if exists(certs.CA, certs.Cert, certs.Key) {
		if persist {
			return certs, nil
		} else {
			bb, err := os.ReadFile(certs.CA)
			if err != nil {
				return nil, err
			}
			certs.CA = string(bb)
			bb, err = os.ReadFile(certs.Cert)
			if err != nil {
				return nil, err
			}
			certs.Cert = string(bb)
			bb, err = os.ReadFile(certs.Key)
			if err != nil {
				return nil, err
			}
			certs.Key = string(bb)
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
		return certs, err
	}
	crt, err := x509.CreateCertificate(rand.Reader,
		&template,
		&template,
		&key.PublicKey,
		key,
	)
	if err != nil {
		return certs, err
	}
	var publicKey, privateKey bytes.Buffer
	if err = pem.Encode(&publicKey, &pem.Block{Type: "CERTIFICATE", Bytes: crt}); err != nil {
		return certs, err
	}
	if err = pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return certs, err
	}
	if persist {
		if err = os.WriteFile(certs.CA, publicKey.Bytes(), 0644); err != nil {
			return certs, err
		}
		if err = os.WriteFile(certs.Cert, publicKey.Bytes(), 0644); err != nil {
			return certs, err
		}
		if err = os.WriteFile(certs.Key, privateKey.Bytes(), 0644); err != nil {
			return certs, err
		}
	} else {
		certs.CA = publicKey.String()
		certs.Cert = publicKey.String()
		certs.Key = publicKey.String()
	}
	return certs, nil
}
