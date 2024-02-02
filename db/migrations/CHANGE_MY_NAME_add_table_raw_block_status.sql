-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add table raw_block_status';
CREATE TABLE IF NOT EXISTS
    raw_block_status (
        chain_id INT NOT NULL,
        block_id INT NOT NULL,
        block_hash bytea NOT NULL UNIQUE,
        indexed_bt bool NOT NULL DEFAULT FALSE,
        PRIMARY KEY (chain_id, block_id)
    );
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_raw_block_status_block_hash ON raw_block_status (block_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove table raw_block_status';
DROP INDEX CONCURRENTLY idx_raw_block_status_block_hash;
DROP TABLE IF EXISTS raw_block_status CASCADE;
-- +goose StatementEnd
