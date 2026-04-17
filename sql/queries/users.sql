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

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
    AND revoked_at IS NULL
    AND expires_at > NOW();
