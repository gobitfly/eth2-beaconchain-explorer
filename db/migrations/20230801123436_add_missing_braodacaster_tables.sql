-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add table node_jobs_bls_changes_validators';
CREATE TABLE IF NOT EXISTS
    node_jobs_bls_changes_validators (
        validatorindex INT NOT NULL,
        node_job_id VARCHAR(40) NOT NULL,
        PRIMARY KEY (validatorindex)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove table node_jobs_bls_changes_validators';
DROP TABLE IF EXISTS node_jobs_bls_changes_validators CASCADE;
-- +goose StatementEnd
