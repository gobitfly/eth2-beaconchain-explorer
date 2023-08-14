-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column eth1_deposits for from_address_text';
ALTER TABLE eth1_deposits ADD COLUMN from_address_text TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create new eth1_deposits index for from_address_text'; 
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_from_address_text ON eth1_deposits (from_address_text);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'populate new eth1_deposits column from_address_text, 1000 at a time'; 
DO
$do$
DECLARE
    count INTEGER default 0;
BEGIN
LOOP
    count := (count + 1);
    RAISE NOTICE 'count = %', count;
    EXIT WHEN (SELECT count(*) FROM (SELECT from_address_text FROM eth1_deposits WHERE from_address_text = '' LIMIT(1)) AS sub) = 0;
    WITH to_update AS (
        SELECT tx_hash, merkletree_index
        FROM eth1_deposits
        WHERE from_address_text = ''
        LIMIT 1000
    )
    UPDATE eth1_deposits SET from_address_text = ENCODE(from_address, 'hex')  
    WHERE EXISTS (SELECT * FROM to_update 
   	    WHERE eth1_deposits.tx_hash = to_update.tx_hash
	    AND eth1_deposits.merkletree_index = to_update.merkletree_index);
END LOOP;
END;
$do$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop from_address_text index from eth1_deposits '; 
DROP INDEX IF EXISTS idx_eth1_deposits_from_address_text;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'drop column from_address_text from eth1_deposits';
ALTER TABLE eth1_deposits DROP COLUMN from_address_text;
-- +goose StatementEnd
