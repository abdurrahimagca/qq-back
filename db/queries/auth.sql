-- name: InsertAuth :one
INSERT INTO auth (email, provider, provider_id)
VALUES (sqlc.arg(email), sqlc.arg(provider), sqlc.arg(provider_id))
RETURNING id;

-- name: InsertUser :one
INSERT INTO users (auth_id, username, display_name, avatar_url)
VALUES (sqlc.arg(auth_id), sqlc.arg(username), sqlc.narg(display_name), sqlc.narg(avatar_url))
RETURNING id;

-- name: InsertAuthOtpCode :one
INSERT INTO auth_otp_codes (auth_id, code)
VALUES (sqlc.arg(auth_id), sqlc.arg(code))
RETURNING id;

-- name: UpdateUser :one
UPDATE users
SET username = sqlc.narg(username), display_name = sqlc.narg(display_name), avatar_url = sqlc.narg(avatar_url)
WHERE id = sqlc.arg(id)
RETURNING id;