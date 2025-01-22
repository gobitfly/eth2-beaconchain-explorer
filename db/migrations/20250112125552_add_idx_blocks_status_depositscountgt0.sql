-- +goose NO TRANSACTION

-- +goose Up
-- +goose StatementBegin
SELECT 'creating idx_blocks_status_depositscountgt0';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_depositscountgt0 ON blocks (status, (depositscount > 0)) where depositscount > 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping idx_blocks_status_depositscountgt0';
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_status_depositscountgt0;
-- +goose StatementEnd
