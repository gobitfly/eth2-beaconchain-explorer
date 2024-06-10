-- +goose NO TRANSACTION
-- +goose Up

SELECT 'create index missing indices';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_deposits_block_slot_block_root ON public.blocks_deposits USING btree (block_slot, block_root);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rocketpool_dao_proposals_member_votes_id ON public.rocketpool_dao_proposals_member_votes USING btree (id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validators_activationeligibilityepoch ON public.validators USING btree (activationeligibilityepoch);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validator_attestation_streaks_status_longest ON public.validator_attestation_streaks USING btree (status, longest);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validator_attestation_streaks_status_current ON public.validator_attestation_streaks USING btree (status, current);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_validatorindex_period ON public.sync_committees USING btree (validatorindex, period DESC);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_sync_committees_period;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_period ON public.sync_committees USING btree (period);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_validators_tags_tag_user_id ON public.users_validators_tags USING btree (tag, user_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_service_status_name_last_update ON public.service_status USING btree (name, last_update);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_proposer_exec_block_number ON public.blocks USING btree (status, proposer, exec_block_number);
-- +goose StatementEnd

-- +goose Down
SELECT 'drop index missing indices';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_deposits_block_slot_block_root;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_rocketpool_dao_proposals_member_votes_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_validators_activationeligibilityepoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_validator_attestation_streaks_status_longest;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_validator_attestation_streaks_status_current;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_sync_committees_validatorindex_period;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_sync_committees_period;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_period ON public.sync_committees USING btree (validatorindex, period DESC);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_users_validators_tags_tag_user_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_service_status_name_last_update;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_proposer_exec_block_number;
-- +goose StatementEnd
