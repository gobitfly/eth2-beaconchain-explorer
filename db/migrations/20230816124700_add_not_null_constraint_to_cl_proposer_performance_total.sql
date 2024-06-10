-- +goose NO TRANSACTION
-- +goose Up
SELECT 'up SQL query - add not null constraint to column cl_proposer_performance_total in validator_performance';
-- +goose StatementBegin
ALTER TABLE validator_performance ALTER COLUMN cl_proposer_performance_total SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
SELECT 'up SQL query - drop not null constraint to column cl_proposer_performance_total in validator_performance';
-- +goose StatementBegin
ALTER TABLE validator_performance ALTER COLUMN cl_proposer_performance_total DROP NOT NULL;
-- +goose StatementEnd