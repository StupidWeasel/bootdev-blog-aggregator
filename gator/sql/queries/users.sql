-- name: CreateUser :one
INSERT INTO users (name)
VALUES (
    $1
)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE name = $1
LIMIT 1;

-- name: GetUsers :many
SELECT name
FROM users;

-- name: ResetUsers :exec
TRUNCATE posts, feed_follows, feeds, users;
