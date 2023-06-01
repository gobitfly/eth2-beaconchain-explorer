-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query with new columns for validator_stats_status';
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS failed_attestations_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS sync_duties_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS withdrawals_deposits_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS balance_exported BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS cl_rewards_exported BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_validator_stats_status_failed_attestations ON validator_stats_status (day, failed_attestations_exported);
CREATE INDEX IF NOT EXISTS idx_validator_stats_status_sync_duties ON validator_stats_status (day, sync_duties_exported);
CREATE INDEX IF NOT EXISTS idx_validator_stats_status_withdrawals_deposits ON validator_stats_status (day, withdrawals_deposits_exported);
CREATE INDEX IF NOT EXISTS idx_validator_stats_status_balance ON validator_stats_status (day, balance_exported);
CREATE INDEX IF NOT EXISTS idx_validator_stats_cl_rewards_exported ON validator_stats_status (day, cl_rewards_exported);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query columns for validator_stats_status';
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS failed_attestations_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS sync_duties_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS withdrawals_deposits_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS balance_exported;
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS cl_rewards_exported;
DROP INDEX IF EXISTS idx_validator_stats_status_failed_attestations;
DROP INDEX IF EXISTS idx_validator_stats_status_sync_duties;
DROP INDEX IF EXISTS idx_validator_stats_status_withdrawals_deposits;
DROP INDEX IF EXISTS idx_validator_stats_status_balance;
DROP INDEX IF EXISTS idx_validator_stats_cl_rewards_exported;
-- +goose StatementEnd
