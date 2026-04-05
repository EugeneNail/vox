package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// HttpHandler returns an HTTP status code and a response payload.
type HttpHandler func(request *http.Request) (status int, payload any)

// WriteJsonResponse wraps a HttpHandler and writes its result as JSON.
func WriteJsonResponse(handler HttpHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		status, payload := handler(request)

		var response []byte
		var err error

		if responseError, ok := payload.(error); ok {
			response, err = json.Marshal(map[string]string{
				"message": responseError.Error(),
			})
		} else {
			response, err = json.Marshal(payload)
		}

		if err != nil {
			status = http.StatusInternalServerError
			response = []byte(fmt.Sprintf(`{"message":"failed to encode response: %s"}`, err.Error()))
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		if _, err := writer.Write(response); err != nil {
			log.Printf("writing json response: %v", err)
		}
	}
}
