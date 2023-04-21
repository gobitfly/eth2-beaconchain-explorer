-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS cl_proposer_rewards_gwei BIGINT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS cl_proposer_rewards_gwei_total BIGINT;
ALTER TABLE validator_performance ADD COLUMN IF NOT EXISTS cl_proposer_performance_total BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE validator_stats DROP COLUMN IF EXISTS cl_proposer_rewards_gwei;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS cl_proposer_rewards_gwei_total;
ALTER TABLE validator_performance DROP COLUMN IF EXISTS cl_proposer_performance_total;
-- +goose StatementEnd
