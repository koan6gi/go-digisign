package network

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	currentCert *x509.Certificate
	currentKey  *rsa.PrivateKey
)

type GenerateResponse struct {
	Status string `json:"status"`
}

type SignRequest struct {
	Data []byte `form:"data"`
	Key  []byte `form:"key"`
}

type SignResponse struct {
	Signature []byte `json:"signature"`
}

type VerifyRequest struct {
	Data      []byte `form:"data"`
	Signature []byte `form:"signature"`
	Cert      []byte `form:"cert"`
}

func (r *Router) generateHandler(c *gin.Context) {
	cert, key, err := r.signer.GenCertAndKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка генерации ключей",
			"details": err.Error(),
		})
		return
	}

	certPEM := r.signer.CertToBytes(cert)
	keyPEM := r.signer.KeyToBytes(key)

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	addFileToZip(zipWriter, "certificate.pem", certPEM)
	addFileToZip(zipWriter, "private_key.pem", keyPEM)

	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка создания архива",
			"details": err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=credentials.zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

func addFileToZip(zipWriter *zip.Writer, filename string, data []byte) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func (r *Router) signHandler(c *gin.Context) {
	dataFile, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Необходимо загрузить файл данных",
			"details": err.Error(),
		})
		return
	}

	file, err := dataFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка чтения файла данных",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка вычисления хеша",
			"details": err.Error(),
		})
		return
	}
	hash := hasher.Sum(nil)

	keyContent := c.PostForm("key")
	key, err := r.signer.BytesToKey([]byte(keyContent))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверный формат приватного ключа",
			"details": err.Error(),
		})
		return
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка создания подписи",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signature": base64.StdEncoding.EncodeToString(signature),
	})
}

func (r *Router) verifyHandler(c *gin.Context) {
	dataFile, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Необходимо загрузить исходный файл",
			"details": err.Error(),
		})
		return
	}

	file, err := dataFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка чтения исходного файла",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка вычисления хеша",
			"details": err.Error(),
		})
		return
	}
	hash := hasher.Sum(nil)

	signatureContent := c.PostForm("signature")
	if signatureContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Необходимо предоставить подпись",
		})
		return
	}

	signature, err := base64.StdEncoding.DecodeString(signatureContent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверный формат подписи (ожидается base64)",
			"details": err.Error(),
		})
		return
	}

	if len(signature) != 256 { // For RSA-2048
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Неверный размер подписи",
			"expected": 256,
			"actual":   len(signature),
		})
		return
	}

	certContent := c.PostForm("cert")
	cert, err := r.signer.BytesToCert([]byte(certContent))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверный формат сертификата",
			"details": err.Error(),
		})
		return
	}

	pubKey := cert.PublicKey.(*rsa.PublicKey)
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash, signature); err != nil {
		log.Printf("Verify failed: hash=%x signature=%x...", hash, signature[:16])
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Проверка подписи не пройдена",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Подпись верна!",
	})
}
