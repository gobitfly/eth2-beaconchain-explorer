-- +goose Up
-- +goose StatementBegin
ALTER TABLE eth1_deposits ALTER COLUMN tx_input DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE eth1_deposits ALTER COLUMN tx_input SET NOT NULL;

-- +goose StatementEnd
