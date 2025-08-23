-- name: CreateToken :one
-- CreateToken inserts a new token into the database.
INSERT INTO tokens (id, user_id, token_hash, type, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTokenByHash :one
-- GetTokenByHash retrieves a token and the associated user's active status.
-- This is useful for verifying a token and checking if the user's account is already active.
SELECT t.*, u.is_active as user_is_active
FROM tokens t
         JOIN users u ON t.user_id = u.id
WHERE t.token_hash = $1
LIMIT 1;

-- name: DeleteTokenByID :exec
-- DeleteTokenByID removes a specific token from the database by its ID.
-- This is typically used after a token has been successfully used.
DELETE FROM tokens
WHERE id = $1;

-- name: DeleteTokensByUserIDAndType :exec
-- DeleteTokensByUserIDAndType removes all tokens of a specific type for a given user.
-- This is useful for invalidating all existing password reset tokens when a new one is requested.
DELETE FROM tokens
WHERE user_id = $1 AND type = $2;


-- name: IncrementTokenAttempts :exec
-- IncrementTokenAttempts increases the attempt count for a specific token by one.
UPDATE tokens
SET attempts = attempts + 1
WHERE id = $1;

-- name: GetLatestTokenByUserIDAndType :one
-- GetLatestTokenByUserIDAndType retrieves the most recent token for a user of a specific type.
SELECT * FROM tokens
WHERE user_id = $1 AND type = $2
ORDER BY created_at DESC
LIMIT 1;
