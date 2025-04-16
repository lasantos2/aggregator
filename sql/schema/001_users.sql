-- +goose Up
CREATE TABLE users (
id uuid PRIMARY KEY,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
name TEXT UNIQUE NOT NULL
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
DROP TABLE users;