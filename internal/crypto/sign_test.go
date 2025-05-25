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

	if err := signer.Verify(data, signature, cert); err != nil {
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

func TestEndToEnd(t *testing.T) {
    // 1. Генерируем тестовые данные
    testData := []byte("Test data for verification")

    // 2. Подписываем данные
    signature, err := signer.Sign(testData, key)
    if err != nil {
        t.Fatalf("Sign failed: %v", err)
    }

    // 3. Проверяем подпись
    if err := signer.Verify(testData, signature, cert); err != nil {
        t.Fatalf("Verify failed: %v", err)
    }

    // 4. Меняем один байт в данных - проверка должна провалиться
    testData[0] ^= 0xFF
    if err := signer.Verify(testData, signature, cert); err == nil {
        t.Fatal("Verify should fail with modified data")
    }
}
