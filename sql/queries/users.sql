-- name: CreateUser :one
INSERT INTO users (
    id,
    nickname,
    created_at,
    updated_at,
    hashed_password
) VALUES (
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    $2
)
RETURNING *;

-- name: GetUserByNickname :one
SELECT * FROM users
WHERE nickname = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;