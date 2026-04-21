-- +goose Up
CREATE TABLE messages (
    id UUID PRIMARY KEY NOT NULL,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE messages;