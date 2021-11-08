package exporter

import "time"

func syncCommitteesExporter() {
	for {
		exportSyncCommittee()
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommittee() error {
	return nil
}
