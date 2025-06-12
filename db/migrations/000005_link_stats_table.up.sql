-- Migration: 000005_link_stats_table.up.sql

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create link_stats table
CREATE TABLE IF NOT EXISTS link_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    link_id UUID NOT NULL,
    click_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT,
    country TEXT,
    device_type TEXT,

    -- Single foreign key constraint
    CONSTRAINT fk_link_stats_link_id FOREIGN KEY (link_id)
        REFERENCES short_links(id) ON DELETE CASCADE
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_link_stats_link_id ON link_stats(link_id);
CREATE INDEX IF NOT EXISTS idx_link_stats_click_time ON link_stats(click_time);
CREATE INDEX IF NOT EXISTS idx_link_stats_country ON link_stats(country);

-- Set ownership if needed
-- ALTER TABLE link_stats OWNER TO postgres;