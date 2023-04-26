-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE graffitiwall DROP CONSTRAINT graffitiwall_pkey, ADD PRIMARY KEY(slot);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- remove overwritten pixels
DELETE FROM graffitiwall AS todel
    WHERE EXISTS (
        SELECT 1
        FROM graffitiwall AS newer
        WHERE newer.x = todel.x AND newer.y = todel.y
        AND newer.slot > todel.slot
    );
ALTER TABLE graffitiwall DROP CONSTRAINT graffitiwall_pkey, ADD PRIMARY KEY(x, y);
-- +goose StatementEnd
