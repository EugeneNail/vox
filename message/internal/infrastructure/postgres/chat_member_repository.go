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
		SELECT chat_uuid, user_uuid, role, last_seen_revision, joined_at
		FROM chat_members
		WHERE chat_uuid = $1 AND user_uuid = $2
	`

	var member domain.ChatMember
	if err := repository.database.QueryRowContext(ctx, query, chatUuid, userUuid).Scan(
		&member.ChatUuid,
		&member.UserUuid,
		&member.Role,
		&member.LastSeenRevision,
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
		SELECT chat_uuid, user_uuid, role, last_seen_revision, joined_at
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
			&member.LastSeenRevision,
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

// CreateMany persists chat members in PostgreSQL.
func (repository *ChatMemberRepository) CreateMany(ctx context.Context, members []domain.ChatMember) error {
	if len(members) == 0 {
		return nil
	}

	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning create chat members transaction: %w", err)
	}
	defer transaction.Rollback()

	for _, member := range members {
		if _, err := transaction.ExecContext(
			ctx,
			`INSERT INTO chat_members (chat_uuid, user_uuid, role, last_seen_revision, joined_at) VALUES ($1, $2, $3, $4, $5)`,
			member.ChatUuid,
			member.UserUuid,
			member.Role,
			member.LastSeenRevision,
			member.JoinedAt,
		); err != nil {
			return fmt.Errorf("creating member %q for chat %q: %w", member.UserUuid, member.ChatUuid, err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing create chat members transaction: %w", err)
	}

	return nil
}

// SetLastSeenRevision sets the member's last seen revision to the provided value.
func (repository *ChatMemberRepository) SetLastSeenRevision(ctx context.Context, chatUuid uuid.UUID, userUuid uuid.UUID, revision int64) error {
	if _, err := repository.database.ExecContext(
		ctx,
		`UPDATE chat_members
		SET last_seen_revision = $3
		WHERE chat_uuid = $1 AND user_uuid = $2`,
		chatUuid,
		userUuid,
		revision,
	); err != nil {
		return fmt.Errorf(
			"updating last seen revision for member %q in chat %q to revision %d: %w",
			userUuid,
			chatUuid,
			revision,
			err,
		)
	}

	return nil
}

// DeleteByChatUuidAndUserUuid removes a member from a chat in PostgreSQL.
func (repository *ChatMemberRepository) DeleteByChatUuidAndUserUuid(ctx context.Context, chatUuid uuid.UUID, userUuid uuid.UUID) error {
	if _, err := repository.database.ExecContext(
		ctx,
		`DELETE FROM chat_members WHERE chat_uuid = $1 AND user_uuid = $2`,
		chatUuid,
		userUuid,
	); err != nil {
		return fmt.Errorf("deleting member %q from chat %q: %w", userUuid, chatUuid, err)
	}

	return nil
}
