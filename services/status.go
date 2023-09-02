package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"os"
)

// Report the status of a particular service, will add current Pid and executable name
func ReportStatus(name, status string, metadata *json.RawMessage) {
	if !utils.Config.ReportServiceStatus {
		return
	}
	pid := os.Getpid()
	execName, err := os.Executable()
	if err != nil {
		execName = "Unknown"
	}

	version := version.Version

	_, err = db.WriterDb.Exec(`
		INSERT INTO service_status (name, executable_name, version, pid, status, metadata, last_update) VALUES ($1, $2, $3, $4, $5, $6, NOW()) 
		ON CONFLICT (name, executable_name, version, pid) DO UPDATE SET
		status = excluded.status,
		metadata = excluded.metadata,
		last_update = excluded.last_update
	`, name, execName, version, pid, status, metadata)

	if err != nil {
		utils.LogError(err, "error reporting service status", 0, map[string]interface{}{"name": name, "status": status})
	}
}
