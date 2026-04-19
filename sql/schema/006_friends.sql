-- +goose Up
CREATE TABLE friends (
    user_id UUID REFERENCES users(id),
    friend_id UUID REFERENCES users(id),
    request_status TEXT NOT NULL DEFAULT 'pending', -- pending, accepted
    PRIMARY KEY (user_id, friend_id)
);

-- +goose Down
DROP TABLE friends;