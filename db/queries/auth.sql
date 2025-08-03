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

-- name: SearchUserByAuthID :one
SELECT * FROM users WHERE auth_id = sqlc.arg(auth_id) LIMIT 1;

-- name: SearchAuthByEmail :one
SELECT * FROM auth WHERE email = sqlc.arg(email) LIMIT 1;


-- name: DeleteOtpCodeById :exec
DELETE FROM auth_otp_codes WHERE id = sqlc.arg(id);

-- name: GetUserIdAndEmailByOtpCode :one
SELECT users.id, auth.email 
FROM users 
JOIN auth ON users.auth_id = auth.id 
JOIN auth_otp_codes ON auth.id = auth_otp_codes.auth_id 
WHERE auth_otp_codes.code = sqlc.arg(code) AND auth_otp_codes.expires_at > CURRENT_TIMESTAMP 
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE auth_id = (SELECT id FROM auth WHERE email = sqlc.arg(email)) LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = sqlc.arg(id) LIMIT 1;