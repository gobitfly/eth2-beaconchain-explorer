-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add an index for the amount of the withdrawals';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_amount ON blocks_withdrawals (amount);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove the index for the amount of the withdrawals';
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_amount;
-- +goose StatementEnd
