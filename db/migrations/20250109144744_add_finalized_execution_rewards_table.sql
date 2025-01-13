-- +goose Up
-- +goose StatementBegin
CREATE TABLE execution_rewards_finalized (
	epoch int4 NOT NULL,
	slot int4 NOT NULL,
	proposer int4 NOT NULL,
	value numeric NOT NULL,
	CONSTRAINT finalized_execution_rewards_pk PRIMARY KEY (slot)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX finalized_execution_rewards_epoch_idx ON execution_rewards_finalized USING btree (epoch);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE UNIQUE INDEX finalized_execution_rewards_proposer_idx ON execution_rewards_finalized USING btree (proposer, slot);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE execution_rewards_finalized;
-- +goose StatementEnd
