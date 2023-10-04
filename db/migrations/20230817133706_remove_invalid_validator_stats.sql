-- +goose NO TRANSACTION
-- +goose Up
SELECT 'up SQL query - remove invalid validator stats';
-- +goose StatementBegin
delete from validator_performance where validatorindex = 2147483647;
delete from validator_stats where validatorindex = 2147483647;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd