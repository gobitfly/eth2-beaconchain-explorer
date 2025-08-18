-- +goose Up
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_validatorindex ON blocks_bls_change (validatorindex);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS  idx_blocks_bls_change_validatorindex;
-- +goose StatementEnd
