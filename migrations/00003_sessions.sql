-- +goose Up

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT  gen_random_uuid(),

    user_id UUID NOT NULL,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    is_active boolean NOT NULL DEFAULT TRUE,
    user_agent TEXT,
    ip_address TEXT,

    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_usersessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_usersessions_refresh_token_hash ON user_sessions(refresh_token_hash);

CREATE TRIGGER trg_user_sessions_updated_at
BEFORE UPDATE ON user_sessions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();



-- +goose Down
DROP TABLE user_sessions;
DROP TRIGGER IF EXISTS trg_user_sessions_updated_at ON user_sessions;

