-- +goose Up
CREATE TABLE chats (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE chats;