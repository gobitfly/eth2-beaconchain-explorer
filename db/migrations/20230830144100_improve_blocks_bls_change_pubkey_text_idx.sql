-- +goose NO TRANSACTION

-- +goose Up

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_bls_change_pubkey_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_pubkey_text ON blocks_bls_change USING gin (pubkey_text gin_trgm_ops);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_address_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals USING gin (address_text gin_trgm_ops);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address ON blocks_withdrawals (address);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_bls_change_pubkey_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_pubkey_text ON blocks_bls_change (pubkey_text);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_address_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_search ON blocks_withdrawals (validatorindex, block_slot, address_text);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_address;
-- +goose StatementEnd
