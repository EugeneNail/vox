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
		SELECT uuid, name, avatar, chat_type, revision, created_by_user_uuid, created_at, updated_at
		FROM chats
		WHERE uuid = $1
	`

	var chat domain.Chat
	if err := repository.database.QueryRowContext(ctx, query, chatUuid).Scan(
		&chat.Uuid,
		&chat.Name,
		&chat.Avatar,
		&chat.ChatType,
		&chat.Revision,
		&chat.CreatedByUserUuid,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding chat by uuid %q: %w", chatUuid, err)
	}

	return &chat, nil
}

// FindAllByMemberUuid returns chats that include the given member.
func (repository *ChatRepository) FindAllByMemberUuid(ctx context.Context, memberUuid uuid.UUID) ([]domain.Chat, error) {
	const query = `
		SELECT c.uuid, c.name, c.avatar, c.chat_type, c.revision, c.created_by_user_uuid, c.created_at, c.updated_at
		FROM chats c
		INNER JOIN chat_members cm ON cm.chat_uuid = c.uuid
		WHERE cm.user_uuid = $1
		ORDER BY c.updated_at DESC
	`

	rows, err := repository.database.QueryContext(ctx, query, memberUuid)
	if err != nil {
		return nil, fmt.Errorf("finding chats by member uuid %q: %w", memberUuid, err)
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(
			&chat.Uuid,
			&chat.Name,
			&chat.Avatar,
			&chat.ChatType,
			&chat.Revision,
			&chat.CreatedByUserUuid,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning chat for member uuid %q: %w", memberUuid, err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading chats by member uuid %q: %w", memberUuid, err)
	}

	return chats, nil
}

// FindDirectByMemberUuids returns a direct chat with exactly the given two members or nil when it does not exist.
func (repository *ChatRepository) FindDirectByMemberUuids(ctx context.Context, firstMemberUuid uuid.UUID, secondMemberUuid uuid.UUID) (*domain.Chat, error) {
	const query = `
		SELECT c.uuid, c.name, c.avatar, c.chat_type, c.revision, c.created_by_user_uuid, c.created_at, c.updated_at
		FROM chats c
		WHERE c.chat_type = $3
		  AND 2 = (
			SELECT COUNT(*)
			FROM chat_members cm_all
			WHERE cm_all.chat_uuid = c.uuid
		  )
		  AND EXISTS (
			SELECT 1
			FROM chat_members cm_first
			WHERE cm_first.chat_uuid = c.uuid
			  AND cm_first.user_uuid = $1
		  )
		  AND EXISTS (
			SELECT 1
			FROM chat_members cm_second
			WHERE cm_second.chat_uuid = c.uuid
			  AND cm_second.user_uuid = $2
		  )
		LIMIT 1
	`

	var chat domain.Chat
	if err := repository.database.QueryRowContext(ctx, query, firstMemberUuid, secondMemberUuid, domain.ChatTypeDirect).Scan(
		&chat.Uuid,
		&chat.Name,
		&chat.Avatar,
		&chat.ChatType,
		&chat.Revision,
		&chat.CreatedByUserUuid,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf(
			"finding direct chat by member uuids %q and %q: %w",
			firstMemberUuid,
			secondMemberUuid,
			err,
		)
	}

	return &chat, nil
}

// CreateWithMembers persists a new chat and its initial members in PostgreSQL.
func (repository *ChatRepository) CreateWithMembers(ctx context.Context, chat domain.Chat, members []domain.ChatMember) error {
	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning create chat transaction for chat %q: %w", chat.Uuid, err)
	}
	defer transaction.Rollback()

	if _, err := transaction.ExecContext(
		ctx,
		`INSERT INTO chats (uuid, name, avatar, chat_type, revision, created_by_user_uuid, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chat.Uuid,
		chat.Name,
		chat.Avatar,
		chat.ChatType,
		chat.Revision,
		chat.CreatedByUserUuid,
		chat.CreatedAt,
		chat.UpdatedAt,
	); err != nil {
		return fmt.Errorf("creating chat %q: %w", chat.Uuid, err)
	}

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
		return fmt.Errorf("committing create chat transaction for chat %q: %w", chat.Uuid, err)
	}

	return nil
}

// Update persists chat metadata changes in PostgreSQL.
func (repository *ChatRepository) Update(ctx context.Context, chat domain.Chat) error {
	if _, err := repository.database.ExecContext(
		ctx,
		`UPDATE chats SET name = $2, avatar = $3, revision = $4, updated_at = $5 WHERE uuid = $1`,
		chat.Uuid,
		chat.Name,
		chat.Avatar,
		chat.Revision,
		chat.UpdatedAt,
	); err != nil {
		return fmt.Errorf("updating chat %q: %w", chat.Uuid, err)
	}

	return nil
}

// SetRevision sets the chat revision to the provided value.
func (repository *ChatRepository) SetRevision(ctx context.Context, chatUuid uuid.UUID, revision int64) error {
	if _, err := repository.database.ExecContext(
		ctx,
		`UPDATE chats SET revision = $2 WHERE uuid = $1`,
		chatUuid,
		revision,
	); err != nil {
		return fmt.Errorf("setting revision %d for chat %q: %w", revision, chatUuid, err)
	}

	return nil
}
