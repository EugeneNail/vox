CREATE TABLE chats (
    uuid UUID PRIMARY KEY,
    name TEXT NULL,
    is_direct BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    server_uuid UUID NULL
);
