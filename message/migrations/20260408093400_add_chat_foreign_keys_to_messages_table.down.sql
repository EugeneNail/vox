ALTER TABLE messages
    DROP CONSTRAINT messages_chat_member_fkey;

ALTER TABLE messages
    DROP CONSTRAINT messages_chat_uuid_fkey;
