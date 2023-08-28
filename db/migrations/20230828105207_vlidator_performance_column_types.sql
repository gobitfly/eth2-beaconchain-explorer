-- +goose NO TRANSACTION
-- +goose Up
SELECT 'up SQL query - change validator performance columns to numeric';
-- +goose StatementBegin
ALTER TABLE validator_performance 
    ALTER COLUMN el_performance_1d TYPE NUMERIC,
    ALTER COLUMN el_performance_7d TYPE NUMERIC,
    ALTER COLUMN el_performance_31d TYPE NUMERIC,
    ALTER COLUMN el_performance_365d TYPE NUMERIC,
    ALTER COLUMN el_performance_total TYPE NUMERIC,
    ALTER COLUMN mev_performance_1d TYPE NUMERIC,
    ALTER COLUMN mev_performance_7d TYPE NUMERIC,
    ALTER COLUMN mev_performance_31d TYPE NUMERIC,
    ALTER COLUMN mev_performance_365d TYPE NUMERIC,
    ALTER COLUMN mev_performance_total TYPE NUMERIC;
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query - we do not revert the validator performance columns to BIGINT as this could cause an out of range error'; 
-- +goose StatementBegin
-- +goose StatementEnd
