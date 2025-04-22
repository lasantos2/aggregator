-- +goose Up
CREATE TABLE users (
id uuid PRIMARY KEY,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
name TEXT UNIQUE NOT NULL
);

CREATE TABLE feeds (
id uuid PRIMARY KEY,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
name TEXT NOT NULL,
url TEXT UNIQUE NOT NULL,
user_id uuid UNIQUE NOT NULL,
FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE feed_follows (
id uuid PRIMARY KEY,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
feed_id uuid NOT NULL,
user_id uuid NOT NULL,
FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);


-- CREATE TABLE users (
-- id INTEGER PRIMARY KEY,
-- name TEXT NOT NULL,
-- age INTEGER NOT NULL,
-- username TEXT UNIQUE NOT NULL,
-- password TEXT NOT NULL,
-- is_admin BOOLEAN
-- );
-- +goose Down
DROP TABLE feed_follows;
DROP TABLE feeds;
DROP TABLE users;

