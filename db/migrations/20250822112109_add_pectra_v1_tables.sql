-- +goose Up
-- Missing Pectra (Electra) v1 tables

-- +goose StatementBegin
SELECT 'creating blocks_exit_requests table';
CREATE TABLE IF NOT EXISTS blocks_exit_requests (
    block_slot INT NOT NULL,
    block_index INT NOT NULL,
    block_root bytea NOT NULL DEFAULT '',
    validator_pubkey bytea NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reject_reason VARCHAR(255),
    block_processed_root bytea,
    slot_processed INT,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (block_slot, block_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_switch_to_compounding_requests table';
CREATE TABLE IF NOT EXISTS blocks_switch_to_compounding_requests (
    block_slot INT NOT NULL,
    block_index INT NOT NULL,
    block_root bytea NOT NULL DEFAULT '',
    validator_pubkey bytea NOT NULL,
    request_index INT NOT NULL,
    address bytea,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (block_slot, block_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_exit_requests';
CREATE INDEX IF NOT EXISTS idx_blocks_exit_requests_validator_pubkey ON blocks_exit_requests (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_exit_requests_status ON blocks_exit_requests (status);
CREATE INDEX IF NOT EXISTS idx_blocks_exit_requests_slot ON blocks_exit_requests (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_exit_requests_block_processed_root ON blocks_exit_requests (block_processed_root);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_switch_to_compounding_requests';
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_validator_pubkey ON blocks_switch_to_compounding_requests (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_status ON blocks_switch_to_compounding_requests (status);
CREATE INDEX IF NOT EXISTS idx_blocks_switch_to_compounding_requests_slot ON blocks_switch_to_compounding_requests (block_slot);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating blocks_withdrawal_requests table';
CREATE TABLE IF NOT EXISTS blocks_withdrawal_requests (
    block_slot INT NOT NULL,
    block_index INT NOT NULL,
    block_root bytea NOT NULL DEFAULT '',
    request_index INT NOT NULL,
    source_address bytea,
    validator_pubkey bytea NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (block_slot, block_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating eth1_consolidation_requests table';
CREATE TABLE IF NOT EXISTS eth1_consolidation_requests (
    tx_hash bytea NOT NULL,
    tx_index INT NOT NULL,
    itx_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    block_ts TIMESTAMP NOT NULL,
    from_address bytea NOT NULL,
    fee BIGINT NOT NULL,
    source_address bytea NOT NULL,
    source_pubkey bytea NOT NULL,
    target_pubkey bytea NOT NULL,
    PRIMARY KEY (tx_hash, tx_index, itx_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for blocks_withdrawal_requests';
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_validator_pubkey ON blocks_withdrawal_requests (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_source_address ON blocks_withdrawal_requests (source_address);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_status ON blocks_withdrawal_requests (status);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawal_requests_slot ON blocks_withdrawal_requests (block_slot);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating eth1_withdrawal_requests table';
CREATE TABLE IF NOT EXISTS eth1_withdrawal_requests (
    tx_hash bytea NOT NULL,
    tx_index INT NOT NULL,
    itx_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    block_ts TIMESTAMP NOT NULL,
    from_address bytea NOT NULL,
    fee BIGINT NOT NULL,
    source_address bytea NOT NULL,
    validator_pubkey bytea NOT NULL,
    amount BIGINT NOT NULL,
    PRIMARY KEY (tx_hash, tx_index, itx_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for eth1_consolidation_requests';
CREATE INDEX IF NOT EXISTS idx_eth1_consolidation_requests_source_pubkey ON eth1_consolidation_requests (source_pubkey);
CREATE INDEX IF NOT EXISTS idx_eth1_consolidation_requests_target_pubkey ON eth1_consolidation_requests (target_pubkey);
CREATE INDEX IF NOT EXISTS idx_eth1_consolidation_requests_block_number ON eth1_consolidation_requests (block_number);
CREATE INDEX IF NOT EXISTS idx_eth1_consolidation_requests_block_ts ON eth1_consolidation_requests (block_ts);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'creating indexes for eth1_withdrawal_requests';
CREATE INDEX IF NOT EXISTS idx_eth1_withdrawal_requests_validator_pubkey ON eth1_withdrawal_requests (validator_pubkey);
CREATE INDEX IF NOT EXISTS idx_eth1_withdrawal_requests_source_address ON eth1_withdrawal_requests (source_address);
CREATE INDEX IF NOT EXISTS idx_eth1_withdrawal_requests_block_number ON eth1_withdrawal_requests (block_number);
CREATE INDEX IF NOT EXISTS idx_eth1_withdrawal_requests_block_ts ON eth1_withdrawal_requests (block_ts);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping eth1_withdrawal_requests table';
DROP TABLE IF EXISTS eth1_withdrawal_requests;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping eth1_consolidation_requests table';
DROP TABLE IF EXISTS eth1_consolidation_requests;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_withdrawal_requests table';
DROP TABLE IF EXISTS blocks_withdrawal_requests;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_switch_to_compounding_requests table';
DROP TABLE IF EXISTS blocks_switch_to_compounding_requests;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'dropping blocks_exit_requests table';
DROP TABLE IF EXISTS blocks_exit_requests;
-- +goose StatementEnd