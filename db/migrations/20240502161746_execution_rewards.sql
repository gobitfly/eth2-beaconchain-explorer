-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - set all empty execution_block_hashes to NULL');
UPDATE blocks SET exec_block_hash = NULL WHERE exec_block_hash = '\x';
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('up SQL query - add unique constraint on execution_block_hashes');
ALTER TABLE blocks ADD CONSTRAINT blocks_exec_block_hash_unique UNIQUE (exec_block_hash);
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('up SQL query - create execution_payload table');
CREATE TABLE execution_payloads (
	block_hash bytea NOT NULL,
	fee_recipient_reward numeric(78, 18) NULL,
	CONSTRAINT execution_payloads_pk PRIMARY KEY (block_hash)
);
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('up SQL query - create index on execution_payloads table');
CREATE UNIQUE INDEX execution_payloads_block_hash_idx ON public.execution_payloads USING btree (block_hash, fee_recipient_reward);
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('up SQL query - prefilling execution_payloads table with empty values');
INSERT INTO execution_payloads (block_hash) SELECT exec_block_hash FROM blocks where exec_block_hash IS NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('up SQL query - add foreign key constraint to blocks table');
ALTER TABLE blocks ADD CONSTRAINT blocks_execution_payloads_fk FOREIGN KEY (exec_block_hash) REFERENCES execution_payloads(block_hash);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - drop foreign key constraint to execution_payloads table');
ALTER TABLE blocks DROP CONSTRAINT blocks_execution_payloads_fk;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('down SQL query - drop execution_payloads table');
DROP TABLE execution_payloads;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('down SQL query - drop unique constraint on execution_block_hashes');
ALTER TABLE blocks DROP CONSTRAINT blocks_exec_block_hash_unique;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT('down SQL query - set all NULL execution_block_hashes to empty (better safe than sorry)');
UPDATE blocks SET exec_block_hash = '\x' WHERE exec_block_hash IS NULL;
-- +goose StatementEnd
