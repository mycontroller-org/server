package hashed

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	DefaultEncryptionPrefix = "ENCRYPTED;MC;AES256;BASE64;"
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
func Encrypt(plainText, secret, encryptionPrefix string) (string, error) {
	if encryptionPrefix == "" {
		encryptionPrefix = DefaultEncryptionPrefix
	}
	if strings.HasPrefix(plainText, encryptionPrefix) || plainText == "" {
		return plainText, nil
	}
	key := []byte(secret)
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

	return fmt.Sprintf("%s%s", encryptionPrefix, base64.StdEncoding.EncodeToString(bytes)), nil
}

// Decrypt performs decryption and return plan text
func Decrypt(cipherText, secret, encryptionPrefix string) (string, error) {
	if encryptionPrefix == "" {
		encryptionPrefix = DefaultEncryptionPrefix
	}
	if !strings.HasPrefix(cipherText, encryptionPrefix) {
		return cipherText, nil
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText[len(encryptionPrefix):])
	if err != nil {
		return "", err
	}

	key := []byte(secret)
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
