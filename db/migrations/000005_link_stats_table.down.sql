-- Drop indexes
DROP INDEX IF EXISTS idx_link_stats_country;
DROP INDEX IF EXISTS idx_link_stats_click_time;
DROP INDEX IF EXISTS idx_link_stats_link_id;

-- Drop table
DROP TABLE IF EXISTS link_stats;