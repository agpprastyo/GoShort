-- name: GetUserDashboardStats :one
-- Mengambil statistik ringkas untuk dashboard pengguna: jumlah total link dan jumlah total klik.
SELECT
    (SELECT count(*) FROM short_links WHERE user_id = sqlc.arg(user_id))::int AS total_links,
    (SELECT count(ls.id)
     FROM link_stats ls
              JOIN short_links sl ON ls.link_id = sl.id
     WHERE sl.user_id = sqlc.arg(user_id)
    )::int AS total_clicks;

-- name: GetUserLinksWithStats :many
-- Mengambil daftar link milik pengguna beserta jumlah klik untuk setiap link, dengan paginasi.
-- Menggunakan LEFT JOIN untuk memastikan link yang belum pernah diklik (0 klik) tetap muncul.
SELECT
    sl.id,
    sl.short_code,
    sl.original_url,
    sl.title,
    sl.created_at,
    count(ls.id)::int as click_count
FROM short_links sl
         LEFT JOIN link_stats ls ON sl.id = ls.link_id
WHERE sl.user_id = $1
GROUP BY sl.id
ORDER BY sl.created_at DESC
LIMIT $2
    OFFSET $3;

-- name: GetUserClicksByCountry :many
-- Mengelompokkan jumlah klik berdasarkan negara untuk semua link milik pengguna.
-- Berguna untuk membuat diagram statistik geografis.
SELECT
    ls.country,
    count(ls.id)::int as clicks
FROM link_stats ls
         JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.user_id = $1 AND ls.country IS NOT NULL
GROUP BY ls.country
ORDER BY clicks DESC;

-- name: GetUserClicksByReferrer :many
-- Mengelompokkan jumlah klik berdasarkan sumber trafik (referrer) untuk semua link milik pengguna.
-- Dibatasi dengan LIMIT untuk mengambil N sumber teratas.
SELECT
    ls.referrer,
    count(ls.id)::int as clicks
FROM link_stats ls
         JOIN short_links sl ON ls.link_id = sl.id
WHERE sl.user_id = $1 AND ls.referrer IS NOT NULL AND ls.referrer != ''
GROUP BY ls.referrer
ORDER BY clicks DESC
LIMIT $2;


-- name: GetUserClickTimeline :many
-- Mengambil data time-series jumlah klik per hari untuk pengguna tertentu dalam rentang waktu.
-- Berguna untuk membuat grafik tren klik dari waktu ke waktu.
SELECT
    date_trunc('day', ls.click_time)::date as click_date,
    count(ls.id)::int as clicks_count
FROM link_stats ls
         JOIN short_links sl ON ls.link_id = sl.id
WHERE
    sl.user_id = $1 AND
    ls.click_time >= $2 AND
    ls.click_time <= $3
GROUP BY click_date
ORDER BY click_date ASC;