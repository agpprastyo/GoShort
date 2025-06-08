-- Create enum type for shortlink ordering
CREATE TYPE shortlink_order_column AS ENUM (
    'title',
    'is_active',
    'created_at',
    'updated_at',
    'expired_at'
);