package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

type ChatMemberRepository struct {
	database *sql.DB
}

// NewChatMemberRepository constructs a PostgreSQL-backed chat member repository.
func NewChatMemberRepository(database *sql.DB) *ChatMemberRepository {
	return &ChatMemberRepository{
		database: database,
	}
}

// FindByChatUuidAndMemberUuid returns a chat member by chat and member UUIDs or nil when it does not exist.
func (repository *ChatMemberRepository) FindByChatUuidAndMemberUuid(ctx context.Context, chatUuid uuid.UUID, memberUuid uuid.UUID) (*domain.ChatMember, error) {
	const query = `
		SELECT chat_uuid, member_uuid, created_at, updated_at
		FROM chat_members
		WHERE chat_uuid = $1 AND member_uuid = $2
	`

	var chatMember domain.ChatMember

	if err := repository.database.QueryRowContext(ctx, query, chatUuid, memberUuid).Scan(
		&chatMember.ChatUuid,
		&chatMember.MemberUuid,
		&chatMember.CreatedAt,
		&chatMember.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding chat member %q in chat %q: %w", memberUuid, chatUuid, err)
	}

	return &chatMember, nil
}

// Create persists a new chat member in PostgreSQL.
func (repository *ChatMemberRepository) Create(ctx context.Context, chatMember domain.ChatMember) error {
	_, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO chat_members (chat_uuid, member_uuid, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		chatMember.ChatUuid,
		chatMember.MemberUuid,
		chatMember.CreatedAt,
		chatMember.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating chat member %q in chat %q: %w", chatMember.MemberUuid, chatMember.ChatUuid, err)
	}

	return nil
}
