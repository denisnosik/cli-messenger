-- name: CreateMessage :one
INSERT INTO messages (
    id,
    chat_id,
    sender_id,
    content,
    created_at
) VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    NOW()
)
RETURNING *;

-- name: GetMessagesByChat :many
SELECT messages.content, users.nickname, messages.created_at 
FROM messages
JOIN users ON users.id = messages.sender_id
WHERE messages.chat_id = $1
ORDER BY messages.created_at
LIMIT $2;

