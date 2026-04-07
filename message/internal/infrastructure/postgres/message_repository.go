package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
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

// FindLastByChatUuid returns the latest messages from the given chat.
func (repository *MessageRepository) FindLastByChatUuid(ctx context.Context, chatUuid uuid.UUID, length int) ([]domain.Message, error) {
	const query = `
		SELECT uuid, chat_uuid, user_uuid, text, created_at, updated_at
		FROM messages
		WHERE chat_uuid = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := repository.database.QueryContext(ctx, query, chatUuid, length)
	if err != nil {
		return nil, fmt.Errorf("finding last messages by chat uuid %q with length %d: %w", chatUuid, length, err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(
			&message.Uuid,
			&message.ChatUuid,
			&message.UserUuid,
			&message.Text,
			&message.CreatedAt,
			&message.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning message for chat uuid %q: %w", chatUuid, err)
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading messages by chat uuid %q: %w", chatUuid, err)
	}

	return messages, nil
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
