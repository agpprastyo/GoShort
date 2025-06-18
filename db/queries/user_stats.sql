-- name: GetUserLinkStats :one
SELECT
  COUNT(DISTINCT sl.id) AS total_links,
  COUNT(DISTINCT sl.id) FILTER (WHERE sl.is_active) AS active_links,
  COUNT(DISTINCT sl.id) FILTER (WHERE NOT sl.is_active) AS inactive_links,
  COALESCE(SUM(ls.clicks), 0) AS total_clicks
FROM short_links sl
LEFT JOIN (
  SELECT link_id, COUNT(*) AS clicks
  FROM link_stats
  GROUP BY link_id
) ls ON sl.id = ls.link_id
WHERE sl.user_id = $1;

-- name: GetLinkClickStatsByDateRange :many
SELECT
  DATE_TRUNC($2, click_time) AS period,
  COUNT(*) AS clicks
FROM link_stats
WHERE link_id = $1
  AND click_time BETWEEN $3 AND $4
GROUP BY period
ORDER BY period DESC;