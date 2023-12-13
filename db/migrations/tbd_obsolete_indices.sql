-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - drop obsolete indices';
DROP INDEX CONCURRENTLY IF EXISTS idx_validators_lastattestationslot;
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_attestations_source_root;
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_attestations_target_root;
DROP INDEX CONCURRENTLY IF EXISTS idx_proposal_assignments_epoch;
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_bls_change_address;
DROP INDEX CONCURRENTLY IF EXISTS idx_rocketpool_dao_proposals_member_votes_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_validator_performance_balance;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - restore indices';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validators_lastattestationslot ON validators (lastattestationslot);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_attestations_source_root ON blocks_attestations (source_root);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_attestations_target_root ON blocks_attestations (target_root);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_proposal_assignments_epoch ON proposal_assignments (epoch);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_bls_change_address ON blocks_bls_change (address);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rocketpool_dao_proposals_member_votes_id ON rocketpool_dao_proposals_member_votes (id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validator_performance_balance ON validator_performance (balance);
-- +goose StatementEnd
