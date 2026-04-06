CREATE TABLE direct_chats (
    uuid UUID PRIMARY KEY,
    first_member_uuid UUID NOT NULL,
    second_member_uuid UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX direct_chats_first_member_uuid_idx
    ON direct_chats (first_member_uuid);

CREATE INDEX direct_chats_second_member_uuid_idx
    ON direct_chats (second_member_uuid);
