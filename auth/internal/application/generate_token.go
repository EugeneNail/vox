package application

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken signs a JWT for the given token type and expiration time.
func GenerateToken(tokenType string, userUuid string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userUuid,
		"type": tokenType,
		"iat":  time.Now().UTC().Unix(),
		"exp":  expiresAt.Unix(),
	})

	signedToken, err := token.SignedString([]byte(TokenSalt))
	if err != nil {
		return "", fmt.Errorf("signing jwt token: %w", err)
	}

	return signedToken, nil
}
