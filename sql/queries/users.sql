-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = $1;

-- name: Reset :exec
DELETE FROM users;
DELETE FROM feeds;
DELETE FROM feed_follows;

-- name: GetUsers :many
SELECT * FROM users;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedUser :one
SELECT users.name FROM users JOIN feeds ON users.id = $1;

-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
INSERT INTO feed_follows (id, created_at, updated_at, feed_id, user_id)
VALUES ( $1,$2,$3,$4,$5)
RETURNING *
)
SELECT 
inserted_feed_follow.*,
feeds.name AS feed_name,
users.name AS user_name
FROM inserted_feed_follow 
INNER JOIN users ON inserted_feed_follow.user_id = users.id
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT *, feeds.name AS feed_name, users.name AS user_name 
FROM feed_follows
INNER JOIN users ON feed_follows.user_id = users.id
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE users.id = $1;

-- name: DeleteFeed :exec
DELETE FROM feed_follows 
USING users, feeds
WHERE feed_follows.user_id = users.id
AND feed_follows.feed_id = feeds.id
AND users.id = $1
AND feeds.url = $2;


-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $2
WHERE id = $3;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds ORDER BY last_fetched_at ASC;


