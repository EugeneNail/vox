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

INSERT INTO direct_chats (uuid, first_member_uuid, second_member_uuid, created_at, updated_at)
SELECT
    c.uuid,
    MIN(cm.user_uuid::TEXT)::UUID,
    MAX(cm.user_uuid::TEXT)::UUID,
    c.created_at,
    c.updated_at
FROM chats c
INNER JOIN chat_members cm ON cm.chat_uuid = c.uuid
GROUP BY c.uuid, c.created_at, c.updated_at
HAVING COUNT(*) = 2;
