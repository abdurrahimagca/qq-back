-- name: InsertAuth :one
INSERT INTO auth (email, provider, provider_id)
VALUES (sqlc.arg(email), sqlc.arg(provider), sqlc.arg(provider_id))
RETURNING id;

-- name: InsertUser :one
INSERT INTO users (auth_id, username, display_name, avatar_key)
VALUES (sqlc.arg(auth_id), sqlc.arg(username), sqlc.narg(display_name), sqlc.narg(avatar_key))
RETURNING id;

-- name: InsertAuthOtpCode :one
INSERT INTO auth_otp_codes (auth_id, code)
VALUES (sqlc.arg(auth_id), sqlc.arg(code))
RETURNING id;

-- name: UpdateUser :one
UPDATE users
SET username = COALESCE(sqlc.narg(username), username), 
    display_name = COALESCE(sqlc.narg(display_name), display_name), 
    avatar_key = COALESCE(sqlc.narg(avatar_key), avatar_key), 
    privacy_level = COALESCE(sqlc.narg(privacy_level), privacy_level)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetUserIdAndEmailByOtpCode :one
SELECT users.id, auth.email , auth.id as auth_id
FROM users 
JOIN auth ON users.auth_id = auth.id 
JOIN auth_otp_codes ON auth.id = auth_otp_codes.auth_id 
WHERE auth_otp_codes.code = sqlc.arg(code) AND auth_otp_codes.expires_at > CURRENT_TIMESTAMP 
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE auth_id = (SELECT id FROM auth WHERE email = sqlc.arg(email)) LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = sqlc.arg(id) LIMIT 1;

-- name: DeleteOtpCodeEntryByAuthID :exec
DELETE FROM auth_otp_codes WHERE auth_id = sqlc.arg(auth_id);

-- name: DeleteOtpCodesByEmail :one
DELETE FROM auth_otp_codes WHERE auth_id = (SELECT id FROM auth WHERE email = sqlc.arg(email)) RETURNING COUNT(*);