-- Migration Down
DROP TRIGGER IF EXISTS update_short_links_updated_at ON short_links;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS short_links;