-- name: GetShortLink :one
SELECT * FROM short_links
WHERE id = $1 LIMIT 1;

-- name: GetShortLinkByCode :one
SELECT * FROM short_links
WHERE short_code = $1 LIMIT 1;

-- name: GetActiveShortLinkByCode :one
SELECT * FROM short_links
WHERE short_code = $1
AND is_active = true
AND (expired_at IS NULL OR expired_at > NOW())
AND (click_limit IS NULL OR click_limit > 0)
LIMIT 1;

-- name: ListShortLinks :many
SELECT * FROM short_links
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;



-- name: ListUserShortLinks :many
SELECT * FROM short_links
WHERE user_id = $1
  -- Search functionality - search by title (handles NULL titles)
  AND (@search_text::text = '' OR (title IS NOT NULL AND title ILIKE '%' || @search_text || '%'))
  -- Date range filtering for created_at
  AND (@start_date::timestamptz IS NULL OR created_at >= @start_date)
  AND (@end_date::timestamptz IS NULL OR created_at <= @end_date)
ORDER BY
    CASE
        WHEN @order_by::shortlink_order_column = 'title' AND @ascending::bool = true THEN title
END ASC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'title' AND @ascending::bool = false THEN title
END DESC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'is_active' AND @ascending::bool = true THEN is_active::int
END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'is_active' AND @ascending::bool = false THEN is_active::int
END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'created_at' AND @ascending::bool = true THEN created_at
END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'created_at' AND @ascending::bool = false THEN created_at
END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'updated_at' AND @ascending::bool = true THEN updated_at
END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'updated_at' AND @ascending::bool = false THEN updated_at
END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'expired_at' AND @ascending::bool = true THEN expired_at
END ASC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'expired_at' AND @ascending::bool = false THEN expired_at
END DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: CountUserShortLinks :one
SELECT COUNT(*)
FROM short_links
WHERE user_id = $1
  -- Search functionality - search by title (handles NULL titles)
  AND (@search_text::text = '' OR (title IS NOT NULL AND title ILIKE '%' || @search_text || '%'))
  -- Date range filtering for created_at
  AND (@start_date::timestamptz IS NULL OR created_at >= @start_date)
  AND (@end_date::timestamptz IS NULL OR created_at <= @end_date);

-- name: ListUserShortLinksWithCountClick :many
SELECT sl.*,
       COALESCE(ls.click_count, 0) AS total_clicks
FROM short_links sl
         LEFT JOIN (
    SELECT link_id, COUNT(*) AS click_count
    FROM link_stats
    GROUP BY link_id
) ls ON sl.id = ls.link_id
WHERE sl.user_id = $1
  -- Search functionality - search by title (handles NULL titles)
  AND (@search_text::text = '' OR (sl.title IS NOT NULL AND sl.title ILIKE '%' || @search_text || '%'))
  -- Date range filtering for created_at
  AND (@start_date::timestamptz IS NULL OR sl.created_at >= @start_date)
  AND (@end_date::timestamptz IS NULL OR sl.created_at <= @end_date)
GROUP BY sl.id, ls.click_count
ORDER BY
    CASE
        WHEN @order_by::shortlink_order_column = 'title' AND @ascending::bool = true THEN sl.title
        END ASC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'title' AND @ascending::bool = false THEN sl.title
        END DESC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'is_active' AND @ascending::bool = true THEN sl.is_active::int
        END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'is_active' AND @ascending::bool = false THEN sl.is_active::int
        END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'created_at' AND @ascending::bool = true THEN sl.created_at
        END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'created_at' AND @ascending::bool = false THEN sl.created_at
        END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'updated_at' AND @ascending::bool = true THEN sl.updated_at
        END ASC,
    CASE
        WHEN @order_by::shortlink_order_column = 'updated_at' AND @ascending::bool = false THEN sl.updated_at
        END DESC,
    CASE
        WHEN @order_by::shortlink_order_column = 'expired_at' AND @ascending::bool = true THEN sl.expired_at
        END ASC NULLS LAST,
    CASE
        WHEN @order_by::shortlink_order_column = 'expired_at' AND @ascending::bool = false THEN sl.expired_at
        END DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: CreateShortLink :one
INSERT INTO short_links (
  id, user_id, original_url, short_code, title, is_active, click_limit, expired_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateShortLink :one
UPDATE short_links
SET
  original_url = COALESCE($2, original_url),
  short_code = COALESCE($3, short_code),
  title = COALESCE($4, title),
  is_active = COALESCE($5, is_active),
  click_limit = COALESCE($6, click_limit),
  expired_at = COALESCE($7, expired_at)
WHERE id = $1
RETURNING *;

-- name: DecrementClickLimit :one
UPDATE short_links
SET click_limit = click_limit - 1
WHERE id = $1 AND click_limit > 0
RETURNING *;

-- name: DeactivateShortLink :one
UPDATE short_links
SET is_active = false
WHERE id = $1
RETURNING *;

-- name: DeleteUserShortLink :exec
DELETE FROM short_links
WHERE id = $1;

-- name: CheckShortCodeExists :one
SELECT EXISTS(
  SELECT 1 FROM short_links
  WHERE short_code = $1
) AS exists;

-- name: ToggleShortLinkStatus :one
UPDATE short_links
SET is_active = NOT is_active
WHERE id = $1
RETURNING *;

