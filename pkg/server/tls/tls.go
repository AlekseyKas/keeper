package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
)

var log logging.Logger

// Prepare - initialize ssl keys if they are not exists
func Prepare() {
	log = zerolog.New().WithPrefix("tls")
	if !helpers.FileExists("./ssl") {
		err := os.MkdirAll("./ssl", 0700)
		if err != nil {
			log.Fatal(err, "ssl folder can not be created")
		}
		err = generate()
		if err != nil {
			log.Fatal(err)
		}
	}
	if !helpers.FileExists("./ssl/server.pem") || !helpers.FileExists("./ssl/server.crt") {
		err := generate()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func generate() error {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Error(err, "private key can not be generated")
		return errors.New("certificate generation failed")
	}
	public := key.PublicKey
	serial, err := rand.Int(rand.Reader, big.NewInt(9999999999))
	if err != nil {
		log.Error(err, "serial number generation failed")
		return errors.New("certificate generation failed")
	}
	certificate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"keeper Inc."},
			CommonName:   "keeper.local",
			Country:      []string{"RU"},
		},
		NotBefore: time.Now().Add(-time.Hour * 24 * 365),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),
	}
	certDer, err := x509.CreateCertificate(rand.Reader, &certificate, &certificate, &public, key)
	if err != nil {
		log.Error(err, "failed to create x509 certificate.")
		return errors.New("certificate generation failed")
	}
	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Error(err, "failed to create x509 key.")
		return errors.New("certificate generation failed")
	}

	certFile, err := os.Create("./ssl/server.crt")
	if err != nil {
		log.Error(err, "certificate cannot be saved")
		return errors.New("certificate generation failed")
	}
	defer func() {
		certFile.Close()
	}()
	keyFile, err := os.Create("./ssl/server.pem")
	if err != nil {
		log.Error(err, "certificate cannot be saved")
		return errors.New("certificate generation failed")
	}
	defer func() {
		keyFile.Close()
	}()
	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	}
	keyBlock := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDer,
	}
	pem.Encode(keyFile, &keyBlock)
	pem.Encode(certFile, &certBlock)
	return nil
}
