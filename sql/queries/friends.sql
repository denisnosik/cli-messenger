-- name: CreateFriendRequest :one
INSERT INTO friends (
    user_id,
    friend_id
) VALUES (
    $1,
    $2
)
RETURNING *;

-- name: AcceptFriendRequest :exec
UPDATE friends
SET request_status = 'accepted'
WHERE user_id = $1 AND friend_id = $2;

-- name: CreateFriendship :exec
INSERT INTO friends (
    user_id, 
    friend_id, 
    request_status
) VALUES (
    $1, 
    $2, 
    'accepted'
);

-- name: DeleteFriendRequest :exec
DELETE FROM friends
WHERE user_id = $1 AND friend_id = $2;

-- name: GetFriendshipStatus :one
SELECT user_id, request_status FROM friends
WHERE (user_id = $1 AND friend_id = $2)
    OR (user_id = $2 AND friend_id = $1)
LIMIT 1;