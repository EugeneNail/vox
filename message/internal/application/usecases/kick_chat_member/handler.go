package kick_chat_member

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")
var ErrChatMemberNotFound = errors.New("chat member not found")

// Handler removes a member from a group chat through the kick_chat_member use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to remove a member from a chat.
type Command struct {
	ChatUuid   uuid.UUID
	UserUuid   uuid.UUID
	MemberUuid uuid.UUID
}

// NewHandler constructs a kick_chat_member handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle removes a member from a group chat.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()
	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	if command.MemberUuid == uuid.Nil {
		validationError.AddViolation("memberUuid", "Required")
	}

	if command.UserUuid != uuid.Nil && command.MemberUuid == command.UserUuid {
		validationError.AddViolation("memberUuid", "Must not match the authenticated user UUID")
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

	if chat.ChatType != domain.ChatTypeGroup {
		validationError := validation.NewError()
		validationError.AddViolation("chatUuid", "Members can only be removed from group chats")
		return validationError
	}

	currentMember, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if currentMember == nil {
		return ErrChatAccessDenied
	}

	if chat.CreatedByUserUuid != command.UserUuid {
		return ErrChatAccessDenied
	}

	targetMember, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.MemberUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.MemberUuid, command.ChatUuid, err)
	}

	if targetMember == nil {
		return ErrChatMemberNotFound
	}

	members, err := handler.chatMemberRepository.FindAllByChatUuid(ctx, command.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding members by chat uuid %q: %w", command.ChatUuid, err)
	}

	if len(members) <= 2 {
		validationError := validation.NewError()
		validationError.AddViolation("memberUuid", "Group chats must keep at least two members")
		return validationError
	}

	if err := handler.chatMemberRepository.DeleteByChatUuidAndUserUuid(ctx, command.ChatUuid, command.MemberUuid); err != nil {
		return fmt.Errorf("kicking member %q from chat %q: %w", command.MemberUuid, command.ChatUuid, err)
	}

	return nil
}
