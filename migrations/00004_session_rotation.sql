-- +goose Up

-- Tracks the lineage of a refresh token chain so we can distinguish a
-- legitimate concurrent-refresh race from genuine token reuse, and so we
-- can revoke every descendant of a compromised session in one query.

ALTER TABLE user_sessions
    ADD COLUMN family_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN replaced_by UUID REFERENCES user_sessions(id),
    ADD COLUMN revoked_at TIMESTAMPTZ;

CREATE INDEX idx_usersessions_family_id ON user_sessions(family_id);


-- +goose Down
DROP INDEX IF EXISTS idx_usersessions_family_id;

ALTER TABLE user_sessions
    DROP COLUMN family_id,
    DROP COLUMN replaced_by,
    DROP COLUMN revoked_at;
