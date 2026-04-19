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

-- name: MarkMessagesAsRead :exec
UPDATE messages
SET is_read = TRUE
WHERE chat_id = $1 AND sender_id != $2 AND is_read = FALSE;

-- name: GetUnreadMessages :many
SELECT users.nickname, COUNT(*) as count
FROM messages
JOIN chats ON chats.id = messages.chat_id
JOIN chat_members ON chat_members.chat_id = chats.id AND chat_members.user_id = $1
JOIN users ON users.id = messages.sender_id
WHERE messages.is_read = FALSE AND messages.sender_id != $1
GROUP BY users.nickname;