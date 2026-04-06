package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenSalt = "TEMP_SALT"
const loginTokenType = "login-token"

type authenticatedUserContextKey string

const userUuidContextKey authenticatedUserContextKey = "userUuid"

// RequireAuthenticatedUser validates a login token and stores the user UUID in request context.
func RequireAuthenticatedUser(next http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authorizationHeader := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			writeUnauthorized(writer)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
		if tokenString == "" {
			writeUnauthorized(writer)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return []byte(tokenSalt), nil
		})
		if err != nil || !token.Valid {
			writeUnauthorized(writer)
			return
		}

		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != loginTokenType {
			writeUnauthorized(writer)
			return
		}

		userUuidString, ok := claims["sub"].(string)
		if !ok || userUuidString == "" {
			writeUnauthorized(writer)
			return
		}

		userUuid, err := uuid.Parse(userUuidString)
		if err != nil {
			writeUnauthorized(writer)
			return
		}

		ctx := context.WithValue(request.Context(), userUuidContextKey, userUuid)
		next.ServeHTTP(writer, request.WithContext(ctx))
	}
}

// UserUuidFromContext extracts the authenticated user UUID from request context.
func UserUuidFromContext(ctx context.Context) (uuid.UUID, bool) {
	userUuid, ok := ctx.Value(userUuidContextKey).(uuid.UUID)
	return userUuid, ok
}

func writeUnauthorized(writer http.ResponseWriter) {
	response, _ := json.Marshal(map[string]string{
		"message": "invalid token",
	})

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusUnauthorized)
	_, _ = writer.Write(response)
}
