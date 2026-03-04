package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+base64.StdEncoding.EncodeToString(salt)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(salt) + "$" + string(hashedPassword), nil
}

func VerifyPassword(password, hashedPassword string) bool {
	parts := splitHash(hashedPassword)
	if len(parts) != 2 {
		return false
	}

	salt := parts[0]
	hash := parts[1]

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt))
	return err == nil
}

func splitHash(hashedPassword string) []string {
	var parts []string
	for i := 0; i < len(hashedPassword); i++ {
		if hashedPassword[i] == '$' {
			parts = append(parts, hashedPassword[:i])
			parts = append(parts, hashedPassword[i+1:])
			break
		}
	}
	return parts
}

func GenerateVerificationCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Read([]byte{0})
		code[i] = byte('0' + n%10)
	}
	return string(code)
}
