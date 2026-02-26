package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Hash generates password hash using bcrypt algorithm
// https://gowebexamples.com/password-hashing/
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckHash compares hash and password and returns bool value
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
