CREATE TABLE chats (
    uuid UUID PRIMARY KEY,
    name TEXT,
    avatar TEXT,
    created_by_user_uuid UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
