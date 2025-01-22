-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add is_archived column';
ALTER TABLE users_val_dashboards ADD COLUMN IF NOT EXISTS is_archived TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop is_archived column';
ALTER TABLE users_val_dashboards DROP COLUMN IF EXISTS is_archived;
-- +goose StatementEnd