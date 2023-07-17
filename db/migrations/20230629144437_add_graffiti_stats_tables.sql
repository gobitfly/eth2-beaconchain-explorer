-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add graffiti_stats tables';
CREATE TABLE IF NOT EXISTS
    graffiti_stats (
        day INTEGER NOT NULL,
        graffiti BYTEA NOT NULL,
        graffiti_text TEXT NOT NULL,
        count INTEGER NOT NULL,
        proposer_count INTEGER NOT NULL,
        PRIMARY KEY (graffiti, day)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove graffiti_stats tables';
DROP TABLE IF EXISTS graffiti_stats;
-- +goose StatementEnd
