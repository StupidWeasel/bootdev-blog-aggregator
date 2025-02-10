-- name: CreatePost :execrows
INSERT INTO posts (title, url, description, published_at, feed_id)
SELECT unnest(@titles::text[]),
       unnest(@urls::text[]),
       unnest(@descriptions::text[]),
       unnest(@published_ats::timestamp with time zone[]),
       unnest(@feed_ids::bigint[])
ON CONFLICT (url) DO UPDATE
SET
    title = EXCLUDED.title,
    description = EXCLUDED.description
WHERE   posts.title IS DISTINCT FROM EXCLUDED.title 
        OR posts.description IS DISTINCT FROM EXCLUDED.description;

-- name: GetPostsForUser_Forward :many
SELECT p.* 
FROM posts p
JOIN feed_follows f ON p.feed_id = f.feed_id
WHERE f.user_id = $1
  AND (
        NOT sqlc.arg(has_cursor)::boolean
        OR p.published_at < sqlc.arg(cursor_time)
        OR (p.published_at = sqlc.arg(cursor_time) AND p.id < sqlc.arg(cursor_id))
      )
ORDER BY p.published_at DESC, p.id DESC
LIMIT $2;

-- name: GetPostsForUser_Backward :many
SELECT p.* 
FROM posts p
JOIN feed_follows f ON p.feed_id = f.feed_id
WHERE f.user_id = $1
  AND (
        NOT sqlc.arg(has_cursor)::boolean
        OR p.published_at > sqlc.arg(cursor_time)
        OR (p.published_at = sqlc.arg(cursor_time) AND p.id > sqlc.arg(cursor_id))
      )
ORDER BY p.published_at DESC, p.id DESC
LIMIT $2;