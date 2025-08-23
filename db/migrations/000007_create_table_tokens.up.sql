CREATE TYPE token_type AS ENUM ('registration_verification', 'password_reset', 'email_change_verification');

CREATE TABLE tokens
(
    id          UUID                     PRIMARY KEY ,
    user_id     UUID                     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255)             NOT NULL UNIQUE,
    type        token_type               NOT NULL,
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_type ON tokens(type);