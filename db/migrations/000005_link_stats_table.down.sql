-- Migration: 000005_link_stats_table.down.sql
-- Reverses the creation of the link_stats table and its indexes

-- Drop indexes
DROP INDEX IF EXISTS idx_link_stats_country;
DROP INDEX IF EXISTS idx_link_stats_click_time;
DROP INDEX IF EXISTS idx_link_stats_link_id;

-- Drop table
DROP TABLE IF EXISTS link_stats;