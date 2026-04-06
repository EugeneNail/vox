CREATE TABLE messages (
    uuid UUID PRIMARY KEY,
    chat_uuid UUID NOT NULL,
    user_uuid UUID NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX messages_chat_uuid_created_at_idx
    ON messages (chat_uuid, created_at DESC);
