ALTER TABLE chats
    ADD COLUMN chat_type SMALLINT NOT NULL DEFAULT 0;

UPDATE chats
SET chat_type = CASE
    WHEN is_private = TRUE THEN 0
    ELSE 1
END;

ALTER TABLE chats
    DROP COLUMN is_private;
