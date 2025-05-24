package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/koan6gi/go-digisign/internal/config"
)

type Signer interface {
	GenCertAndKey() (*x509.Certificate, *rsa.PrivateKey, error)
	Sign(data []byte, key *rsa.PrivateKey) ([]byte, error)
	Verify(data []byte, signature []byte, cert *x509.Certificate) error
	CertToBytes(cert *x509.Certificate) []byte
	KeyToBytes(key *rsa.PrivateKey) []byte
	BytesToCert(data []byte) (*x509.Certificate, error)
	BytesToKey(data []byte) (*rsa.PrivateKey, error)
}

type CryptoError struct {
	Type    int
	Err     error
	content string
}

func (e *CryptoError) Error() string { return e.content }

// types of crypto errors
const (
	GenKeyErr = iota
	GenCertErr
	SignErr
	DataErr
)

type RSASigner struct {
	serialNumber *big.Int
	organization []string
	commonName   string
	keyLength    int
	duration     time.Duration
}

const rsaKeyLength = 2048

func NewRSASigner() *RSASigner {
	return &RSASigner{
		serialNumber: big.NewInt(1),
		organization: []string{config.Organization},
		commonName:   config.CommonName,
		keyLength:    rsaKeyLength,
		duration:     365 * 24 * time.Hour,
	}
}

func (s *RSASigner) GenCertAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, s.keyLength)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenKeyErr,
			Err:     err,
			content: fmt.Sprint("can't create a rsa private key:", err),
		}
	}

	certTemplate := &x509.Certificate{
		SerialNumber: s.serialNumber,
		Subject: pkix.Name{
			Organization: s.organization,
			CommonName:   s.commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(s.duration),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		IsCA:      false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenCertErr,
			Err:     err,
			content: fmt.Sprint("can't create x509 cert:", err),
		}
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenCertErr,
			Err:     err,
			content: fmt.Sprint("can't parse der cert:", err),
		}
	}

	s.serialNumber.Add(s.serialNumber, big.NewInt(1))
	return cert, privateKey, nil
}

func (s *RSASigner) Sign(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, &CryptoError{
			Type:    SignErr,
			Err:     err,
			content: fmt.Sprint("can't sign:", err),
		}
	}
	return signature, nil
}

func (s *RSASigner) Verify(data []byte, signature []byte, cert *x509.Certificate) error {
	hashed := sha256.Sum256(data)
	pubKey := cert.PublicKey.(*rsa.PublicKey)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
}

func (s *RSASigner) CertToBytes(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

func (s *RSASigner) KeyToBytes(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func (s *RSASigner) BytesToCert(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     nil,
			content: fmt.Sprint("can't parse certificate"),
		}
	}

	return x509.ParseCertificate(block.Bytes)
}

func (s *RSASigner) BytesToKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     nil,
			content: fmt.Sprint("can't parse key"),
		}
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
