-- +goose Up
-- +goose StatementBegin
SELECT 'creating idx_blocks_withdrawals_validatorindex_slot';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_validatorindex_slot ON blocks_withdrawals (validatorindex, block_slot DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping idx_blocks_withdrawals_validatorindex_slot';
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_validatorindex_slot;
-- +goose StatementEnd
