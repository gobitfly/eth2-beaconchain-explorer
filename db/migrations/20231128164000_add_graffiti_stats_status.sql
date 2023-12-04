-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add table graffiti_stats_status';
CREATE TABLE IF NOT EXISTS
    graffiti_stats_status (
        day INT NOT NULL,
        status BOOLEAN NOT NULL,
        PRIMARY KEY (DAY)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop table graffiti_stats_status';
DROP TABLE IF EXISTS graffiti_stats_status;
-- +goose StatementEnd
