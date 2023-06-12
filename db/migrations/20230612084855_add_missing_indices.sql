-- +goose Up
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY ON public.node_jobs USING btree (type, status);
CREATE INDEX CONCURRENTLY ON public.validators USING btree (activationepoch, status);
CREATE INDEX CONCURRENTLY ON public.blocks_bls_change USING btree (block_root, validatorindex);
CREATE INDEX CONCURRENTLY ON public.eth1_deposits USING btree (from_address, publickey);
CREATE INDEX CONCURRENTLY ON public.blocks USING btree (status, proposer);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX CONCURRENTLY node_jobs_type_status_idx;
DROP INDEX CONCURRENTLY validators_activationepoch_status_idx;
DROP INDEX CONCURRENTLY blocks_bls_change_block_root_validatorindex_idx;
DROP INDEX CONCURRENTLY eth1_deposits_from_address_publickey_idx;
DROP INDEX CONCURRENTLY blocks_status_proposer_idx;
-- +goose StatementEnd
