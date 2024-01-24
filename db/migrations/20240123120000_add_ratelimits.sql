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

SELECT 'up SQL query - add view app_subs_view';
CREATE OR REPLACE VIEW app_subs_view AS
    SELECT users_app_subscriptions.id,
        users_app_subscriptions.user_id,
        users_app_subscriptions.product_id,
        users_app_subscriptions.created_at,
        users_app_subscriptions.updated_at,
        users_app_subscriptions.validate_remotely,
        users_app_subscriptions.active,
        users_app_subscriptions.store,
        users_app_subscriptions.expires_at,
        users_app_subscriptions.reject_reason,
        users_app_subscriptions.receipt_hash
    FROM users_app_subscriptions;
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
