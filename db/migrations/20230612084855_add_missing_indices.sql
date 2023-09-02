-- +goose NO TRANSACTION
-- +goose Up

SELECT 'create index for impact level 4 or higher';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_node_jobs_type_status ON public.node_jobs USING btree (type, status);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validators_activationepoch_status ON public.validators USING btree (activationepoch, status);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_block_root_validatorindex ON public.blocks_bls_change USING btree (block_root, validatorindex);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_from_address_publickey ON public.eth1_deposits USING btree (from_address, publickey);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_proposer ON public.blocks USING btree (status, proposer);
-- +goose StatementEnd

SELECT 'create index for impact level 3';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_exec_block_hash_slot ON public.blocks USING btree (exec_block_hash, slot);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_deposits_block_root_publickey ON public.blocks_deposits USING btree (block_root, publickey);
-- +goose StatementEnd

SELECT 'create index for impact level 2';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address_block_root_withdrawalindex ON public.blocks_withdrawals USING btree (address, block_root, withdrawalindex);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_block_root_validatorindex ON public.blocks_withdrawals USING btree (block_root, validatorindex);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_epoch_status_slot ON public.blocks USING btree (epoch, status, slot);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_exec_block_number ON public.blocks USING btree (status, exec_block_number);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_proposer_slot ON public.blocks USING btree (proposer, slot);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_epoch_exec_block_number ON public.blocks USING btree (status, epoch, exec_block_number);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_slot_epoch ON public.blocks USING btree (status, slot, epoch);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rocketpool_minipools_address ON public.rocketpool_minipools USING btree (address);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_blockroot_epoch ON public.blocks USING btree (status, blockroot, epoch);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_graffitiwall_slot_x ON public.graffitiwall USING btree (slot, x);
-- +goose StatementEnd

SELECT 'create index for impact level 1';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_exec_fee_recipient_exec_block_number ON public.blocks USING btree (exec_fee_recipient, exec_block_number);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_proposer_epoch_exec_block_number ON public.blocks USING btree (proposer, epoch, exec_block_number);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_graffiti ON public.blocks USING btree (graffiti);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relays_blocks_tag_id ON public.relays_blocks USING btree (tag_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_indicator_time ON public.chart_series USING btree (indicator, "time");
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_indicator ON public.chart_series USING btree (indicator);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_block_root_block_slot ON public.blocks_withdrawals USING btree (block_root, block_slot);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_blockroot_slot ON public.blocks USING btree (status, blockroot, slot);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_time ON public.chart_series USING btree ("time");
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_period ON public.sync_committees USING btree (period);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_validatorindex ON public.sync_committees USING btree (validatorindex);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ad_configurations_template_id_for_all_users_enabled ON public.ad_configurations USING btree (template_id, for_all_users, enabled);
-- +goose StatementEnd

-- +goose Down
SELECT 'drop index for impact level 4 or higher';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_node_jobs_type_status;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_validators_activationepoch_status;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_bls_change_block_root_validatorindex;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_eth1_deposits_from_address_publickey;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_proposer;
-- +goose StatementEnd

SELECT 'drop index for impact level 3';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_exec_block_hash_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_deposits_block_root_publickey;
-- +goose StatementEnd

SELECT 'drop index for impact level 2';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_address_block_root_withdrawalindex;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_block_root_validatorindex;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_epoch_status_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_exec_block_number;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_proposer_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_epoch_exec_block_number;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_slot_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_rocketpool_minipools_address;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_blockroot_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_graffitiwall_slot_x;
-- +goose StatementEnd

SELECT 'drop index for impact level 1';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_exec_fee_recipient_exec_block_number;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_proposer_epoch_exec_block_number;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_graffiti;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_relays_blocks_tag_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_chart_series_indicator_time;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_chart_series_indicator;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_block_root_block_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_blocks_status_blockroot_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_chart_series_time;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_sync_committees_period;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_sync_committees_validatorindex;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY idx_ad_configurations_template_id_for_all_users_enabled;
-- +goose StatementEnd
