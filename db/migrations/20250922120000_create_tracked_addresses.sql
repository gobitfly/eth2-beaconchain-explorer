-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - create tracked_addresses table');
CREATE TABLE IF NOT EXISTS tracked_addresses (
    address TEXT PRIMARY KEY,
    balance TEXT DEFAULT '0'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - drop tracked_addresses table');
DROP TABLE IF EXISTS tracked_addresses;
-- +goose StatementEnd
