-- +goose Up
-- +goose StatementBegin
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
SELECT 'down SQL query - drop view app_subs_view';
DROP VIEW app_subs_view;
-- +goose StatementEnd
