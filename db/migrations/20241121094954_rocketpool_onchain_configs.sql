-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS rocketpool_onchain_configs (
    rocketpool_storage_address bytea NOT NULL,
    smoothing_pool_address bytea NOT NULL,
    PRIMARY KEY (rocketpool_storage_address)
);

ALTER TABLE rocketpool_minipools ADD COLUMN IF NOT EXISTS validator_index INTEGER;
CREATE INDEX IF NOT EXISTS rocketpool_minipools_validator_index_idx ON rocketpool_minipools (validator_index);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS rocketpool_onchain_configs;

DROP INDEX IF EXISTS rocketpool_minipools_validator_index_idx;
ALTER TABLE rocketpool_minipools DROP COLUMN IF EXISTS validator_index;

-- +goose StatementEnd
