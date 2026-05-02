package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, email string, tokenVersion int, secretKey []byte) (string, error) {
	// 1. Create claims (payload)
	claims := jwt.MapClaims{
		"sub":     userID,
		"email":   email,
		"version": tokenVersion,
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // expiry
		"iat":     time.Now().Unix(),
		"iss":     "your-app-name",
	}

	// 2. Create token with signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Sign token (this generates header.payload.signature)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
