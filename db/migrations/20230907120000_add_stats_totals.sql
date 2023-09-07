-- +goose Up

-- +goose StatementBegin
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS missed_attestations_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS orphaned_attestations_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS proposed_blocks_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS missed_blocks_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS orphaned_blocks_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS attester_slashings_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS proposer_slashings_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS participated_sync_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS missed_sync_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS orphaned_sync_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS withdrawals_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS withdrawals_amount_total BIGINT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS el_rewards_wei_total BIGINT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS cl_rewards_gwei_total BIGINT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS mev_rewards_wei_total BIGINT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS deposits_total INT NOT NULL DEFAULT 0;
ALTER TABLE validator_stats ADD COLUMN IF NOT EXISTS deposits_amount_total BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS
    validator_stats_totals (
        validatorindex INT NOT NULL DEFAULT 0,
        missed_attestations INT NOT NULL DEFAULT 0,
        orphaned_attestations INT NOT NULL DEFAULT 0,
        proposed_blocks INT NOT NULL DEFAULT 0,
        missed_blocks INT NOT NULL DEFAULT 0,
        orphaned_blocks INT NOT NULL DEFAULT 0,
        attester_slashings INT NOT NULL DEFAULT 0,
        proposer_slashings INT NOT NULL DEFAULT 0,
        participated_sync INT NOT NULL DEFAULT 0,
        missed_sync INT NOT NULL DEFAULT 0,
        orphaned_sync INT NOT NULL DEFAULT 0,
        withdrawals INT NOT NULL DEFAULT 0,
        withdrawals_amount BIGINT NOT NULL DEFAULT 0,
        el_rewards_wei NUMERIC NOT NULL DEFAULT 0,
        cl_rewards_gwei BIGINT NOT NULL DEFAULT 0,
        mev_rewards_wei NUMERIC NOT NULL DEFAULT 0,
        deposits INT NOT NULL DEFAULT 0,
        deposits_amount BIGINT NOT NULL DEFAULT 0,
        PRIMARY KEY (validatorindex)
    );
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
-- existed before: missed_attestations_total, participated_sync_total, missed_sync_total, orphaned_sync_total, el_rewards_wei_total, cl_rewards_gwei_total, mev_rewards_wei_total
DROP TABLE IF EXISTS validator_stats_totals;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS missed_attestations_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS orphaned_attestations_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS proposed_blocks_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS missed_blocks_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS orphaned_blocks_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS attester_slashings_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS proposer_slashings_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS participated_sync_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS missed_sync_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS orphaned_sync_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS withdrawals_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS withdrawals_amount_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS el_rewards_wei_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS cl_rewards_gwei_total;
--ALTER TABLE validator_stats DROP COLUMN IF EXISTS mev_rewards_wei_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS deposits_total;
ALTER TABLE validator_stats DROP COLUMN IF EXISTS deposits_amount_total;
-- +goose StatementEnd
