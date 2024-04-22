-- +goose NO TRANSACTION

-- +goose Up
SELECT 'up SQL query - add an index for the amount of the withdrawals';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_amount ON blocks_withdrawals (amount);
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query - remove the index for the amount of the withdrawals';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_amount;
-- +goose StatementEnd
