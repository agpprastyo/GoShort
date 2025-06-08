-- Create link_stats table
CREATE TABLE IF NOT EXISTS link_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    link_id UUID NOT NULL REFERENCES short_links (id) ON DELETE CASCADE,
    click_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT,
    country TEXT,
    device_type TEXT,

    CONSTRAINT fk_link_id FOREIGN KEY (link_id)
        REFERENCES short_links(id) ON DELETE CASCADE
);

-- Create indexes for common queries
CREATE INDEX idx_link_stats_link_id ON link_stats(link_id);
CREATE INDEX idx_link_stats_click_time ON link_stats(click_time);
CREATE INDEX idx_link_stats_country ON link_stats(country);