-- +goose Up
-- +goose StatementBegin
-- Lido Simple DVT: Node operators (minimal fields)
CREATE TABLE IF NOT EXISTS lido_simple_dvt_node_operators (
    operator_id BIGINT PRIMARY KEY,
    signing_key_count BIGINT NOT NULL DEFAULT 0,
    name text not null default ''
);

-- Lido CSM: Signing keys per operator
CREATE TABLE IF NOT EXISTS lido_simple_dvt_signing_keys (
     operator_id BIGINT NOT NULL REFERENCES lido_simple_dvt_node_operators(operator_id) ON DELETE CASCADE,
     pubkey BYTEA NOT NULL,
     PRIMARY KEY (operator_id, pubkey)
);
CREATE INDEX IF NOT EXISTS idx_lido_simple_dvt_signing_keys_operator_id ON lido_simple_dvt_signing_keys (operator_id);
CREATE INDEX IF NOT EXISTS idx_lido_simple_dvt_signing_keys_pubkey ON lido_simple_dvt_signing_keys (pubkey);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lido_simple_dvt_signing_keys;
DROP TABLE IF EXISTS lido_simple_dvt_node_operators;
-- +goose StatementEnd
