package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
)

type MessageRepository struct {
	database *sql.DB
}

// NewMessageRepository constructs a PostgreSQL-backed message repository.
func NewMessageRepository(database *sql.DB) *MessageRepository {
	return &MessageRepository{
		database: database,
	}
}

// Create persists a new message in PostgreSQL.
func (repository *MessageRepository) Create(ctx context.Context, message domain.Message) error {
	_, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO messages (uuid, chat_uuid, user_uuid, text, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		message.Uuid,
		message.ChatUuid,
		message.UserUuid,
		message.Text,
		message.CreatedAt,
		message.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	return nil
}
