-- +goose Up
-- +goose StatementBegin

SELECT('up SQL query - add msg_sender, to_address, and log_index columns to eth1_deposits');
ALTER TABLE eth1_deposits ADD msg_sender bytea NULL;
ALTER TABLE eth1_deposits ADD to_address bytea NULL;
ALTER TABLE eth1_deposits ADD log_index int4 NULL;

SELECT('up SQL query - remove duplicate rows from eth1_deposits');
delete from
	eth1_deposits
where
	merkletree_index in (
	select
		merkletree_index
	from
		(
		select
			merkletree_index,
			row_number() over (partition by merkletree_index) as rn,
			COUNT(*) over (partition by merkletree_index) as cnt
		from
			eth1_deposits
) t
	where
		t.cnt > 1
);

SELECT('up SQL query - changing the primary key of eth1_deposits to be merkletree_index alone');
ALTER TABLE eth1_deposits DROP CONSTRAINT IF EXISTS eth1_deposits_pkey;
ALTER TABLE eth1_deposits ADD PRIMARY KEY (merkletree_index);

SELECT('up SQL query - add block_number index to eth1_deposits');
CREATE INDEX eth1_deposits_block_number_idx ON eth1_deposits (block_number);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SELECT('down SQL query - remove msg_sender, to_address, and log_index columns from eth1_deposits');
ALTER TABLE eth1_deposits DROP COLUMN msg_sender;
ALTER TABLE eth1_deposits DROP COLUMN to_address;
ALTER TABLE eth1_deposits DROP COLUMN log_index;

SELECT('down SQL query - changing the primary key of eth1_deposits back to be tx_hash & merkletree_index');
ALTER TABLE eth1_deposits DROP CONSTRAINT IF EXISTS eth1_deposits_pkey;
ALTER TABLE eth1_deposits ADD PRIMARY KEY (tx_hash, merkletree_index);

SELECT('down SQL query - remove block_number index from eth1_deposits');
DROP INDEX IF EXISTS eth1_deposits_block_number_idx;

-- +goose StatementEnd
