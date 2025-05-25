package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
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
	IsCertValid(cert *x509.Certificate) bool
}

type CryptoError struct {
	Type    int
	Err     error
	Content string
}

func (e *CryptoError) Error() string { return e.Content }
func (e *CryptoError) Unwrap() error { return e.Err }

const (
	GenKeyErr = iota
	GenCertErr
	SignErr
	VerifyErr
	DataErr
)

type RSASigner struct {
	serialNumber *big.Int
	organization []string
	commonName   string
	keyLength    int
	duration     time.Duration
}

func NewRSASigner() *RSASigner {
	keyLength := config.KeyLength
	duration := config.CertDuration

	return &RSASigner{
		serialNumber: big.NewInt(1),
		organization: []string{config.Organization},
		commonName:   config.CommonName,
		keyLength:    keyLength,
		duration:     duration,
	}
}

func (s *RSASigner) GenCertAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, s.keyLength)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenKeyErr,
			Err:     err,
			Content: "failed to generate RSA private key",
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
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenCertErr,
			Err:     err,
			Content: "failed to create x509 certificate",
		}
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, &CryptoError{
			Type:    GenCertErr,
			Err:     err,
			Content: "failed to parse DER certificate",
		}
	}

	s.serialNumber.Add(s.serialNumber, big.NewInt(1))
	return cert, privateKey, nil
}

func (s *RSASigner) Sign(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	if len(data) == 0 {
		return nil, &CryptoError{
			Type:    SignErr,
			Err:     errors.New("empty data"),
			Content: "data to sign cannot be empty",
		}
	}

	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, &CryptoError{
			Type:    SignErr,
			Err:     err,
			Content: "failed to sign data",
		}
	}
	return signature, nil
}

func (s *RSASigner) Verify(data []byte, signature []byte, cert *x509.Certificate) error {
	if len(data) == 0 || len(signature) == 0 {
		return &CryptoError{
			Type:    VerifyErr,
			Err:     errors.New("empty data or signature"),
			Content: "data and signature cannot be empty",
		}
	}

	if !s.IsCertValid(cert) {
		return &CryptoError{
			Type:    VerifyErr,
			Err:     errors.New("certificate expired or not yet valid"),
			Content: "certificate validation failed",
		}
	}

	hashed := sha256.Sum256(data)
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return &CryptoError{
			Type:    VerifyErr,
			Err:     errors.New("invalid public key type"),
			Content: "certificate does not contain RSA public key",
		}
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature); err != nil {
		return &CryptoError{
			Type:    VerifyErr,
			Err:     err,
			Content: "signature verification failed",
		}
	}
	return nil
}

func (s *RSASigner) IsCertValid(cert *x509.Certificate) bool {
	now := time.Now()
	return now.After(cert.NotBefore) && now.Before(cert.NotAfter)
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
	if len(data) == 0 {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     errors.New("empty data"),
			Content: "certificate data cannot be empty",
		}
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     errors.New("invalid PEM block"),
			Content: "failed to decode PEM block containing certificate",
		}
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     err,
			Content: "failed to parse certificate",
		}
	}
	return cert, nil
}

func (s *RSASigner) BytesToKey(data []byte) (*rsa.PrivateKey, error) {
	if len(data) == 0 {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     errors.New("empty data"),
			Content: "key data cannot be empty",
		}
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     errors.New("invalid PEM block"),
			Content: "failed to decode PEM block containing private key",
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, &CryptoError{
			Type:    DataErr,
			Err:     err,
			Content: "failed to parse private key",
		}
	}
	return key, nil
}
