CREATE TABLE profiles (
    user_uuid UUID PRIMARY KEY,
    avatar TEXT,
    name TEXT NOT NULL,
    nickname TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
