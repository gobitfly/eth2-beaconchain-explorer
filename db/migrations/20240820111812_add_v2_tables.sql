-- +goose Up
-- Validator Dashboard

-- +goose StatementBegin
SELECT 'create users_val_dashboards table';
CREATE TABLE IF NOT EXISTS users_val_dashboards (
    id          BIGSERIAL   NOT NULL,
    user_id     BIGINT      NOT NULL,
    network     SMALLINT    NOT NULL, -- indicate gnosis/eth mainnet and potentially testnets
    name        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMP   DEFAULT(NOW()),
    is_archived TEXT,
    primary key (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_groups table';
CREATE TABLE IF NOT EXISTS users_val_dashboards_groups (
    id           SMALLINT    DEFAULT(0),
    dashboard_id BIGINT      NOT NULL,
    name         VARCHAR(50) NOT NULL,
    foreign key (dashboard_id) references users_val_dashboards(id) ON DELETE CASCADE,
    primary key (dashboard_id, id)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_validators table';
CREATE TABLE IF NOT EXISTS users_val_dashboards_validators ( -- a validator must not be in multiple groups
    dashboard_id    BIGINT   NOT NULL,
    group_id        SMALLINT NOT NULL,
    validator_index BIGINT   NOT NULL,
    foreign key (dashboard_id, group_id) references users_val_dashboards_groups(dashboard_id, id) ON DELETE CASCADE,
    primary key (dashboard_id, validator_index)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_sharing table';
CREATE TABLE IF NOT EXISTS users_val_dashboards_sharing (
    dashboard_id  BIGINT      NOT NULL,
    public_id     CHAR(38)    DEFAULT ('v-' || gen_random_uuid()::text) UNIQUE, -- prefix with "v" for validator dashboards. Public ID to dashboard
    name          VARCHAR(50) NOT NULL,
    shared_groups bool        NOT NULL, -- all groups or default 0
    foreign key (dashboard_id) references users_val_dashboards(id) ON DELETE CASCADE,
    primary key (public_id)
);
-- +goose StatementEnd

-- Account Dashboard

-- +goose StatementBegin
SELECT 'create users_val_dashboards_groups table';
CREATE TABLE IF NOT EXISTS users_acc_dashboards (
    id            BIGSERIAL   NOT NULL,
    user_id       BIGINT      NOT NULL,
    name          VARCHAR(50) NOT NULL,
    user_settings JSONB       DEFAULT '{}'::jsonb, -- or do we want to use a separate kv table for this?
    created_at    TIMESTAMP   DEFAULT(NOW()),
    primary key (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_groups table';
CREATE TABLE IF NOT EXISTS users_acc_dashboards_groups (
    id            INT         NOT NULL,
    dashboard_id  BIGINT      NOT NULL,
    name          VARCHAR(50) NOT NULL,
    foreign key (dashboard_id) references users_acc_dashboards(id) ON DELETE CASCADE,
    primary key (dashboard_id, id)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_groups table';
CREATE TABLE IF NOT EXISTS users_acc_dashboards_accounts ( -- an account must not be in multiple groups
    dashboard_id BIGINT   NOT NULL,
    group_id     SMALLINT NOT NULL,
    address      BYTEA    NOT NULL,
    foreign key (dashboard_id, group_id) references users_acc_dashboards_groups(dashboard_id, id) ON DELETE CASCADE,
    primary key (dashboard_id, address)
);
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'create users_val_dashboards_groups table';
CREATE TABLE IF NOT EXISTS users_acc_dashboards_sharing (
    dashboard_id    BIGINT      NOT NULL,
    public_id       CHAR(38)    DEFAULT('a-' || gen_random_uuid()::text) UNIQUE, -- prefix with "a" for validator dashboards
    name            VARCHAR(50) NOT NULL,
    user_settings   JSONB       DEFAULT '{}'::jsonb, -- snapshots users_dashboards.user_settings at the time of creating the share
    shared_groups   bool        NOT NULL, -- all groups or default 0
    tx_notes_shared BOOLEAN     NOT NULL, -- not snapshoted
    foreign key (dashboard_id) references users_acc_dashboards(id) ON DELETE CASCADE,
    primary key (public_id)
);
-- +goose StatementEnd

-- Notification Dashboard (wip)

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_not_dashboards (
    id         BIGINT,
    user_id    BIGINT,
    name       VARCHAR(50) NOT NULL,
    created_at timestamp,
    primary key (id)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'delete users_val_dashboards_validators table';
DROP TABLE IF EXISTS users_val_dashboards_validators;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_val_dashboards_groups table';
DROP TABLE IF EXISTS users_val_dashboards_groups;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_val_dashboards_sharing table';
DROP TABLE IF EXISTS users_val_dashboards_sharing;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_val_dashboards table';
DROP TABLE IF EXISTS users_val_dashboards;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'delete users_acc_dashboards_accounts table';
DROP TABLE IF EXISTS users_acc_dashboards_accounts;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_acc_dashboards_groups table';
DROP TABLE IF EXISTS users_acc_dashboards_groups;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_acc_dashboards_sharing table';
DROP TABLE IF EXISTS users_acc_dashboards_sharing;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'delete users_acc_dashboards table';
DROP TABLE IF EXISTS users_acc_dashboards;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT 'delete users_not_dashboards table';
DROP TABLE IF EXISTS users_not_dashboards;
-- +goose StatementEnd
