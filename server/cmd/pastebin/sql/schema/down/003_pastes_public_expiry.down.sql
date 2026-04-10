-- +goose Down
ALTER TABLE pastes
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS is_public;
