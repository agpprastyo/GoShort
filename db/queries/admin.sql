-- name: AdminListShortLinks :many
SELECT * FROM short_links
    WHERE  TRUE
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



-- name: AdminGetShortLinkByID :one
SELECT * FROM short_links
WHERE id = @id::uuid;


-- name: AdminGetShortLinksByUserID :many
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

-- name: AdminToggleShortLinkStatus :exec
UPDATE short_links
SET is_active = NOT is_active
WHERE id = $1;