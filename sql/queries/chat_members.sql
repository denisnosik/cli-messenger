-- name: CreateChatMember :one
INSERT INTO chat_members (
    chat_id,
    user_id
) VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetChatMember :one
SELECT * FROM chat_members
WHERE chat_id = $1 AND user_id = $2;