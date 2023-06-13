-- +goose Up
-- +goose StatementBegin
SELECT 'add column activationeligibilityepoch to validator_queue_deposits';
ALTER TABLE validator_queue_deposits ADD COLUMN IF NOT EXISTS activationeligibilityepoch BIGINT;

SELECT 'populate activationeligibilityepoch data on validator_queue_deposits from validators table';
UPDATE validator_queue_deposits 
SET 
	activationeligibilityepoch=validators.activationeligibilityepoch
FROM validators
Where validators.validatorindex = validator_queue_deposits.validatorindex;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'drop column activationeligibilityepoch from validator_queue_deposits';
ALTER TABLE validator_queue_deposits DROP COLUMN IF EXISTS activationeligibilityepoch;
-- +goose StatementEnd
