-- name: GetFeedUrl :one
SELECT * FROM feed WHERE url = $1 LIMIT 1;
