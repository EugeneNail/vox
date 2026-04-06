package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

type DirectChatRepository struct {
	database *sql.DB
}

// NewDirectChatRepository constructs a PostgreSQL-backed direct chat repository.
func NewDirectChatRepository(database *sql.DB) *DirectChatRepository {
	return &DirectChatRepository{
		database: database,
	}
}

// FindByUuid returns a direct chat by UUID or nil when the direct chat does not exist.
func (repository *DirectChatRepository) FindByUuid(ctx context.Context, chatUuid uuid.UUID) (*domain.DirectChat, error) {
	const query = `
		SELECT uuid, first_member_uuid, second_member_uuid, created_at, updated_at
		FROM direct_chats
		WHERE uuid = $1
	`

	var chat domain.DirectChat

	if err := repository.database.QueryRowContext(ctx, query, chatUuid).Scan(
		&chat.Uuid,
		&chat.FirstMemberUuid,
		&chat.SecondMemberUuid,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding direct chat by uuid %q: %w", chatUuid, err)
	}

	return &chat, nil
}

// FindAllByMemberUuid returns direct chats that include the given member.
func (repository *DirectChatRepository) FindAllByMemberUuid(ctx context.Context, memberUuid uuid.UUID) ([]domain.DirectChat, error) {
	const query = `
		SELECT uuid, first_member_uuid, second_member_uuid, created_at, updated_at
		FROM direct_chats
		WHERE first_member_uuid = $1 OR second_member_uuid = $1
		ORDER BY updated_at DESC
	`

	rows, err := repository.database.QueryContext(ctx, query, memberUuid)
	if err != nil {
		return nil, fmt.Errorf("finding direct chats by member uuid %q: %w", memberUuid, err)
	}
	defer rows.Close()

	chats := make([]domain.DirectChat, 0)
	for rows.Next() {
		var chat domain.DirectChat
		if err := rows.Scan(
			&chat.Uuid,
			&chat.FirstMemberUuid,
			&chat.SecondMemberUuid,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning direct chat for member uuid %q: %w", memberUuid, err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading direct chats by member uuid %q: %w", memberUuid, err)
	}

	return chats, nil
}

// Create persists a new direct chat in PostgreSQL.
func (repository *DirectChatRepository) Create(ctx context.Context, chat domain.DirectChat) error {
	_, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO direct_chats (uuid, first_member_uuid, second_member_uuid, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		chat.Uuid,
		chat.FirstMemberUuid,
		chat.SecondMemberUuid,
		chat.CreatedAt,
		chat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating direct chat %q: %w", chat.Uuid, err)
	}

	return nil
}
