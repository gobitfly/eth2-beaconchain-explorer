-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add total deposits count and amount columns to stats';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS deposits_total INT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS deposits_amount_total BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove total deposits count and amount columns from stats';
ALTER TABLE validator_stats DROP COLUMN IF EXISTS deposits_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS deposits_amount_total;
-- +goose StatementEnd
