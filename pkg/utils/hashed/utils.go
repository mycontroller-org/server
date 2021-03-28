package hashed

import "golang.org/x/crypto/bcrypt"

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
