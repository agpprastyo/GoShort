-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountLinks :one
SELECT COUNT(*) FROM short_links;

-- name: CountActiveLinks :one
SELECT COUNT(*) FROM short_links WHERE is_active = TRUE;

-- name: CountInactiveLinks :one
SELECT COUNT(*) FROM short_links WHERE is_active = FALSE;