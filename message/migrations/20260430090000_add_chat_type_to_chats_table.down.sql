ALTER TABLE chats
    ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE chats
SET is_private = CASE
    WHEN chat_type = 0 THEN TRUE
    ELSE FALSE
END;

ALTER TABLE chats
    DROP COLUMN chat_type;
