-- name: AddFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeeds :many
SELECT f.name, f.url, f.user_id, u.name as username
FROM feeds f
JOIN users u
ON f.user_id = u.id;

-- name: GetFeed :one
SELECT id, name, url
FROM feeds
WHERE url = $1
LIMIT 1;