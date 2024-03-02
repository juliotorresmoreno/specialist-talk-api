package utils

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"math/big"
	"strings"

	"github.com/xdg-go/pbkdf2"
)

func getConfig() string {
	return "your_secret_key"
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func generatePepper() []byte {
	return []byte(getConfig())
}

func HashPassword(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	pepper := generatePepper()
	passwordWithPepper := append([]byte(password), pepper...)

	hash := pbkdf2.Key(passwordWithPepper, salt, 1000, 64, sha512.New)

	result := base64.StdEncoding.EncodeToString(hash) + "." + base64.StdEncoding.EncodeToString(salt)
	return result, nil
}

func ComparePassword(password string, storedInfo string) (bool, error) {
	parts := strings.Split(storedInfo, ".")
	if len(parts) != 2 {
		return false, errors.New("invalid storedInfo format")
	}

	storedHash, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	storedSalt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	pepper := generatePepper()
	passwordWithPepper := append([]byte(password), pepper...)

	newHash := pbkdf2.Key(passwordWithPepper, storedSalt, 1000, 64, sha512.New)
	match := subtle.ConstantTimeCompare(newHash, storedHash) == 1

	return match, nil
}

func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result += string(charset[randomIndex.Int64()])
	}

	return result, nil
}
