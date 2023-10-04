-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS blocks_transactions;
DROP TABLE IF EXISTS validator_balances_recent;

ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_blob_gas_used INT NOT NULL DEFAULT 0;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_excess_blob_gas INT NOT NULL DEFAULT 0;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_blob_transactions_count INT NOT NULL DEFAULt 0;

CREATE TABLE IF NOT EXISTS
    blocks_blob_sidecars (
        block_slot INT NOT NULL,
        block_root BYTEA NOT NULL,
        index INT NOT NULL,
        kzg_commitment BYTEA NOT NULL,
        kzg_proof BYTEA NOT NULL,
        blob_versioned_hash BYTEA NOT NULL,
        PRIMARY KEY (block_root, index)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
    blocks_transactions (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        raw bytea NOT NULL,
        txhash bytea NOT NULL,
        nonce INT NOT NULL,
        gas_price bytea NOT NULL,
        gas_limit BIGINT NOT NULL,
        sender bytea NOT NULL,
        recipient bytea NOT NULL,
        amount bytea NOT NULL,
        payload bytea NOT NULL,
        max_priority_fee_per_gas BIGINT,
        max_fee_per_gas BIGINT,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    validator_balances_recent (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        balance BIGINT NOT NULL,
        PRIMARY KEY (epoch, validatorindex)
    );

ALTER TABLE blocks DROP COLUMN IF EXISTS exec_blob_gas_used;
ALTER TABLE blocks DROP COLUMN IF EXISTS exec_excess_blob_gas;
ALTER TABLE blocks DROP COLUMN IF EXISTS exec_blob_transactions_count;
-- +goose StatementEnd