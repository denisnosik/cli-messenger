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
