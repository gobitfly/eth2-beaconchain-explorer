-- +goose NO TRANSACTION

-- +goose Up

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_bls_change_pubkey_text;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE blocks_bls_change DROP COLUMN IF EXISTS pubkey_text;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_search;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_address_text;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE blocks_withdrawals DROP COLUMN IF EXISTS address_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address ON blocks_withdrawals (address);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
ALTER TABLE blocks_withdrawals ADD COLUMN IF NOT EXISTS address_text TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals USING gin (address_text gin_trgm_ops);
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE blocks_bls_change ADD COLUMN IF NOT EXISTS pubkey_text TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_pubkey_text ON blocks_bls_change USING gin (pubkey_text gin_trgm_ops);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_search ON blocks_bls_change (validatorindex, block_slot, pubkey_text);
-- +goose StatementEnd
