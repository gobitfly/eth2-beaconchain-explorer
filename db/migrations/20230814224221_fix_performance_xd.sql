-- +goose Up
-- +goose StatementBegin
SELECT 'delete all unneeded performance xd columns';
ALTER TABLE validator_performance DROP COLUMN IF EXISTS performance1d;
ALTER TABLE validator_performance DROP COLUMN IF EXISTS performance7d;
ALTER TABLE validator_performance DROP COLUMN IF EXISTS performance31d;
ALTER TABLE validator_performance DROP COLUMN IF EXISTS performance365d;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'add all unneeded performance xd columns';
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS performance1d BIGINT NOT NULL;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS performance7d BIGINT NOT NULL;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS performance31d BIGINT NOT NULL;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS performance365d BIGINT NOT NULL;
-- +goose StatementEnd