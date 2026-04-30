package add_chat_members

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler adds members to a group chat through the add_chat_members use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to add members to a chat.
type Command struct {
	ChatUuid    uuid.UUID
	UserUuid    uuid.UUID
	MemberUuids []uuid.UUID
}

// NewHandler constructs an add_chat_members handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle adds new members to a group chat.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()
	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	seenMemberUuids := make(map[uuid.UUID]struct{}, len(command.MemberUuids))
	for index, memberUuid := range command.MemberUuids {
		if memberUuid == uuid.Nil {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain empty UUID values")
			continue
		}

		if memberUuid == command.UserUuid {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain the authenticated user UUID")
			continue
		}

		if _, exists := seenMemberUuids[memberUuid]; exists {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain duplicate UUID values")
			continue
		}

		seenMemberUuids[memberUuid] = struct{}{}
	}

	if len(command.MemberUuids) == 0 {
		validationError.AddViolation("memberUuids", "At least one member UUID is required")
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
		validationError.AddViolation("chatUuid", "Members can only be added to group chats")
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

	currentMembers, err := handler.chatMemberRepository.FindAllByChatUuid(ctx, command.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding members by chat uuid %q: %w", command.ChatUuid, err)
	}

	existingMemberUuids := make(map[uuid.UUID]struct{}, len(currentMembers))
	for _, member := range currentMembers {
		existingMemberUuids[member.UserUuid] = struct{}{}
	}

	validationError = validation.NewError()
	members := make([]domain.ChatMember, 0, len(command.MemberUuids))
	now := time.Now().UTC()
	for index, memberUuid := range command.MemberUuids {
		if _, exists := existingMemberUuids[memberUuid]; exists {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "User is already a chat member")
			continue
		}

		members = append(members, domain.ChatMember{
			ChatUuid: command.ChatUuid,
			UserUuid: memberUuid,
			Role:     domain.ChatMemberRoleMember,
			JoinedAt: now,
		})
	}

	if len(validationError.Violations()) > 0 {
		return validationError
	}

	if err := handler.chatMemberRepository.CreateMany(ctx, members); err != nil {
		return fmt.Errorf("adding members to chat %q: %w", command.ChatUuid, err)
	}

	return nil
}
