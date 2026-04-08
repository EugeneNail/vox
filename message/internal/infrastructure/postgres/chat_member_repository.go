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

// FindByChatUuidAndUserUuid returns a chat member or nil when the user is not a member.
func (repository *ChatMemberRepository) FindByChatUuidAndUserUuid(ctx context.Context, chatUuid uuid.UUID, userUuid uuid.UUID) (*domain.ChatMember, error) {
	const query = `
		SELECT chat_uuid, user_uuid, role, joined_at
		FROM chat_members
		WHERE chat_uuid = $1 AND user_uuid = $2
	`

	var member domain.ChatMember
	if err := repository.database.QueryRowContext(ctx, query, chatUuid, userUuid).Scan(
		&member.ChatUuid,
		&member.UserUuid,
		&member.Role,
		&member.JoinedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding member %q in chat %q: %w", userUuid, chatUuid, err)
	}

	return &member, nil
}

// FindAllByChatUuid returns all members of a chat.
func (repository *ChatMemberRepository) FindAllByChatUuid(ctx context.Context, chatUuid uuid.UUID) ([]domain.ChatMember, error) {
	const query = `
		SELECT chat_uuid, user_uuid, role, joined_at
		FROM chat_members
		WHERE chat_uuid = $1
		ORDER BY joined_at ASC
	`

	rows, err := repository.database.QueryContext(ctx, query, chatUuid)
	if err != nil {
		return nil, fmt.Errorf("finding members by chat uuid %q: %w", chatUuid, err)
	}
	defer rows.Close()

	members := make([]domain.ChatMember, 0)
	for rows.Next() {
		var member domain.ChatMember
		if err := rows.Scan(
			&member.ChatUuid,
			&member.UserUuid,
			&member.Role,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning member for chat %q: %w", chatUuid, err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading members by chat uuid %q: %w", chatUuid, err)
	}

	return members, nil
}
