-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - update api_products');
ALTER TABLE api_products ADD COLUMN IF NOT EXISTS bucket TEXT NOT NULL DEFAULT 'default';
ALTER TABLE api_products DROP CONSTRAINT api_products_pkey, ADD PRIMARY KEY (name, bucket, valid_from);
SELECT('up SQL query - update api_ratelimits');
ALTER TABLE api_ratelimits ADD COLUMN IF NOT EXISTS bucket TEXT NOT NULL DEFAULT 'default';
ALTER TABLE api_ratelimits DROP CONSTRAINT api_ratelimits_pkey, ADD PRIMARY KEY (user_id, bucket);
SELECT('up SQL query - update api_statistics');
ALTER TABLE api_statistics ADD COLUMN IF NOT EXISTS bucket TEXT NOT NULL DEFAULT 'default';
ALTER TABLE api_statistics DROP CONSTRAINT api_statistics_pkey, ADD PRIMARY KEY (ts, apikey, endpoint, bucket);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - update api_products');
ALTER TABLE api_products DROP COLUMN IF EXISTS bucket;
ALTER TABLE api_products DROP CONSTRAINT api_products_pkey, ADD PRIMARY KEY (name, valid_from);
SELECT('down SQL query - update api_ratelimits');
ALTER TABLE api_ratelimits DROP COLUMN IF EXISTS bucket;
ALTER TABLE api_ratelimits DROP CONSTRAINT api_ratelimits_pkey, ADD PRIMARY KEY (user_id);
SELECT('down SQL query - update api_statistics');
ALTER TABLE api_statistics DROP COLUMN IF EXISTS bucket;
ALTER TABLE api_statistics DROP CONSTRAINT api_statistics_pkey, ADD PRIMARY KEY (ts, apikey, endpoint);
-- +goose StatementEnd
