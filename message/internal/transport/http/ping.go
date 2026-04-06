package http

import "net/http"

// Ping confirms that the message service is reachable.
func (handler *Handler) Ping(request *http.Request) (int, any) {
	return http.StatusOK, map[string]string{
		"message": "pong",
	}
}
