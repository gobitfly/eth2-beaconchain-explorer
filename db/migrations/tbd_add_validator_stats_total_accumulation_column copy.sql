-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query with new column total_accumulation_exported for validator_stats_status';
ALTER TABLE validator_stats_status ADD COLUMN IF NOT EXISTS total_accumulation_exported BOOLEAN NOT NULL DEFAULT FALSE;

SELECT 'setting new validator_stats_status columns to true for already exported days'; 
UPDATE validator_stats_status
SET 
    total_accumulation_exported = true
WHERE total_performance_exported = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query column total_accumulation_exported for validator_stats_status';
ALTER TABLE validator_stats_status DROP COLUMN IF EXISTS total_accumulation_exported;
-- +goose StatementEnd
