-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    nickname TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    hashed_password TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;