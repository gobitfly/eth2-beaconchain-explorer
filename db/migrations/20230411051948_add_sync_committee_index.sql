-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_sync_committees_period ON sync_committees (validatorindex, period DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sync_committees_period;
-- +goose StatementEnd
