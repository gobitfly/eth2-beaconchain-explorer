-- +goose NO TRANSACTION
-- +goose Up
SELECT 'up SQL query - drop default values from columns in validator_performance';
-- +goose StatementBegin
ALTER TABLE validator_performance ALTER COLUMN cl_performance_1d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_7d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_31d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_365d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_total DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN el_performance_1d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN el_performance_7d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN el_performance_31d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN el_performance_365d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN el_performance_total DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_1d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_7d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_31d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_365d DROP DEFAULT;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_total DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
SELECT 'up SQL query - add default values to columns in validator_performance';
-- +goose StatementBegin
ALTER TABLE validator_performance ALTER COLUMN cl_performance_1d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_7d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_31d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_365d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN cl_performance_total SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN el_performance_1d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN el_performance_7d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN el_performance_31d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN el_performance_365d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN el_performance_total SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_1d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_7d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_31d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_365d SET DEFAULT 0;
ALTER TABLE validator_performance ALTER COLUMN mev_performance_total SET DEFAULT 0;
-- +goose StatementEnd