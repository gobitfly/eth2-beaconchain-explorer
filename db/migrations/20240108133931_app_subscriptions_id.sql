-- +goose Up
-- +goose StatementBegin
ALTER TABLE users_app_subscriptions ADD COLUMN legacy_receipt varchar(150000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users_app_subscriptions DROP COLUMN legacy_receipt;
-- +goose StatementEnd
