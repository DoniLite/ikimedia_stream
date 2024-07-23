package set

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"path"
	"time"
)


func PrintSomething(msg string) {
	fmt.Println(msg)
}

func RSAKeyGenerator(typeOfKey string) (pubKey string, err error) {
	var filepath string
	var privateFile *os.File
	if typeOfKey == "" {
		return pubKey, path.ErrBadPattern
	}
	filename := time.Now().String() + ".pem"
	switch typeOfKey {
	case "config":
		filepath = path.Join("./data/security/", "config", filename)
	case "service":
		filepath = path.Join("./data/security/", "service", filename)
	case "user":
		filepath = path.Join("./data/security/", "user", filename)
	default:
		return pubKey, http.ErrBodyNotAllowed
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return pubKey, err
	}
	pubKey = privateKey.PublicKey.N.String()
	privateKeyPEM := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateFile, err = os.Create(filepath)
	if err != nil {
		return pubKey, err
	}
	defer privateFile.Close()

	err = pem.Encode(privateFile, &privateKeyPEM)
	if err != nil {
		return pubKey, err
	}
	return pubKey, nil
}

func MakeServerDirs() error {
	allDirs := []string{
		"./data",
		"./data/security",
		"./data/user",
		"./data/user/profile",
		"./data/user/posts",
		"./data/csv",
		"./data/security/config",
		"./data/security/service",
		"./data/security/user",
		"./static",
		"./uploads",
	}
	for _, dir := range(allDirs) {
		err := os.MkdirAll(dir,os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func HashFile(data []byte) string {
	hash := sha256.New()
	result := hash.Sum([]byte("data"))
	return hex.EncodeToString(result)
}

func Mailsender() error {
	client, err := smtp.Dial("mail.ikimedia.com:333")
	if err != nil {
		return err
	}
	auth := smtp.CRAMMD5Auth("Ghostify", "Aziadekey2@mail.auth.server")
	client.Auth(auth)
	return nil
}