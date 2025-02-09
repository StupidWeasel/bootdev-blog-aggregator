-- name: CreateFeedFollow :one
INSERT INTO feed_follows (user_id, feed_id)
VALUES (
    $1,
    $2
)
RETURNING feed_follows.*,
(SELECT name FROM users WHERE users.id = feed_follows.user_id) AS user_name,
(SELECT name FROM feeds WHERE feeds.id = feed_follows.feed_id) AS feed_name;

-- name: GetFeedFollows :many
SELECT u.name as user_name, f.name as feed_name, f.url as url
FROM feed_follows ff
JOIN users as u ON u.id = ff.user_id
JOIN feeds as f ON f.id = ff.feed_id
WHERE ff.user_id = $1;