-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add ens lookup tables';
CREATE TABLE IF NOT EXISTS
    ens_names (
        name_hash bytea NOT NULL,
        name TEXT NOT NULL,
        address bytea,
        valid_to TIMESTAMP WITHOUT TIME ZONE,
        PRIMARY KEY (name_hash)
    );
CREATE INDEX IF NOT EXISTS idx_ens_names_name ON ens_names (name);
CREATE INDEX IF NOT EXISTS idx_ens_names_valid_name ON ens_names (name, valid_to);
CREATE INDEX IF NOT EXISTS idx_ens_names_address ON ens_names (address);
CREATE TABLE IF NOT EXISTS
    ens_addresses (
        address bytea NOT NULL,
        name_hash bytea,
        PRIMARY KEY (address)
    );
CREATE INDEX IF NOT EXISTS idx_ens_addresses_name_hash ON ens_addresses (name_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove ens lookup tables';
DROP INDEX IF EXISTS idx_ens_addresses_name_hash
DROP TABLE IF EXISTS ens_addresses CASCADE;
DROP INDEX IF EXISTS idx_ens_names_name
DROP INDEX IF EXISTS idx_ens_names_valid_name
DROP INDEX IF EXISTS idx_ens_names_address
DROP TABLE IF EXISTS ens_names CASCADE;
-- +goose StatementEnd
