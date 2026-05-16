package set_last_seen_revision

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler sets the authenticated member's last seen revision in a chat.
type Handler struct {
	chatRepository                   domain.ChatRepository
	chatMemberRepository             domain.ChatMemberRepository
	lastSeenRevisionUpdatedPublisher events.LastSeenRevisionUpdatedPublisher
}

// Command contains the input required to set the member's last seen revision.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Revision int64
}

// NewHandler constructs a set_last_seen_revision handler with its dependencies.
func NewHandler(
	chatRepository domain.ChatRepository,
	chatMemberRepository domain.ChatMemberRepository,
	lastSeenRevisionUpdatedPublisher events.LastSeenRevisionUpdatedPublisher,
) *Handler {
	return &Handler{
		chatRepository:                   chatRepository,
		chatMemberRepository:             chatMemberRepository,
		lastSeenRevisionUpdatedPublisher: lastSeenRevisionUpdatedPublisher,
	}
}

// Handle validates input, authorizes access and stores the provided last seen revision.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()
	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	if command.Revision < 0 {
		validationError.AddViolation("revision", "Must be greater than or equal to zero")
	}

	if len(validationError.Violations()) > 0 {
		return validationError
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, command.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", command.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	validationError = validation.NewError()
	if command.Revision > chat.Revision {
		validationError.AddViolation("revision", fmt.Sprintf("Must not exceed chat revision %d", chat.Revision))
	}

	if len(validationError.Violations()) > 0 {
		return validationError
	}

	if member.LastSeenRevision < command.Revision {
		member.LastSeenRevision = command.Revision
	}

	if err := handler.chatMemberRepository.SetLastSeenRevision(ctx, command.ChatUuid, command.UserUuid, member.LastSeenRevision); err != nil {
		return fmt.Errorf(
			"setting last seen revision %d for member %q in chat %q: %w",
			member.LastSeenRevision,
			command.UserUuid,
			command.ChatUuid,
			err,
		)
	}

	if err := handler.lastSeenRevisionUpdatedPublisher.Publish(ctx, events.LastSeenRevisionUpdated{
		ChatUuid:         command.ChatUuid,
		UserUuid:         command.UserUuid,
		LastSeenRevision: member.LastSeenRevision,
	}); err != nil {
		return fmt.Errorf(
			"publishing last seen revision updated event for member %q in chat %q: %w",
			command.UserUuid,
			command.ChatUuid,
			err,
		)
	}

	return nil
}
