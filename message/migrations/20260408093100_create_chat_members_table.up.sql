CREATE TABLE chat_members (
    chat_uuid UUID NOT NULL REFERENCES chats(uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL,
    role TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (chat_uuid, user_uuid)
);

CREATE INDEX chat_members_user_uuid_idx
    ON chat_members (user_uuid);
