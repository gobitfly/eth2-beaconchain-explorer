-- +goose Up
-- +goose StatementBegin
SELECT 'delete chain_head table';
DROP TABLE IF EXISTS chain_head;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'create chain_head table';
CREATE TABLE IF NOT EXISTS
    chain_head (
        finalized_block_root bytea,
        finalized_epoch INT,
        finalized_slot INT,
        head_block_root bytea,
        head_epoch INT,
        head_slot INT,
        justified_block_root bytea,
        justified_epoch INT,
        justified_slot INT,
        previous_justified_block_root bytea,
        previous_justified_epoch INT,
        previous_justified_slot INT,
        PRIMARY KEY (finalized_epoch)
    );
-- +goose StatementEnd
