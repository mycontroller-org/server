package hashed

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"golang.org/x/crypto/bcrypt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	EncryptionIdentity = "ENCRYPTED;MC;AES256;BASE64;"
)

// GenerateHash returns hashed string of the password
func GenerateHash(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// IsValidPassword verifies with the hashed string
func IsValidPassword(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

// Encrypt perform encryption on the plain text
func Encrypt(plainText string) (string, error) {
	if strings.HasPrefix(plainText, EncryptionIdentity) || plainText == "" {
		return plainText, nil
	}
	key := []byte(configuration.CFG.Secret)
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	bytes, err := gcm.Seal(nonce, nonce, []byte(plainText), nil), nil
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", EncryptionIdentity, base64.StdEncoding.EncodeToString(bytes)), nil
}

// Decrypt performs decryption and return plan text
func Decrypt(cipherText string) (string, error) {
	if !strings.HasPrefix(cipherText, EncryptionIdentity) {
		return cipherText, nil
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText[len(EncryptionIdentity):])
	if err != nil {
		return "", err
	}

	key := []byte(configuration.CFG.Secret)
	chiperBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(chiperBlock)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", errors.New("cipherText too short")
	}

	nonce, cipherTextBytes := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherTextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}
