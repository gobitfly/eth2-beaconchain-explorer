-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add total withdrawal count and amount columns to stats';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS withdrawals_total INT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS withdrawals_amount_total BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove total withdrawal count and amount columns from stats';
ALTER TABLE validator_stats DROP COLUMN IF EXISTS withdrawals_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS withdrawals_amount_total;
-- +goose StatementEnd
