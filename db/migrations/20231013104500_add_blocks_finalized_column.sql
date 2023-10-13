-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add block finalized column'; 
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS finalized bool not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop block finalized column'; 
ALTER TABLE blocks DROP COLUMN finalized;
-- +goose StatementEnd
