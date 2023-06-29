-- +goose NO TRANSACTION
-- +goose Up

SELECT 'create idx_blocks_withdrawals_withdrawalindex';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_withdrawalindex ON public.blocks_withdrawals USING btree (withdrawalindex DESC);
-- +goose StatementEnd

-- +goose Down
SELECT 'drop idx_blocks_withdrawals_withdrawalindex';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_withdrawalindex;
-- +goose StatementEnd
