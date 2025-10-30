-- +goose Up
-- Pectra (Electra) upgrade database schema changes

-- +goose StatementBegin
SELECT 'adding committeebits column to blocks_attestations table';
ALTER TABLE blocks_attestations ADD COLUMN IF NOT EXISTS committeebits bytea;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_deposit_requests_v2 table';
CREATE TABLE IF NOT EXISTS blocks_deposit_requests_v2 (
    slot_processed INT NOT NULL,
    index_processed INT NOT NULL,
    block_processed_root bytea NOT NULL,
    pubkey bytea NOT NULL,
    withdrawal_credentials bytea NOT NULL,
    amount BIGINT NOT NULL,
    signature bytea NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (slot_processed, index_processed)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_withdrawal_requests_v2 table';
CREATE TABLE IF NOT EXISTS blocks_withdrawal_requests_v2 (
    slot_processed INT NOT NULL,
    index_processed INT NOT NULL,
    block_processed_root bytea NOT NULL,
    source_address bytea NOT NULL,
    validator_pubkey bytea NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (slot_processed, index_processed)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_consolidation_requests_v2 table';
CREATE TABLE IF NOT EXISTS blocks_consolidation_requests_v2 (
    slot_processed INT NOT NULL,
    index_processed INT NOT NULL,
    block_processed_root bytea NOT NULL,
    source_pubkey bytea NOT NULL,
    target_pubkey bytea NOT NULL,
    amount_consolidated BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (slot_processed, index_processed)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_switch_to_compounding_requests_v2 table';
CREATE TABLE IF NOT EXISTS blocks_switch_to_compounding_requests_v2 (
    slot_processed INT NOT NULL,
    index_processed INT NOT NULL,
    block_processed_root bytea NOT NULL,
    validator_pubkey bytea NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (slot_processed, index_processed)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_deposit_requests_v2';
CREATE INDEX IF NOT EXISTS idx_blocks_deposit_requests_v2_pubkey ON blocks_deposit_requests_v2 (pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_deposit_requests_v2_status ON blocks_deposit_requests_v2 (status);
CREATE INDEX IF NOT EXISTS idx_blocks_deposit_requests_v2_slot ON blocks_deposit_requests_v2 (slot_processed);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_withdrawal_requests_v2';
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_v2_pubkey ON blocks_withdrawal_requests_v2 (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_v2_address ON blocks_withdrawal_requests_v2 (source_address);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_v2_status ON blocks_withdrawal_requests_v2 (status);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_v2_slot ON blocks_withdrawal_requests_v2 (slot_processed);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_consolidation_requests_v2';
CREATE INDEX IF NOT EXISTS idx_blocks_consolidation_requests_v2_source_pubkey ON blocks_consolidation_requests_v2 (source_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_consolidation_requests_v2_target_pubkey ON blocks_consolidation_requests_v2 (target_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_consolidation_requests_v2_status ON blocks_consolidation_requests_v2 (status);
CREATE INDEX IF NOT EXISTS idx_blocks_consolidation_requests_v2_slot ON blocks_consolidation_requests_v2 (slot_processed);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_switch_to_compounding_requests_v2';
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_v2_pubkey ON blocks_switch_to_compounding_requests_v2 (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_v2_status ON blocks_switch_to_compounding_requests_v2 (status);
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_v2_slot ON blocks_switch_to_compounding_requests_v2 (slot_processed);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping blocks_switch_to_compounding_requests_v2 table';
DROP TABLE IF EXISTS blocks_switch_to_compounding_requests_v2;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_consolidation_requests_v2 table';
DROP TABLE IF EXISTS blocks_consolidation_requests_v2;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_withdrawal_requests_v2 table';
DROP TABLE IF EXISTS blocks_withdrawal_requests_v2;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_deposit_requests_v2 table';
DROP TABLE IF EXISTS blocks_deposit_requests_v2;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'removing committeebits column from blocks_attestations table';
ALTER TABLE blocks_attestations DROP COLUMN IF EXISTS committeebits;
-- +goose StatementEnd