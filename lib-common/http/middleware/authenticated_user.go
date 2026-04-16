package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
)

func RequireAuthenticatedUser(next http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authorizationHeader := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			writeUnauthorized(writer)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
		userUuid, err := authentication.UserUuidFromLoginToken(tokenString)
		if err != nil {
			writeUnauthorized(writer)
			return
		}

		next.ServeHTTP(writer, request.WithContext(authentication.ContextWithUserUuid(request.Context(), userUuid)))
	}
}

func writeUnauthorized(writer http.ResponseWriter) {
	response, _ := json.Marshal(map[string]string{
		"message": "invalid token",
	})

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusUnauthorized)
	_, _ = writer.Write(response)
}
