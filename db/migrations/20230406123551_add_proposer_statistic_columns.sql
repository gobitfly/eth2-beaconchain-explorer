-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS cl_proposer_rewards_gwei INTEGER;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS cl_proposer_rewards_gwei_total INTEGER;
ALTER TABLE validator_performance ADD COLUMN IF NOT EXISTS cl_proposer_performance_total INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
