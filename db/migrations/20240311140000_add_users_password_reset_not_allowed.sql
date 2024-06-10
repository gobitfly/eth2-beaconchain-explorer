-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column users.password_reset_not_allowed';
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_not_allowed BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop column users.password_reset_not_allowed';
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_not_allowed;
-- +goose StatementEnd
