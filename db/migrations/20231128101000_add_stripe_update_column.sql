-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column stripe_email_pending';
ALTER TABLE users ADD COLUMN IF NOT EXISTS stripe_email_pending BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove column stripe_email_pending';
ALTER TABLE users DROP COLUMN IF EXISTS stripe_email_pending;
-- +goose StatementEnd
