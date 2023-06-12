-- +goose Up
-- +goose StatementBegin
SELECT 'create index for impact level 4 or higher';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_node_jobs_type_status ON public.node_jobs USING btree (type, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validators_activationepoch_status ON public.validators USING btree (activationepoch, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_block_root_validatorindex ON public.blocks_bls_change USING btree (block_root, validatorindex);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_eth1_deposits_from_address_publickey ON public.eth1_deposits USING btree (from_address, publickey);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_proposer ON public.blocks USING btree (status, proposer);

SELECT 'create index for impact level 3';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_exec_block_hash_slot ON public.blocks USING btree (exec_block_hash, slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_deposits_block_root_publickey ON public.blocks_deposits USING btree (block_root, publickey);

SELECT 'create index for impact level 2';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address_block_root_withdrawalindex ON public.blocks_withdrawals USING btree (address, block_root, withdrawalindex);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_block_root_validatorindex ON public.blocks_withdrawals USING btree (block_root, validatorindex);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_epoch_status_slot ON public.blocks USING btree (epoch, status, slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_exec_block_number ON public.blocks USING btree (status, exec_block_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_proposer_slot ON public.blocks USING btree (proposer, slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_epoch_exec_block_number ON public.blocks USING btree (status, epoch, exec_block_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_slot_epoch ON public.blocks USING btree (status, slot, epoch);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rocketpool_minipools_address ON public.rocketpool_minipools USING btree (address);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_blockroot_epoch ON public.blocks USING btree (status, blockroot, epoch);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_graffitiwall_slot_x ON public.graffitiwall USING btree (slot, x);

SELECT 'create index for impact level 1';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_exec_fee_recipient_exec_block_number ON public.blocks USING btree (exec_fee_recipient, exec_block_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_proposer_epoch_exec_block_number ON public.blocks USING btree (proposer, epoch, exec_block_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_graffiti ON public.blocks USING btree (graffiti);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relays_blocks_tag_id ON public.relays_blocks USING btree (tag_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_indicator_time ON public.chart_series USING btree (indicator, "time");
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_indicator ON public.chart_series USING btree (indicator);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_block_root_block_slot ON public.blocks_withdrawals USING btree (block_root, block_slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_status_blockroot_slot ON public.blocks USING btree (status, blockroot, slot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chart_series_time ON public.chart_series USING btree ("time");
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_period ON public.sync_committees USING btree (period);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sync_committees_validatorindex ON public.sync_committees USING btree (validatorindex);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ad_configurations_template_id_for_all_users_enabled ON public.ad_configurations USING btree (template_id, for_all_users, enabled);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'drop index for impact level 4 or higher';
DROP INDEX CONCURRENTLY idx_node_jobs_type_status;
DROP INDEX CONCURRENTLY idx_validators_activationepoch_status;
DROP INDEX CONCURRENTLY idx_blocks_bls_change_block_root_validatorindex;
DROP INDEX CONCURRENTLY idx_eth1_deposits_from_address_publickey;
DROP INDEX CONCURRENTLY idx_blocks_status_proposer;

SELECT 'drop index for impact level 3';
DROP INDEX CONCURRENTLY idx_blocks_exec_block_hash_slot;
DROP INDEX CONCURRENTLY idx_blocks_deposits_block_root_publickey;

SELECT 'drop index for impact level 2';
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_address_block_root_withdrawalindex;
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_block_root_validatorindex;
DROP INDEX CONCURRENTLY idx_blocks_epoch_status_slot;
DROP INDEX CONCURRENTLY idx_blocks_status_exec_block_number;
DROP INDEX CONCURRENTLY idx_blocks_proposer_slot;
DROP INDEX CONCURRENTLY idx_blocks_status_epoch_exec_block_number;
DROP INDEX CONCURRENTLY idx_blocks_status_slot_epoch;
DROP INDEX CONCURRENTLY idx_rocketpool_minipools_address;
DROP INDEX CONCURRENTLY idx_blocks_status_blockroot_epoch;
DROP INDEX CONCURRENTLY idx_graffitiwall_slot_x;

SELECT 'drop index for impact level 1';
DROP INDEX CONCURRENTLY idx_blocks_exec_fee_recipient_exec_block_number;
DROP INDEX CONCURRENTLY idx_blocks_proposer_epoch_exec_block_number;
DROP INDEX CONCURRENTLY idx_blocks_graffiti;
DROP INDEX CONCURRENTLY idx_relays_blocks_tag_id;
DROP INDEX CONCURRENTLY idx_chart_series_indicator_time;
DROP INDEX CONCURRENTLY idx_chart_series_indicator;
DROP INDEX CONCURRENTLY idx_blocks_withdrawals_block_root_block_slot;
DROP INDEX CONCURRENTLY idx_blocks_status_blockroot_slot;
DROP INDEX CONCURRENTLY idx_chart_series_time;
DROP INDEX CONCURRENTLY idx_sync_committees_period;
DROP INDEX CONCURRENTLY idx_sync_committees_validatorindex;
DROP INDEX CONCURRENTLY idx_ad_configurations_template_id_for_all_users_enabled;
-- +goose StatementEnd
