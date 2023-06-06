-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query with new columns for validator_stats_status';
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS failed_attestations_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS sync_duties_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS withdrawals_deposits_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS balance_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS cl_rewards_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS el_rewards_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS total_performance_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS block_stats_exported BOOLEAN NOT NULL DEFAULT FALSE;

SELECT 'setting new validator_stats_status columns to true for already exported days'; 
UPDATE validator_stats_status
SET 
    failed_attestations_exported = true, 
    sync_duties_exported = true, 
    withdrawals_deposits_exported = true, 
    balance_exported = true, 
    cl_rewards_exported = true, 
    el_rewards_exported = true, 
    total_performance_exported = true, 
    block_stats_exported = true
WHERE status=true;

SELECT 'dropping unused income_exported column'; 
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS income_exported;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query columns for validator_stats_status';
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS failed_attestations_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS sync_duties_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS withdrawals_deposits_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS balance_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS cl_rewards_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS el_rewards_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS total_performance_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS block_stats_exported;
-- +goose StatementEnd
