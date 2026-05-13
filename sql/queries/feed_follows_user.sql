-- name: GetFeedFollowsForUser :many
SELECT *, feed.name AS feed_name, users.name AS user_name FROM feed_follows
INNER JOIN feed ON feed_follows.feed_id = feed.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;
