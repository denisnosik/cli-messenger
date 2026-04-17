-- name: CreateChat :one
INSERT INTO chats (
    id,
    created_at
) VALUES (
    gen_random_uuid(),
    NOW()
)
RETURNING *;

-- name: GetChatByID :one
SELECT * FROM chats
WHERE id = $1
LIMIT 1;

-- name: GetChatByTwoUsers :one
SELECT chats.id FROM chats
JOIN chat_members cm1 ON cm1.chat_id = chats.id AND cm1.user_id = $1
JOIN chat_members cm2 ON cm2.chat_id = chats.id AND cm2.user_id = $2;
