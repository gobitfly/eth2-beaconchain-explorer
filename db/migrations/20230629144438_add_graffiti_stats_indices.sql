-- +goose NO TRANSACTION
-- +goose Up
SELECT 'up SQL query - add graffiti_stats indices';

-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_graffiti_stats_day ON graffiti_stats (day);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_graffiti_stats_day_graffiti_text ON graffiti_stats USING gin (graffiti_text gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query - remove graffiti_stats indices';

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_graffiti_stats_day;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_graffiti_stats_day_graffiti_text;
-- +goose StatementEnd
