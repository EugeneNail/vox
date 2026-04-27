package events

const (
	// UserOpenedChatStream is the Redis stream for chat view open events.
	UserOpenedChatStream = "UserOpenedChat"
	// MessageCreatedStream is the Redis stream for created messages.
	MessageCreatedStream = "message.created"
	// MessageEditedStream is the Redis stream for edited messages.
	MessageEditedStream = "message.edited"
	// MessageDeletedStream is the Redis stream for deleted messages.
	MessageDeletedStream = "message.deleted"
)
