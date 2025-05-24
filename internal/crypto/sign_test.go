package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"log"
	"os"
	"testing"
)

var (
	signer Signer
	cert   *x509.Certificate
	key    *rsa.PrivateKey
	err    error
	data   []byte = []byte("Test data")
)

func TestMain(m *testing.M) {
	signer = NewRSASigner()

	cert, key, err = signer.GenCertAndKey()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestSigner(t *testing.T) {
	signature, err := signer.Sign(data, key)
	if err != nil {
		t.Errorf("%v,\n", err)
	}

	log.Println("create a digital sign: success")

	if !signer.Verify(data, signature, cert) {
		t.Errorf("%v,\n", err)
	}

	log.Println("verify a digital sign: success")

	keyPEM := signer.KeyToBytes(key)
	log.Println("key to bytes:\n", string(keyPEM))

	_, err = signer.BytesToKey(keyPEM)
	if err != nil {
		t.Errorf("%v,\n", err)
	}
	log.Println("bytes to key: success")

	certPEM := signer.CertToBytes(cert)
	log.Println("cert to bytes:\n", string(certPEM))

	_, err = signer.BytesToCert(certPEM)
	if err != nil {
		t.Errorf("%v,\n", err)
	}
	log.Println("bytes to key: success")
}
