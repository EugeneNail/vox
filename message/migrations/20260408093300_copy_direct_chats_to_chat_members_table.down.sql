DELETE FROM chat_members
WHERE chat_uuid IN (
    SELECT uuid
    FROM direct_chats
);
