-- name: GetNextFeedToFetch :one
SELECT * FROM feed
ORDER BY last_fetched_at ASC, last_fetched_at NULLS FIRST;
