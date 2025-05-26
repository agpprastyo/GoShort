-- Migration Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS short_links (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL,
    title VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    click_limit INTEGER,
    expired_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT short_code_unique UNIQUE (short_code)
);

-- Add indexes for performance
CREATE INDEX idx_short_links_user_id ON short_links(user_id);
CREATE INDEX idx_short_links_short_code ON short_links(short_code);
CREATE INDEX idx_short_links_is_active ON short_links(is_active);

-- Add trigger to update updated_at automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_short_links_updated_at
    BEFORE UPDATE ON short_links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

