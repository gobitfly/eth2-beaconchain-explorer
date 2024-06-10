-- +goose Up
-- +goose StatementBegin
SELECT 'create index for withdrawals recipient';
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_recipient_index ON blocks_withdrawals (address, withdrawalindex DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'drop index for withdrawals recipient';
DROP INDEX IF EXISTS idx_blocks_withdrawals_recipient_index;
-- +goose StatementEnd
