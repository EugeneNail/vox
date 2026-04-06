package middleware

import (
	"fmt"
	"net/http"
)

// RejectLargeRequest rejects requests whose declared body size exceeds the limit.
func RejectLargeRequest(maxBytes int64, next http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.ContentLength > maxBytes {
			writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(fmt.Sprintf("Request size must not exceed %d bytes", maxBytes)))
			return
		}

		next.ServeHTTP(writer, request)
	}
}
