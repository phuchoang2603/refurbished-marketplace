-- +goose Up
ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS family_id UUID;
UPDATE refresh_tokens SET family_id = id WHERE family_id IS NULL;
ALTER TABLE refresh_tokens ALTER COLUMN family_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS refresh_tokens_family_id_idx ON refresh_tokens (
    family_id
);

-- +goose Down
DROP INDEX IF EXISTS refresh_tokens_family_id_idx;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS family_id;
