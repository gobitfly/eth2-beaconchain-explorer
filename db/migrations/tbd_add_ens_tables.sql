-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add ens lookup tables';
CREATE TABLE IF NOT EXISTS
    ens (
        name_hash bytea NOT NULL,
        ens_name TEXT NOT NULL,
        address bytea,
        is_primary_name BOOLEAN NOT NULL DEFAULT FALSE,
        valid_to TIMESTAMP WITHOUT TIME ZONE,
        PRIMARY KEY (name_hash)
    );
CREATE INDEX IF NOT EXISTS idx_ens_name ON ens (ens_name);
CREATE INDEX IF NOT EXISTS idx_ens_valid_name ON ens (ens_name, valid_to);
CREATE INDEX IF NOT EXISTS idx_ens_valid_address_primary ON ens (address, valid_to, is_primary_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove ens lookup tables';
DROP INDEX IF EXISTS idx_ens_name;
DROP INDEX IF EXISTS idx_ens_valid_name;
DROP INDEX IF EXISTS idx_ens_valid_address_primary;
DROP TABLE IF EXISTS ens CASCADE;
-- +goose StatementEnd
