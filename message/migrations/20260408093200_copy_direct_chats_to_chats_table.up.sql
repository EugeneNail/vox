INSERT INTO chats (uuid, name, avatar, created_by_user_uuid, created_at, updated_at)
SELECT uuid, NULL, NULL, first_member_uuid, created_at, updated_at
FROM direct_chats;
