INSERT INTO chat_members (chat_uuid, user_uuid, role, joined_at)
SELECT uuid, first_member_uuid, 'member', created_at
FROM direct_chats
ON CONFLICT (chat_uuid, user_uuid) DO NOTHING;

INSERT INTO chat_members (chat_uuid, user_uuid, role, joined_at)
SELECT uuid, second_member_uuid, 'member', created_at
FROM direct_chats
ON CONFLICT (chat_uuid, user_uuid) DO NOTHING;
