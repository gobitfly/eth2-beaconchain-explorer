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

SELECT 'up SQL query - add table api_keys';
CREATE TABLE IF NOT EXISTS
    api_keys (
        user_id INT NOT NULL,
        api_key VARCHAR(256) NOT NULL,
        valid_until TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        changed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        PRIMARY KEY (user_id, api_key)
    );

SELECT 'up SQL query - add table api_weights';
CREATE TABLE IF NOT EXISTS
    api_weights (
        bucket VARCHAR(20) NOT NULL,
        endpoint TEXT NOT NULL,
        method TEXT NOT NULL,
        params TEXT NOT NULL,
        weight INT NOT NULL DEFAULT 0,
        valid_from TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT TO_TIMESTAMP(0),
        PRIMARY KEY (endpoint, valid_from)
    );

SELECT 'up SQL query - add table api_products';
CREATE TABLE IF NOT EXISTS
    api_products (
        name VARCHAR(20) NOT NULL,
        stripe_price_id VARCHAR(256) NOT NULL,
        second INT NOT NULL DEFAULT 0,
        hour INT NOT NULL DEFAULT 0,
        month INT NOT NULL DEFAULT 0,
        valid_from TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT TO_TIMESTAMP(0),
        PRIMARY KEY (name, valid_from)
    ); 
INSERT INTO api_products (name, stripe_price_id, second, hour, month) VALUES
    ('free'    , 'price_free'    ,  5, 0,     30000),
    ('sapphire', 'price_sapphire', 10, 0,    500000),
    ('emerald' , 'price_emerald' , 10, 0,   1000000),
    ('diamond' , 'price_diamond' , 30, 0,   6000000),
    ('custom2' , 'price_custom2' , 50, 0,  13000000),
    ('custom1' , 'price_custom1' , 50, 0, 500000000),
    ('whale'   , 'price_whale'   , 25, 0,    700000),
    ('goldfish', 'price_goldfish', 20, 0,    200000),
    ('plankton', 'price_plankton', 20, 0,    120000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop table api_ratelimits';
DROP TABLE IF EXISTS api_ratelimits;
SELECT 'down SQL query - drop table api_keys';
DROP TABLE IF EXISTS api_keys;
SELECT 'down SQL query - drop table api_weights';
DROP TABLE IF EXISTS api_weights;
-- +goose StatementEnd
