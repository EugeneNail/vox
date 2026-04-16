package authentication

import (
	"context"
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

var ErrInvalidToken = errors.New("invalid token")

type userUuidContextKey string

const contextUserUuidKey userUuidContextKey = "userUuid"

type TokenSigner struct{}

func NewTokenSigner() *TokenSigner {
	return &TokenSigner{}
}

func (signer *TokenSigner) NewLoginToken(userUuid string) (string, error) {
	return signer.generateToken(loginTokenType, userUuid, loginTokenTTL)
}

func (signer *TokenSigner) NewRefreshToken(userUuid string) (string, error) {
	return signer.generateToken(refreshTokenType, userUuid, refreshTokenTTL)
}

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

func UserUuidFromLoginToken(tokenString string) (uuid.UUID, error) {
	return userUuidFromToken(tokenString, loginTokenType)
}

func UserUuidFromRefreshToken(tokenString string) (uuid.UUID, error) {
	return userUuidFromToken(tokenString, refreshTokenType)
}

func UserUuidFromContext(ctx context.Context) (uuid.UUID, bool) {
	userUuid, ok := ctx.Value(contextUserUuidKey).(uuid.UUID)
	return userUuid, ok
}

func ContextWithUserUuid(ctx context.Context, userUuid uuid.UUID) context.Context {
	return context.WithValue(ctx, contextUserUuidKey, userUuid)
}

func userUuidFromToken(tokenString string, tokenType string) (uuid.UUID, error) {
	if tokenString == "" {
		return uuid.Nil, ErrInvalidToken
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return []byte(tokenSalt), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	actualTokenType, ok := claims["type"].(string)
	if !ok || actualTokenType != tokenType {
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
