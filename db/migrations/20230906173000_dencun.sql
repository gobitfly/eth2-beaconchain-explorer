-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_balances_recent;

ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_blob_gas_used INT NOT NULL DEFAULT 0;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_excess_blob_gas INT NOT NULL DEFAULT 0;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS exec_blob_transactions_count INT NOT NULL DEFAULt 0;
CREATE TABLE IF NOT EXISTS
    blocks_blobs (
        block_slot INT NOT NULL,
        block_root BYTEA NOT NULL,
        index INT NOT NULL,
        kzg_commitment BYTEA NOT NULL,
        blob_versioned_hash BYTEA NOT NULL,
        PRIMARY KEY (block_root, index)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd