// Package security содержит операции с паролями и случайными одноразовыми секретами.
package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"golang.org/x/crypto/bcrypt"
	"unicode/utf8"
)

func NewToken() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}
func TokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func HashPassword(password string) (string, error) {
	if utf8.RuneCountInString(password) < 12 || len(password) > 72 {
		return "", apperror.New(apperror.Validation, "password must contain at least 12 characters and at most 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
