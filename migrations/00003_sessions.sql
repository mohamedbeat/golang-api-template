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
    family_id UUID NOT NULL DEFAULT gen_random_uuid(), 
    replaced_by UUID REFERENCES user_sessions(id),     
    revoked_at TIMESTAMPTZ,                                
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_usersessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_usersessions_refresh_token_hash ON user_sessions(refresh_token_hash);
CREATE INDEX idx_usersessions_family_id ON user_sessions(family_id);

CREATE TRIGGER trg_user_sessions_updated_at
BEFORE UPDATE ON user_sessions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();



-- +goose Down
DROP INDEX IF EXISTS idx_usersessions_family_id;
DROP INDEX IF EXISTS idx_usersessions_user_id;
DROP INDEX IF EXISTS idx_usersessions_refresh_token_hash;
DROP TABLE user_sessions;
DROP TRIGGER IF EXISTS trg_user_sessions_updated_at ON user_sessions;

