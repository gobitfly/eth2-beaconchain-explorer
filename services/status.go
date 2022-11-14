package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/version"
	"os"
	"time"
)

// Report the status of a particular service, will add current Pid and executable name
func ReportStatus(name, status string, metadata *json.RawMessage) {
	pid := os.Getpid()
	execName, err := os.Executable()
	if err != nil {
		execName = "Unknown"
	}

	version := version.Version

	_, err = db.WriterDb.Exec(`
		INSERT INTO service_status (name, executable_name, version, pid, status, metadata, last_update) VALUES ($1, $2, $3, $4, $5, $6, $7) 
		ON CONFLICT (name, executable_name, version, pid) DO UPDATE SET
		status = excluded.status,
		metadata = excluded.metadata,
		last_update = excluded.last_update
	`, name, execName, version, pid, status, metadata, time.Now())

	if err != nil {
		logger.Errorf("error reporting service status: %v", err)
	}
}
