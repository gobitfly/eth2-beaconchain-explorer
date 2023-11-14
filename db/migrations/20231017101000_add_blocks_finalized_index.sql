-- +goose NO TRANSACTION

-- +goose Up

-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_finalized ON blocks (finalized);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_finalized;
-- +goose StatementEnd
