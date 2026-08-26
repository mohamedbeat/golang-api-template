-- +goose Up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP EXTENSION IF EXISTS "pgcrypto";
DROP FUNCTION IF EXISTS set_updated_at();
