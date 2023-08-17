package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func GenerateRsaPrivateKey(bits int) *rsa.PrivateKey {
	privkey, _ := rsa.GenerateKey(rand.Reader, bits)
	return (privkey)
}

func GenerateRsaPrivateKeyBytes(bits int) []byte {
	privkey := GenerateRsaPrivateKey(bits)
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	return (privkey_bytes)
}

func GenerateRsaPrivateKeyAsPemBytes(bits int) []byte {
	privkey_bytes := GenerateRsaPrivateKeyBytes(bits)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return (privkey_pem)
}

func GenerateRsaPrivateKeyAsPemStr(bits int) string {
	privkey_pem := GenerateRsaPrivateKeyAsPemBytes(bits)
	return string(privkey_pem)
}

func ConvertRsaPrivateKeyToPemString(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}
