-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, last_fetched_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: DeleteFeeds :exec
DELETE FROM feeds;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: MarkFeedAsFetched :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $2
WHERE id = $3;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: GetFeeds :many
SELECT * FROM feeds;