CREATE TABLE chat_members (
    chat_uuid UUID NOT NULL,
    member_uuid UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (chat_uuid, member_uuid)
);

CREATE INDEX chat_members_member_uuid_idx
    ON chat_members (member_uuid);
