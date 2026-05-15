ALTER TABLE chat_members
    ADD COLUMN last_seen_revision BIGINT NOT NULL DEFAULT 0;
