-- +goose Down
ALTER TABLE pastes
    DROP COLUMN IF EXISTS title;
