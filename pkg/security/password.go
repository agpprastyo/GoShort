package security

import (
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt hashing cost
const DefaultCost = 10

// HashPassword creates a bcrypt hash from a plain text password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plaintext password against a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
