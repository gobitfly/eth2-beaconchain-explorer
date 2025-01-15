-- +goose NO TRANSACTION
-- +goose Up
SELECT 'creating idx_blocks_withdrawals_validatorindex_slot';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_validatorindex_slot ON blocks_withdrawals (validatorindex, block_slot DESC);
-- +goose StatementEnd

-- +goose Down
SELECT 'dropping idx_blocks_withdrawals_validatorindex_slot';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_validatorindex_slot;
-- +goose StatementEnd
