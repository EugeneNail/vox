DELETE FROM chats
WHERE uuid IN (
    SELECT uuid
    FROM direct_chats
);
