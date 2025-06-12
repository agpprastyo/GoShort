-- name: GetLinkStats :many
-- Returns all stats for a specific link owned by the user
SELECT
    ls.id, ls.link_id, ls.click_time, ls.ip_address,
    ls.user_agent, ls.referrer, ls.country, ls.device_type
FROM link_stats ls
JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.id = $1 AND sl.user_id = $2
ORDER BY ls.click_time DESC
LIMIT $3 OFFSET $4;

-- name: GetLinkStatsCount :one
-- Returns the total count of stats entries for a specific link
SELECT COUNT(*)
FROM link_stats ls
JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.id = $1 AND sl.user_id = $2;

-- name: GetLinkStatsGroupedByCountry :many
-- Returns stats grouped by country for a specific link
SELECT ls.country, COUNT(*) as clicks
FROM link_stats ls
JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.id = $1 AND sl.user_id = $2
GROUP BY ls.country
ORDER BY clicks DESC;

-- name: GetLinkStatsGroupedByDate :many
-- Returns stats grouped by date for a specific link
SELECT
    DATE(ls.click_time) as date,
    COUNT(*) as clicks
FROM link_stats ls
JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.id = $1 AND sl.user_id = $2
GROUP BY DATE(ls.click_time)
ORDER BY date DESC;

-- name: GetLinkStatsByDateRange :many
-- Returns all stats for a specific link within a date range
SELECT
    ls.id, ls.link_id, ls.click_time, ls.ip_address,
    ls.user_agent, ls.referrer, ls.country, ls.device_type
FROM link_stats ls
JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.id = $1
  AND sl.user_id = $2
  AND ls.click_time BETWEEN $3 AND $4
ORDER BY ls.click_time DESC;


-- name: CreateLinkStat :one
-- Records a click event when someone accesses a short link with custom UUID v7
INSERT INTO link_stats (
    id,
    link_id,
    ip_address,
    user_agent,
    referrer,
    country,
    device_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;