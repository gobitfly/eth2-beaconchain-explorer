-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add block finalized column'; 
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS finalized bool not null default false;
UPDATE blocks SET finalized = true WHERE epoch <= (SELECT COALESCE(MAX(epoch) - 3, 0) FROM epochs WHERE finalized) AND NOT finalized;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop block finalized column'; 
ALTER TABLE blocks DROP COLUMN finalized;
-- +goose StatementEnd
