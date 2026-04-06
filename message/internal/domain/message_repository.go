package domain

import "context"

// MessageRepository describes message persistence required by the domain and application.
type MessageRepository interface {
	Create(ctx context.Context, message Message) error
}
