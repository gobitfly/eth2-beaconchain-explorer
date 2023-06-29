-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - remove graffiti_stats tables';
CREATE TABLE IF NOT EXISTS
    graffiti_stats (
        day INTEGER NOT NULL,
        graffiti BYTEA NOT NULL,
        graffiti_text TEXT NOT NULL,
        count INTEGER NOT NULL,
        proposer_count INTEGER NOT NULL,
        PRIMARY KEY (graffiti, day)
    );
CREATE INDEX IF NOT EXISTS idx_graffiti_stats_day ON graffiti_stats (day);
CREATE INDEX IF NOT EXISTS idx_graffiti_stats_day_graffiti_text ON graffiti_stats USING gin (graffiti_text gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove ens lookup tables';
DROP INDEX IF EXISTS idx_graffiti_stats_day;
DROP INDEX IF EXISTS idx_graffiti_stats_day_graffiti_text;
DROP TABLE IF EXISTS graffiti_stats;
-- +goose StatementEnd
