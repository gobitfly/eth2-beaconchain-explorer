-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add table api_ratelimits';
CREATE TABLE IF NOT EXISTS
    api_ratelimits (
        user_id INT NOT NULL,
        second INT NOT NULL DEFAULT 0,
        hour INT NOT NULL DEFAULT 0,
        month INT NOT NULL DEFAULT 0,
        valid_until TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        changed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL, 
        PRIMARY KEY (user_id)
    );

CREATE INDEX IF NOT EXISTS idx_api_ratelimits_changed_at_valid_until ON api_ratelimits (changed_at, valid_until);

SELECT 'up SQL query - add table api_keys';
CREATE TABLE IF NOT EXISTS
    api_keys (
        api_key VARCHAR(256) NOT NULL UNIQUE,
        user_id INT NOT NULL,
        valid_until TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT '9999-12-31 23:59:59',
        changed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (api_key)
    );

CREATE INDEX IF NOT EXISTS idx_api_keys_changed_at_valid_until ON api_keys (changed_at, valid_until);

SELECT 'up SQL query - add table api_weights';
CREATE TABLE IF NOT EXISTS
    api_weights (
        bucket VARCHAR(20) NOT NULL,
        endpoint TEXT NOT NULL,
        method TEXT NOT NULL,
        params TEXT NOT NULL,
        weight INT NOT NULL DEFAULT 1,
        valid_from TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT TO_TIMESTAMP(0),
        PRIMARY KEY (endpoint, valid_from)
    );

SELECT 'up SQL query - add table api_products';
CREATE TABLE IF NOT EXISTS
    api_products (
        name VARCHAR(20) NOT NULL,
        stripe_price_id VARCHAR(256) NOT NULL DEFAULT '',
        second INT NOT NULL DEFAULT 0,
        hour INT NOT NULL DEFAULT 0,
        month INT NOT NULL DEFAULT 0,
        valid_from TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT TO_TIMESTAMP(0),
        PRIMARY KEY (name, valid_from)
    );
INSERT INTO api_products (name, second, hour, month) VALUES
    ('nokey', 2, 1000, 0),
    ('free', 10, 0, 0),
    ('unlimited', 100, 0, 0)
ON CONFLICT DO NOTHING;

ALTER TABLE api_statistics ADD COLUMN IF NOT EXISTS endpoint TEXT NOT NULL DEFAULT '';
ALTER TABLE api_statistics DROP CONSTRAINT IF EXISTS api_statistics_pkey;
ALTER TABLE api_statistics ADD PRIMARY KEY (ts, apikey, endpoint);
ALTER TABLE api_statistics ALTER COLUMN call SET DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop table api_ratelimits';
DROP TABLE IF EXISTS api_ratelimits;
SELECT 'down SQL query - drop index idx_api_ratelimits_changed_at';
DROP INDEX IF EXISTS idx_api_ratelimits_changed_at;
SELECT 'down SQL query - drop table api_keys';
DROP TABLE IF EXISTS api_keys;
SELECT 'down SQL query - drop index idx_api_keys_changed_at';
DROP INDEX IF EXISTS idx_api_keys_changed_at;
SELECT 'down SQL query - drop table api_weights';
DROP TABLE IF EXISTS api_weights;
SELECT 'down SQL query - drop table api_products';
DROP TABLE IF EXISTS api_products;
SELECT 'down SQL query - drop column api_statistics.endpoint';
ALTER TABLE api_statistics DROP COLUMN IF EXISTS endpoint;
ALTER TABLE api_statistics DROP CONSTRAINT IF EXISTS api_statistics_pkey;
ALTER TABLE api_statistics ADD PRIMARY KEY (ts, apikey, call);
-- +goose StatementEnd
