CREATE TABLE attachments (
    uuid UUID PRIMARY KEY,
    message_uuid UUID NOT NULL REFERENCES messages(uuid),
    name TEXT NOT NULL
);

CREATE INDEX attachments_message_uuid_idx
    ON attachments (message_uuid);
