-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = @username AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUserPassword :exec
UPDATE users SET password = @password, updated_at = NOW() WHERE id = @id;

-- name: UpdateUserUsername :exec
UPDATE users SET username = @username, updated_at = NOW() WHERE id = @id;