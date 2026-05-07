-- name: GetFeeds :many
SELECT f.name, f.url, u.name as user_name
FROM feed as f
JOIN users as u ON f.user_id = u.id
ORDER BY f.name ASC;
