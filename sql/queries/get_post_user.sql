-- name: GetPostUser :many
SELECT * FROM users
WHERE id = $1
ORDER BY created_at DESC
LIMIT $2;
