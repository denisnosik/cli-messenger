-- name: CreateChatMember :one
INSERT INTO chat_members (
    chat_id,
    user_id
) VALUES (
    $1,
    $2
)
RETURNING *;