ALTER TABLE messages
    ADD CONSTRAINT messages_chat_member_fkey
    FOREIGN KEY (chat_uuid, user_uuid) REFERENCES chat_members(chat_uuid, user_uuid);
