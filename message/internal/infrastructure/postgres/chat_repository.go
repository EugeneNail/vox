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

// FindAllDirectByMemberUuid returns direct chats that include the given member.
func (repository *ChatRepository) FindAllDirectByMemberUuid(ctx context.Context, memberUuid uuid.UUID) ([]domain.Chat, error) {
	const query = `
		SELECT chats.uuid, chats.name, chats.is_direct, chats.created_at, chats.updated_at, chats.server_uuid
		FROM chats
		INNER JOIN chat_members ON chat_members.chat_uuid = chats.uuid
		WHERE chat_members.member_uuid = $1 AND chats.is_direct = TRUE
		ORDER BY chats.updated_at DESC
	`

	rows, err := repository.database.QueryContext(ctx, query, memberUuid)
	if err != nil {
		return nil, fmt.Errorf("finding direct chats by member uuid %q: %w", memberUuid, err)
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.Uuid, &chat.Name, &chat.IsDirect, &chat.CreatedAt, &chat.UpdatedAt, &chat.ServerUuid); err != nil {
			return nil, fmt.Errorf("scanning direct chat for member uuid %q: %w", memberUuid, err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading direct chats by member uuid %q: %w", memberUuid, err)
	}

	return chats, nil
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
