package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/sirupsen/logrus"
)

var validatorSchedulerLogger = logrus.StandardLogger().WithField("module", "services.validator_tagger_scheduler")

// JobStatus represents the status stored in validator_tagger_job_runs.status
const (
	JobStatusScheduled = "scheduled"
	JobStatusRunning   = "running"
	JobStatusOK        = "ok"
	JobStatusError     = "error"
	JobStatusSkipped   = "skipped"
)

// JobRun represents a row in validator_tagger_job_runs.
type JobRun struct {
	ID             int64
	JobName        string
	RunGroupID     *string
	ScheduledAtUTC time.Time
	StartedAt      sql.NullTime
	FinishedAt     sql.NullTime
	Status         string
	ErrorText      sql.NullString
	TriggeredBy    string
}

// ensureJobRow ensures a scheduled row exists for the given job and time.
// If it already exists, it does nothing. It never returns an error if the row already exists.
func ensureJobRow(ctx context.Context, jobName string, scheduledAt time.Time, runGroupID *string, triggeredBy string) error {
	_, err := db.WriterDb.Exec(
		`INSERT INTO validator_tagger_job_runs(job_name, run_group_id, scheduled_at_utc, status, triggered_by)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT DO NOTHING`,
		jobName, runGroupID, scheduledAt.UTC(), JobStatusScheduled, triggeredBy,
	)
	if err != nil {
		return fmt.Errorf("ensureJobRow insert: %w", err)
	}
	return nil
}

// tryStartJob attempts to transition a job row from scheduled->running and returns the id if successful.
func tryStartJob(ctx context.Context, jobName string, scheduledAt time.Time) (int64, bool, error) {
	var id int64
	err := db.WriterDb.QueryRow(
		`UPDATE validator_tagger_job_runs
		  SET status = $1, started_at = NOW()
		 WHERE job_name = $2 AND scheduled_at_utc = $3 AND status = $4
		 RETURNING id`,
		JobStatusRunning, jobName, scheduledAt.UTC(), JobStatusScheduled,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("tryStartJob update: %w", err)
	}
	return id, true, nil
}

// finishJob sets final status and error (optional).
func finishJob(ctx context.Context, id int64, finalStatus string, errText *string) error {
	_, err := db.WriterDb.Exec(
		`UPDATE validator_tagger_job_runs
		    SET status = $1, finished_at = NOW(), error_text = $2
		  WHERE id = $3`,
		finalStatus, errText, id,
	)
	if err != nil {
		return fmt.Errorf("finishJob update: %w", err)
	}
	return nil
}
