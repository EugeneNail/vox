ALTER TABLE chats
ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE chats
SET is_private = TRUE
WHERE uuid IN (
    SELECT chat_uuid
    FROM chat_members
    GROUP BY chat_uuid
    HAVING COUNT(*) = 2
);
