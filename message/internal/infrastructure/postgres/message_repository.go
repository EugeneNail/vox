package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

// FindByUuid returns a message by UUID or nil when the message does not exist.
func (repository *MessageRepository) FindByUuid(ctx context.Context, messageUuid uuid.UUID) (*domain.Message, error) {
	const query = `
		SELECT
			m.uuid,
			m.chat_uuid,
			m.user_uuid,
			m.revision,
			m.text,
			m.created_at,
			m.updated_at,
			a.uuid,
			a.name
		FROM messages m
		LEFT JOIN attachments a ON a.message_uuid = m.uuid
		WHERE m.uuid = $1
		ORDER BY a.uuid
	`

	messages, err := repository.findMessages(ctx, query, "message by uuid", messageUuid)
	if err != nil {
		return nil, fmt.Errorf("finding message by uuid %q: %w", messageUuid, err)
	}

	if len(messages) == 0 {
		return nil, nil
	}

	return &messages[0], nil
}

// ListFromRevision returns chat messages with revision greater than or equal to the provided revision.
func (repository *MessageRepository) ListFromRevision(ctx context.Context, chatUuid uuid.UUID, revision int64) ([]domain.Message, error) {
	const query = `
		WITH scoped_messages AS (
			SELECT uuid, chat_uuid, user_uuid, revision, text, created_at, updated_at
			FROM messages
			WHERE chat_uuid = $1 AND revision >= $2
			ORDER BY revision ASC
		)
		SELECT
			scoped_messages.uuid,
			scoped_messages.chat_uuid,
			scoped_messages.user_uuid,
			scoped_messages.revision,
			scoped_messages.text,
			scoped_messages.created_at,
			scoped_messages.updated_at,
			attachments.uuid,
			attachments.name
		FROM scoped_messages
		LEFT JOIN attachments ON attachments.message_uuid = scoped_messages.uuid
		ORDER BY scoped_messages.revision ASC, attachments.uuid
	`

	messages, err := repository.findMessages(ctx, query, "messages by chat uuid from revision", chatUuid, revision)
	if err != nil {
		return nil, fmt.Errorf("listing messages by chat uuid %q from revision %d: %w", chatUuid, revision, err)
	}

	return messages, nil
}

// Create persists a new message in PostgreSQL.
func (repository *MessageRepository) Create(ctx context.Context, message domain.Message) error {
	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning create message transaction for message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}
	defer transaction.Rollback()

	if _, err := transaction.ExecContext(
		ctx,
		`INSERT INTO messages (uuid, chat_uuid, user_uuid, revision, text, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		message.Uuid,
		message.ChatUuid,
		message.UserUuid,
		message.Revision,
		message.Text,
		message.CreatedAt,
		message.UpdatedAt,
	); err != nil {
		return fmt.Errorf("creating message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	if len(message.Attachments) > 0 {
		placeholders := make([]string, 0, len(message.Attachments))
		arguments := make([]any, 0, len(message.Attachments)*3)

		for index, attachment := range message.Attachments {
			placeholderIndex := index*3 + 1
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", placeholderIndex, placeholderIndex+1, placeholderIndex+2))
			arguments = append(arguments, attachment.Uuid, attachment.MessageUuid, attachment.Name)
		}

		query := "INSERT INTO attachments (uuid, message_uuid, name) VALUES " + strings.Join(placeholders, ", ")
		if _, err := transaction.ExecContext(ctx, query, arguments...); err != nil {
			return fmt.Errorf("creating %d attachments for message %q: %w", len(message.Attachments), message.Uuid, err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing create message transaction for message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	return nil
}

// Update persists mutable message fields in PostgreSQL.
func (repository *MessageRepository) Update(ctx context.Context, message domain.Message) error {
	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning update message transaction for message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}
	defer transaction.Rollback()

	if _, err := transaction.ExecContext(
		ctx,
		`UPDATE messages SET text = $2, updated_at = $3 WHERE uuid = $1`,
		message.Uuid,
		message.Text,
		message.UpdatedAt,
	); err != nil {
		return fmt.Errorf("updating text for message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	if _, err := transaction.ExecContext(ctx, `DELETE FROM attachments WHERE message_uuid = $1`, message.Uuid); err != nil {
		return fmt.Errorf("deleting attachments for message %q: %w", message.Uuid, err)
	}

	if len(message.Attachments) > 0 {
		placeholders := make([]string, 0, len(message.Attachments))
		arguments := make([]any, 0, len(message.Attachments)*3)

		for index, attachment := range message.Attachments {
			placeholderIndex := index*3 + 1
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", placeholderIndex, placeholderIndex+1, placeholderIndex+2))
			arguments = append(arguments, attachment.Uuid, attachment.MessageUuid, attachment.Name)
		}

		query := "INSERT INTO attachments (uuid, message_uuid, name) VALUES " + strings.Join(placeholders, ", ")
		if _, err := transaction.ExecContext(ctx, query, arguments...); err != nil {
			return fmt.Errorf("creating %d attachments for message %q: %w", len(message.Attachments), message.Uuid, err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing update message transaction for message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	return nil
}

// Delete removes the message from PostgreSQL.
func (repository *MessageRepository) Delete(ctx context.Context, messageUuid uuid.UUID) error {
	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning delete message transaction for message %q: %w", messageUuid, err)
	}
	defer transaction.Rollback()

	if _, err := transaction.ExecContext(ctx, `DELETE FROM attachments WHERE message_uuid = $1`, messageUuid); err != nil {
		return fmt.Errorf("deleting attachments for message %q: %w", messageUuid, err)
	}

	if _, err := transaction.ExecContext(ctx, `DELETE FROM messages WHERE uuid = $1`, messageUuid); err != nil {
		return fmt.Errorf("deleting message %q: %w", messageUuid, err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing delete message transaction for message %q: %w", messageUuid, err)
	}

	return nil
}

func (repository *MessageRepository) findMessages(ctx context.Context, query string, scope string, args ...any) ([]domain.Message, error) {
	// TODO: remove this thin wrapper and call scanMessages directly from repository methods.
	rows, err := repository.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("finding %s: %w", scope, err)
	}
	defer rows.Close()

	return scanMessages(rows, scope)
}

func scanMessages(rows *sql.Rows, scope string) ([]domain.Message, error) {
	messages := make([]domain.Message, 0)
	messageIndexes := make(map[uuid.UUID]int)

	for rows.Next() {
		var message domain.Message
		var attachmentUuid uuid.NullUUID
		var attachmentName sql.NullString

		if err := rows.Scan(
			&message.Uuid,
			&message.ChatUuid,
			&message.UserUuid,
			&message.Revision,
			&message.Text,
			&message.CreatedAt,
			&message.UpdatedAt,
			&attachmentUuid,
			&attachmentName,
		); err != nil {
			return nil, fmt.Errorf("scanning %s: %w", scope, err)
		}

		index, exists := messageIndexes[message.Uuid]
		if !exists {
			index = len(messages)
			messageIndexes[message.Uuid] = index
			messages = append(messages, message)
		}

		if attachmentUuid.Valid {
			messages[index].Attachments = append(messages[index].Attachments, domain.Attachment{
				Uuid:        attachmentUuid.UUID,
				Name:        attachmentName.String,
				MessageUuid: message.Uuid,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", scope, err)
	}

	return messages, nil
}
