-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column for blocks_withdrawals';
ALTER TABLE blocks_withdrawals ADD COLUMN IF NOT EXISTS address_text TEXT NOT NULL DEFAULT '';

SELECT 'populate new blocks_withdrawals column'; 
UPDATE blocks_withdrawals SET address_text = ENCODE(address, 'hex') WHERE address_text = '';

SELECT 'create new blocks_withdrawals indices'; 
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals (address_text);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_block_slot ON blocks_withdrawals (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_search ON blocks_withdrawals (validatorindex, block_slot, address_text);

SELECT 'add column for blocks_bls_change';
ALTER TABLE blocks_bls_change ADD COLUMN IF NOT EXISTS pubkey_text TEXT NOT NULL DEFAULT '';

SELECT 'populate new blocks_bls_change column'; 
UPDATE blocks_bls_change SET pubkey_text = ENCODE(pubkey, 'hex') WHERE pubkey_text = '';

SELECT 'create new blocks_bls_change indices'; 
CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_pubkey_text ON blocks_bls_change (pubkey_text);
CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_block_slot ON blocks_bls_change (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_search ON blocks_bls_change (validatorindex, block_slot, pubkey_text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop blocks_withdrawals indices'; 
DROP INDEX IF EXISTS idx_blocks_withdrawals_address_text;
DROP INDEX IF EXISTS idx_blocks_withdrawals_block_slot;
DROP INDEX IF EXISTS idx_blocks_withdrawals_search;

SELECT 'drop column for blocks_withdrawals';
ALTER TABLE blocks_withdrawals DROP COLUMN IF EXISTS address_text;

SELECT 'drop blocks_bls_change indices'; 
DROP INDEX IF EXISTS idx_blocks_bls_change_pubkey_text;
DROP INDEX IF EXISTS idx_blocks_bls_change_block_slot;
DROP INDEX IF EXISTS idx_blocks_bls_change_search;

SELECT 'drop column for blocks_bls_change';
ALTER TABLE blocks_bls_change DROP COLUMN IF EXISTS pubkey_text;
-- +goose StatementEnd
