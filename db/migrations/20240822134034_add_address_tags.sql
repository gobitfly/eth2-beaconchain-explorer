-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - create address_names table');
CREATE TABLE IF NOT EXISTS address_names (
    address bytea NOT NULL UNIQUE,
    name TEXT NOT NULL,
    PRIMARY KEY (address, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - drop address_names table');
DROP TABLE IF EXISTS address_names;
-- +goose StatementEnd
