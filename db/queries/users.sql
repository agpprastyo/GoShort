-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE
  -- Search functionality - optional search across multiple fields
    (@search_text::text = '' OR
     username ILIKE '%' || @search_text || '%' OR
     email ILIKE '%' || @search_text || '%' OR
     first_name ILIKE '%' || @search_text || '%' OR
     last_name ILIKE '%' || @search_text || '%')
  -- Date range filtering - both dates optional
  AND (@start_date::timestamptz IS NULL OR created_at >= @start_date)
  AND (@end_date::timestamptz IS NULL OR created_at <= @end_date)
ORDER BY
    CASE
        WHEN @order_by::user_order_column = 'created_at' AND @ascending::bool = true THEN created_at
        END ASC,
    CASE
        WHEN @order_by::user_order_column = 'created_at' AND @ascending::bool = false THEN created_at
        END DESC,
    CASE
        WHEN @order_by::user_order_column = 'username' AND @ascending::bool = true THEN username
        END ASC,
    CASE
        WHEN @order_by::user_order_column = 'username' AND @ascending::bool = false THEN username
        END DESC,
    CASE
        WHEN @order_by::user_order_column = 'email' AND @ascending::bool = true THEN email
        END ASC,
    CASE
        WHEN @order_by::user_order_column = 'email' AND @ascending::bool = false THEN email
        END DESC,
    CASE
        WHEN @order_by::user_order_column = 'first_name' AND @ascending::bool = true THEN first_name
        END ASC,
    CASE
        WHEN @order_by::user_order_column = 'first_name' AND @ascending::bool = false THEN first_name
        END DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersByRole :many
SELECT * FROM users
WHERE role = @role::user_role
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (
  id, username, password_hash, email, first_name, last_name, role
) VALUES (
  $1, $2, $3, $4, $5, $6, @role::user_role
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
  username = COALESCE($2, username),
  password_hash = COALESCE($3, password_hash),
  email = COALESCE($4, email),
  first_name = COALESCE($5, first_name),
  last_name = COALESCE($6, last_name),
  role = COALESCE(@role::user_role, role)
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;