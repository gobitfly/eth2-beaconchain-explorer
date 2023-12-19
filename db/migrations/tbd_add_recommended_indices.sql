-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add indices';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_attestations_block_root_block_slot ON public.blocks_attestations (block_root, block_slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_from_address_text_publickey ON public.eth1_deposits (from_address_text, publickey);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_valid_signature_block_ts ON public.eth1_deposits (valid_signature, block_ts);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_block_number ON public.eth1_deposits (block_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relays_blocks_proposer_pubkey_block_root ON public.relays_blocks (proposer_pubkey, block_root);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_deposits_withdrawalcredentials_block_root ON public.blocks_deposits (withdrawalcredentials, block_root);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_pubkey_block_root ON public.blocks_bls_change (pubkey, block_root);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop indices';
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_attestations_block_root_block_slot;
DROP INDEX CONCURRENTLY IF EXISTS idx_eth1_deposits_from_address_text_publickey;
DROP INDEX CONCURRENTLY IF EXISTS idx_eth1_deposits_valid_signature_block_ts;
DROP INDEX CONCURRENTLY IF EXISTS idx_eth1_deposits_block_number;
DROP INDEX CONCURRENTLY IF EXISTS idx_relays_blocks_proposer_pubkey_block_root;
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_deposits_withdrawalcredentials_block_root;
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_bls_change_pubkey_block_root;
-- +goose StatementEnd