-- +goose Up
CREATE TABLE chat_members (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    PRIMARY KEY (chat_id, user_id)
);

-- +goose Down
DROP TABLE chat_members;