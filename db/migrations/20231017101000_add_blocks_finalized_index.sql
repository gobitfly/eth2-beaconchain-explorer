-- +goose NO TRANSACTION

-- +goose Up

-- +goose StatementBegin
SELECT 'create index idx_blocks_finalized';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_finalized ON blocks (finalized);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
SELECT 'drop index idx_blocks_finalized';
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_finalized;
-- +goose StatementEnd
