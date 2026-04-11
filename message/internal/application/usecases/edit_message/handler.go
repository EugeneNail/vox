package edit_message

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/EugeneNail/vox/message/internal/domain/events"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

var ErrMessageNotFound = errors.New("message not found")
var ErrMessageAccessDenied = errors.New("message access denied")
var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

var attachmentNamePattern = regexp.MustCompile(`^.+\.[a-z]{2,5}$`)

// Handler edits messages through the edit_message use-case.
type Handler struct {
	messageRepository      domain.MessageRepository
	chatRepository         domain.ChatRepository
	chatMemberRepository   domain.ChatMemberRepository
	messageEditedPublisher events.MessageEditedPublisher
}

// Command contains the input required to edit a message.
type Command struct {
	MessageUuid uuid.UUID
	UserUuid    uuid.UUID
	Text        string
	Attachments []string
}

// NewHandler constructs an edit_message handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository, messageEditedPublisher events.MessageEditedPublisher) *Handler {
	return &Handler{
		messageRepository:      messageRepository,
		chatRepository:         chatRepository,
		chatMemberRepository:   chatMemberRepository,
		messageEditedPublisher: messageEditedPublisher,
	}
}

// Handle validates input and updates message text.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	text := strings.TrimSpace(command.Text)
	attachments, err := buildAttachments(command.MessageUuid, command.Attachments)
	if err != nil {
		return fmt.Errorf("building attachments for message %q: %w", command.MessageUuid, err)
	}

	validator := validation.NewValidator(map[string]any{
		"messageUuid": command.MessageUuid,
		"userUuid":    command.UserUuid,
		"text":        text,
		"attachments.length": len(command.Attachments),
	}, map[string][]rules.Rule{
		"messageUuid": {rules.Required()},
		"userUuid":    {rules.Required()},
		"text":        buildTextRules(text, len(attachments) > 0),
		"attachments.length": {rules.Max(10)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return validationError
		}

		return fmt.Errorf("validating edit message command: %w", err)
	}

	message, err := handler.messageRepository.FindByUuid(ctx, command.MessageUuid)
	if err != nil {
		return fmt.Errorf("finding message by uuid %q: %w", command.MessageUuid, err)
	}

	if message == nil {
		return ErrMessageNotFound
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, message.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", message.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, message.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, message.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	if message.UserUuid != command.UserUuid {
		return ErrMessageAccessDenied
	}

	message.Text = text
	message.Attachments = attachments
	message.UpdatedAt = time.Now().UTC()

	if err := handler.messageRepository.Update(ctx, *message); err != nil {
		return fmt.Errorf("updating message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	if err := handler.messageEditedPublisher.Publish(ctx, events.MessageEdited{
		MessageUuid: message.Uuid,
		ChatUuid:    message.ChatUuid,
		UserUuid:    message.UserUuid,
		Text:        message.Text,
		Attachments: toEventAttachments(message.Attachments),
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}); err != nil {
		log.Printf("publishing message edited event for message %q: %v", message.Uuid, err)
	}

	return nil
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

// TODO
// Refactor the websocket package by extracting a part of it into the transport layer
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
