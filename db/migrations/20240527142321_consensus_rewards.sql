-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - create consensus_payloads table');
CREATE TABLE consensus_payloads (
	slot BIGINT,
	cl_attestations_reward BIGINT, -- gwei
    cl_sync_aggregate_reward BIGINT, -- gwei
    cl_slashing_inclusion_reward BIGINT, -- gwei
	CONSTRAINT consensus_payloads_pk PRIMARY KEY (slot)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - drop consensus_payloads table');
DROP TABLE consensus_payloads;
-- +goose StatementEnd
