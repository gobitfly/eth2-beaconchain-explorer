-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add total missed attestations and total participated, missed and orphaned syncs';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS missed_attestations_total INT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS participated_sync_total INT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS missed_sync_total INT;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS orphaned_sync_total INT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove total missed attestations and total participated, missed and orphaned syncs';
ALTER TABLE validator_stats DROP COLUMN IF EXISTS missed_attestations_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS participated_sync_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS missed_sync_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS orphaned_sync_total;
-- +goose StatementEnd
