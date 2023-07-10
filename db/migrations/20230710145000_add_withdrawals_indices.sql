-- +goose Up
-- +goose StatementBegin
ALTER TABLE blocks_withdrawals ADD COLUMN address_text TEXT NOT NULL DEFAULT '';
UPDATE blocks_withdrawals SET address_text = ENCODE(address, 'hex') WHERE address_text = '';

CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals (address_text);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_block_slot ON blocks_withdrawals (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_search ON blocks_withdrawals (validatorindex, block_slot, address_text);

ALTER TABLE blocks_bls_change ADD COLUMN pubkey_text TEXT NOT NULL DEFAULT '';
UPDATE blocks_bls_change SET pubkey_text = ENCODE(pubkey, 'hex') WHERE pubkey_text = '';

CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_pubkey_text ON blocks_bls_change (pubkey_text);
CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_block_slot ON blocks_bls_change (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_search ON blocks_bls_change (validatorindex, block_slot, pubkey_text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_blocks_withdrawals_address_text;
DROP INDEX IF EXISTS idx_blocks_withdrawals_block_slot;
DROP INDEX IF EXISTS idx_blocks_withdrawals_search;

ALTER TABLE blocks_withdrawals DROP COLUMN address_text;

DROP INDEX IF EXISTS idx_blocks_bls_change_pubkey_text;
DROP INDEX IF EXISTS idx_blocks_bls_change_block_slot;
DROP INDEX IF EXISTS idx_blocks_bls_change_search;

ALTER TABLE blocks_bls_change DROP COLUMN pubkey_text;
-- +goose StatementEnd
