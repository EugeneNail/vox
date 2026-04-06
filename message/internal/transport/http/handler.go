package http

// Handler groups HTTP route handlers for the message service.
type Handler struct{}

// NewHandler constructs a message HTTP handler collection.
func NewHandler() *Handler {
	return &Handler{}
}
