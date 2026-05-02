package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	return salt, err
}

func HashPassword(password string) (string, error) {
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// encode to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// store format (important!)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		64*1024, 1, 4, b64Salt, b64Hash)

	return encoded, nil
}

func CheckPasswordHash(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	var memory uint32
	var time uint32
	var threads uint8

	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	newHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(storedHash)))

	return subtle.ConstantTimeCompare(storedHash, newHash) == 1
}
