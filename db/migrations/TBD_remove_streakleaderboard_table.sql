-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - remove table validator_attestation_streaks';
DROP TABLE IF EXISTS validator_attestation_streaks CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - add table validator_attestation_streaks and indexes';
CREATE TABLE IF NOT EXISTS
    validator_attestation_streaks (
        validatorindex INT NOT NULL,
        status INT NOT NULL,
        START INT NOT NULL,
        LENGTH INT NOT NULL,
        longest BOOLEAN NOT NULL,
        CURRENT BOOLEAN NOT NULL,
        PRIMARY KEY (validatorindex, status, START)
    );
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_validatorindex ON validator_attestation_streaks (validatorindex);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_status ON validator_attestation_streaks (status);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_length ON validator_attestation_streaks (LENGTH);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_start ON validator_attestation_streaks (START);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validator_attestation_streaks_status_longest ON public.validator_attestation_streaks USING btree (status, longest);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_validator_attestation_streaks_status_current ON public.validator_attestation_streaks USING btree (status, current);
-- +goose StatementEnd
