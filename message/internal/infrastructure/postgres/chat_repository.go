package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

type ChatRepository struct {
	database *sql.DB
}

// NewChatRepository constructs a PostgreSQL-backed chat repository.
func NewChatRepository(database *sql.DB) *ChatRepository {
	return &ChatRepository{
		database: database,
	}
}

// FindByUuid returns a chat by UUID or nil when the chat does not exist.
func (repository *ChatRepository) FindByUuid(ctx context.Context, chatUuid uuid.UUID) (*domain.Chat, error) {
	const query = `
		SELECT uuid, name, is_direct, created_at, updated_at, server_uuid
		FROM chats
		WHERE uuid = $1
	`

	var chat domain.Chat

	if err := repository.database.QueryRowContext(ctx, query, chatUuid).Scan(
		&chat.Uuid,
		&chat.Name,
		&chat.IsDirect,
		&chat.CreatedAt,
		&chat.UpdatedAt,
		&chat.ServerUuid,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding chat by uuid %q: %w", chatUuid, err)
	}

	return &chat, nil
}

// Create persists a new chat in PostgreSQL.
func (repository *ChatRepository) Create(ctx context.Context, chat domain.Chat) error {
	_, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO chats (uuid, name, is_direct, created_at, updated_at, server_uuid) VALUES ($1, $2, $3, $4, $5, $6)`,
		chat.Uuid,
		chat.Name,
		chat.IsDirect,
		chat.CreatedAt,
		chat.UpdatedAt,
		chat.ServerUuid,
	)
	if err != nil {
		return fmt.Errorf("creating chat %q: %w", chat.Uuid, err)
	}

	return nil
}
