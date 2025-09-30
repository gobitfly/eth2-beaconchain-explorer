-- +goose Up
-- +goose StatementBegin
-- Validator entities master tables
CREATE TABLE IF NOT EXISTS validator_entities (
    publickey BYTEA PRIMARY KEY,
    entity TEXT,
    sub_entity TEXT
);

CREATE INDEX IF NOT EXISTS idx_validator_entities_entity ON validator_entities (entity);
CREATE INDEX IF NOT EXISTS idx_validator_entities_sub_entity ON validator_entities (sub_entity);

CREATE TABLE IF NOT EXISTS validator_entities_data_periods (
    entity TEXT NOT NULL,
    sub_entity TEXT NOT NULL,
    period TEXT NOT NULL,
    last_updated_at TIMESTAMPTZ NOT NULL,

    balance_end_sum_gwei BIGINT NOT NULL,

    efficiency_dividend BIGINT NOT NULL,
    efficiency_divisor BIGINT NOT NULL,
    efficiency DOUBLE PRECISION NOT NULL,
    roi_dividend NUMERIC NOT NULL,
    roi_divisor NUMERIC NOT NULL,
    attestation_efficiency DOUBLE PRECISION NOT NULL,
    proposal_efficiency DOUBLE PRECISION NOT NULL,
    sync_committee_efficiency DOUBLE PRECISION NOT NULL,

    attestations_scheduled_sum BIGINT NOT NULL,
    attestations_observed_sum BIGINT NOT NULL,
    attestations_head_executed_sum BIGINT NOT NULL,
    attestations_source_executed_sum BIGINT NOT NULL,
    attestations_target_executed_sum BIGINT NOT NULL,
    attestations_missed_rewards_sum BIGINT NOT NULL,
    attestations_reward_rewards_only_sum BIGINT NOT NULL,

    blocks_scheduled_sum BIGINT NOT NULL,
    blocks_proposed_sum BIGINT NOT NULL,

    sync_scheduled_sum BIGINT NOT NULL,
    sync_executed_sum BIGINT NOT NULL,

    slashed_in_period_max BIGINT NOT NULL,
    slashed_amount_sum BIGINT NOT NULL,

    blocks_cl_missed_median_reward_sum BIGINT NOT NULL,
    sync_localized_max_reward_sum BIGINT NOT NULL,
    sync_reward_rewards_only_sum BIGINT NOT NULL,
    inclusion_delay_sum BIGINT NOT NULL,

    -- total execution layer rewards over range (wei)
    execution_rewards_sum_wei NUMERIC NOT NULL DEFAULT 0,

    efficiency_time_bucket_timestamps_sec BIGINT[] NOT NULL,
    efficiency_time_bucket_values DOUBLE PRECISION[] NOT NULL,

    net_share DOUBLE PRECISION NOT NULL,
    status_counts JSONB NOT NULL,

    PRIMARY KEY (entity, sub_entity, period)
);

CREATE INDEX IF NOT EXISTS idx_validator_entities_data_periods_entity ON validator_entities_data_periods (entity);
CREATE INDEX IF NOT EXISTS idx_validator_entities_data_periods_sub_entity ON validator_entities_data_periods (sub_entity);
CREATE INDEX IF NOT EXISTS idx_validator_entities_data_periods_period ON validator_entities_data_periods (period);

-- Lido: Node operators metadata
CREATE TABLE IF NOT EXISTS lido_node_operators (
    operator_id BIGINT PRIMARY KEY,
    active BOOLEAN NOT NULL,
    name TEXT NOT NULL,
    reward_address BYTEA NOT NULL,
    total_vetted_validators BIGINT NOT NULL,
    total_exited_validators BIGINT NOT NULL,
    total_added_validators BIGINT NOT NULL,
    total_deposited_validators BIGINT NOT NULL,
    signing_key_count BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_lido_node_operators_active ON lido_node_operators (active);
CREATE INDEX IF NOT EXISTS idx_lido_node_operators_name ON lido_node_operators (name);
CREATE INDEX IF NOT EXISTS idx_lido_node_operators_reward_address ON lido_node_operators (reward_address);

-- Lido: Signing keys per operator
CREATE TABLE IF NOT EXISTS lido_signing_keys (
    operator_id BIGINT NOT NULL REFERENCES lido_node_operators(operator_id) ON DELETE CASCADE,
    pubkey BYTEA NOT NULL,
    PRIMARY KEY (operator_id, pubkey)
);
CREATE INDEX IF NOT EXISTS idx_lido_signing_keys_operator_id ON lido_signing_keys (operator_id);
CREATE INDEX IF NOT EXISTS idx_lido_signing_keys_pubkey ON lido_signing_keys (pubkey);

-- Lido CSM: Node operators (minimal fields)
CREATE TABLE IF NOT EXISTS lido_csm_node_operators (
    operator_id BIGINT PRIMARY KEY,
    signing_key_count BIGINT NOT NULL DEFAULT 0
);

-- Lido CSM: Signing keys per operator
CREATE TABLE IF NOT EXISTS lido_csm_signing_keys (
    operator_id BIGINT NOT NULL REFERENCES lido_csm_node_operators(operator_id) ON DELETE CASCADE,
    pubkey BYTEA NOT NULL,
    PRIMARY KEY (operator_id, pubkey)
);
CREATE INDEX IF NOT EXISTS idx_lido_csm_signing_keys_operator_id ON lido_csm_signing_keys (operator_id);
CREATE INDEX IF NOT EXISTS idx_lido_csm_signing_keys_pubkey ON lido_csm_signing_keys (pubkey);

-- Validator tagger job runs
CREATE TABLE IF NOT EXISTS validator_tagger_job_runs (
    id BIGSERIAL PRIMARY KEY,
    job_name TEXT NOT NULL,
    run_group_id UUID NULL,
    scheduled_at_utc TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    status TEXT NOT NULL DEFAULT 'scheduled', -- scheduled|running|ok|error|skipped
    error_text TEXT NULL,
    triggered_by TEXT NOT NULL DEFAULT 'scheduler'
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_validator_tagger_job_runs_jobname_scheduled
    ON validator_tagger_job_runs (job_name, scheduled_at_utc);
CREATE INDEX IF NOT EXISTS idx_validator_tagger_job_runs_status
    ON validator_tagger_job_runs (status);
CREATE INDEX IF NOT EXISTS idx_validator_tagger_job_runs_started_at
    ON validator_tagger_job_runs (started_at);
CREATE INDEX IF NOT EXISTS idx_validator_tagger_job_runs_run_group
    ON validator_tagger_job_runs (run_group_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_tagger_job_runs;
DROP TABLE IF EXISTS lido_csm_signing_keys;
DROP TABLE IF EXISTS lido_csm_node_operators;
DROP TABLE IF EXISTS lido_signing_keys;
DROP TABLE IF EXISTS lido_node_operators;
DROP TABLE IF EXISTS validator_entities_data_periods;
DROP TABLE IF EXISTS validator_entities;
-- +goose StatementEnd
