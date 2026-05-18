package create_message

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

var attachmentNamePattern = regexp.MustCompile(`^.+\.[a-z]{2,5}$`)

// Handler creates messages through the create_message use-case.
type Handler struct {
	messageRepository                domain.MessageRepository
	chatRepository                   domain.ChatRepository
	chatMemberRepository             domain.ChatMemberRepository
	lastSeenRevisionUpdatedPublisher events.LastSeenRevisionUpdatedPublisher
	messageCreatedPublisher          events.MessageCreatedPublisher
}

// Command contains the input required to create a message.
type Command struct {
	ChatUuid    uuid.UUID
	UserUuid    uuid.UUID
	Text        string
	Attachments []string
}

// NewHandler constructs a create_message handler with its dependencies.
func NewHandler(
	messageRepository domain.MessageRepository,
	chatRepository domain.ChatRepository,
	chatMemberRepository domain.ChatMemberRepository,
	lastSeenRevisionUpdatedPublisher events.LastSeenRevisionUpdatedPublisher,
	messageCreatedPublisher events.MessageCreatedPublisher,
) *Handler {
	return &Handler{
		messageRepository:                messageRepository,
		chatRepository:                   chatRepository,
		chatMemberRepository:             chatMemberRepository,
		lastSeenRevisionUpdatedPublisher: lastSeenRevisionUpdatedPublisher,
		messageCreatedPublisher:          messageCreatedPublisher,
	}
}

// Handle validates input and creates a new message.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	messageUuid := uuid.UUID(uuidv7.New())
	command.Text = strings.TrimSpace(command.Text)
	attachments, err := buildAttachments(messageUuid, command.Attachments)
	if err != nil {
		return uuid.Nil, fmt.Errorf("building attachments for message %q: %w", messageUuid, err)
	}

	validator := validation.NewValidator(map[string]any{
		"chatUuid":           command.ChatUuid,
		"userUuid":           command.UserUuid,
		"text":               command.Text,
		"attachments.length": len(command.Attachments),
	}, map[string][]rules.Rule{
		"chatUuid":           {rules.Required()},
		"userUuid":           {rules.Required()},
		"text":               buildTextRules(command.Text, len(attachments) > 0),
		"attachments.length": {rules.Max(10)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return uuid.Nil, validationError
		}

		return uuid.Nil, fmt.Errorf("validating create message command: %w", err)
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, command.ChatUuid)
	if err != nil {
		return uuid.Nil, fmt.Errorf("finding chat by uuid %q: %w", command.ChatUuid, err)
	}

	if chat == nil {
		return uuid.Nil, ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return uuid.Nil, fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if member == nil {
		return uuid.Nil, ErrChatAccessDenied
	}

	now := time.Now().UTC()
	// TODO: handle concurrent message creation because chat revision is a shared resource.
	chat.Revision++
	message := domain.Message{
		Uuid:        messageUuid,
		ChatUuid:    command.ChatUuid,
		UserUuid:    command.UserUuid,
		Revision:    chat.Revision,
		Text:        command.Text,
		Attachments: attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := handler.messageRepository.Create(ctx, message); err != nil {
		return uuid.Nil, fmt.Errorf("creating message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	// TODO: make chat revision persistence and message publication transactional if ordering ever matters here.
	if err := handler.chatRepository.SetRevision(ctx, chat.Uuid, chat.Revision); err != nil {
		return uuid.Nil, fmt.Errorf("setting revision %d for chat %q: %w", chat.Revision, chat.Uuid, err)
	}

	if err := handler.chatMemberRepository.SetLastSeenRevision(ctx, chat.Uuid, command.UserUuid, chat.Revision); err != nil {
		return uuid.Nil, fmt.Errorf(
			"setting last seen revision %d for message author %q in chat %q: %w",
			chat.Revision,
			command.UserUuid,
			chat.Uuid,
			err,
		)
	}

	if err := handler.lastSeenRevisionUpdatedPublisher.Publish(ctx, events.LastSeenRevisionUpdated{
		ChatUuid:         chat.Uuid,
		UserUuid:         command.UserUuid,
		LastSeenRevision: chat.Revision,
	}); err != nil {
		log.Printf(
			"publishing last seen revision updated event for message author %q in chat %q revision %d: %v",
			command.UserUuid,
			chat.Uuid,
			chat.Revision,
			err,
		)
	}

	members, err := handler.chatMemberRepository.FindAllByChatUuid(ctx, chat.Uuid)
	recipientUuids := make([]uuid.UUID, 0)
	if err != nil {
		log.Printf("finding chat members for message created event in chat %q: %v", chat.Uuid, err)
	} else {
		recipientUuids = make([]uuid.UUID, 0, len(members))
		for _, currentMember := range members {
			recipientUuids = append(recipientUuids, currentMember.UserUuid)
		}
	}

	if err := handler.messageCreatedPublisher.Publish(ctx, events.MessageCreated{
		MessageUuid:    message.Uuid,
		ChatUuid:       message.ChatUuid,
		UserUuid:       message.UserUuid,
		RecipientUuids: recipientUuids,
		Revision:       message.Revision,
		Text:           message.Text,
		Attachments:    toEventAttachments(message.Attachments),
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,
	}); err != nil {
		log.Printf(
			"publishing message created event for message %q in chat %q to %d recipients: %v",
			message.Uuid,
			message.ChatUuid,
			len(recipientUuids),
			err,
		)
	}

	return message.Uuid, nil
}

func buildTextRules(text string, allowEmpty bool) []rules.Rule {
	if allowEmpty && text == "" {
		return nil
	}

	return []rules.Rule{
		rules.Required(),
		rules.Regex(domain.MessageTextPattern),
		rules.Min(1),
		rules.Max(2000),
	}
}

func buildAttachments(messageUuid uuid.UUID, names []string) ([]domain.Attachment, error) {
	if len(names) == 0 {
		return nil, nil
	}

	validationError := validation.NewError()
	attachments := make([]domain.Attachment, 0, len(names))

	for index, rawName := range names {
		name := strings.TrimSpace(rawName)
		if !attachmentNamePattern.MatchString(name) {
			validationError.AddViolation(fmt.Sprintf("attachments.%d", index), "must end with a five-letter extension")
			continue
		}

		attachments = append(attachments, domain.Attachment{
			Uuid:        uuid.UUID(uuidv7.New()),
			Name:        name,
			MessageUuid: messageUuid,
		})
	}

	if len(validationError.Violations()) > 0 {
		return nil, validationError
	}

	return attachments, nil
}

func toEventAttachments(attachments []domain.Attachment) []events.Attachment {
	result := make([]events.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		result = append(result, events.Attachment{
			Uuid: attachment.Uuid,
			Name: attachment.Name,
		})
	}

	return result
}
