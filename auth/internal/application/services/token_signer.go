package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenSalt = "TEMP_SALT"
const loginTokenType = "login-token"
const refreshTokenType = "refresh-token"
const loginTokenTTL = 15 * time.Minute
const refreshTokenTTL = 7 * 24 * time.Hour

// ErrInvalidToken is returned when a signed token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// TokenSigner issues signed JWT tokens for auth use-cases.
type TokenSigner struct{}

// NewTokenSigner constructs a token signer.
func NewTokenSigner() *TokenSigner {
	return &TokenSigner{}
}

// NewLoginToken signs a short-lived login token for a user.
func (signer *TokenSigner) NewLoginToken(userUuid string) (string, error) {
	return signer.generateToken(loginTokenType, userUuid, loginTokenTTL)
}

// NewRefreshToken signs a refresh token for a user.
func (signer *TokenSigner) NewRefreshToken(userUuid string) (string, error) {
	return signer.generateToken(refreshTokenType, userUuid, refreshTokenTTL)
}

// generateToken signs a JWT for the given token type and ttl.
func (signer *TokenSigner) generateToken(tokenType string, userUuid string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userUuid,
		"type": tokenType,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	})

	signedToken, err := token.SignedString([]byte(tokenSalt))
	if err != nil {
		return "", fmt.Errorf("signing jwt token: %w", err)
	}

	return signedToken, nil
}

// ValidateRefreshToken validates a refresh token and returns its subject UUID.
func (signer *TokenSigner) ValidateRefreshToken(refreshToken string) (uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return []byte(tokenSalt), nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != refreshTokenType {
		return uuid.Nil, ErrInvalidToken
	}

	userUuidString, ok := claims["sub"].(string)
	if !ok || userUuidString == "" {
		return uuid.Nil, ErrInvalidToken
	}

	userUuid, err := uuid.Parse(userUuidString)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userUuid, nil
}
