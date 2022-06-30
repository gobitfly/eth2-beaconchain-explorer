package services

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func FailedDutiesExporter() {
	logger.Infof("starting services.FailedDutiesExporter")
	for {
		latestEpoch, err := db.GetLatestEpoch()
		if err != nil {
			logger.WithError(err).Errorf("error in services.FailedDutiesExporter: db.GetLatestEpoch")
			time.Sleep(time.Second * 10)
			continue
		}
		var lastExportedFailedEpoch sql.NullInt64
		err = db.ReaderDb.Get(&lastExportedFailedEpoch, `select max(epoch) from validator_failed_duties_status where status`)
		if err != nil {
			logger.WithError(err).Errorf("error in services.FailedDutiesExporter: getting lastExportedFailedEpoch")
			time.Sleep(time.Second * 10)
			continue
		}
		if lastExportedFailedEpoch.Valid && latestEpoch <= uint64(lastExportedFailedEpoch.Int64) {
			time.Sleep(time.Second * 10)
			continue
		}
		err = exportFailedDuties(uint64(lastExportedFailedEpoch.Int64) + 1)
		if err != nil {
			logger.WithError(err).Errorf("error in services.FailedDutiesExporter: exportFailedDuties")
			time.Sleep(time.Second * 10)
			continue
		}

		time.Sleep(time.Second * 2)
	}
}

func exportFailedDuties(epoch uint64) error {
	tx, err := db.WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: creating db-tx: %w", err)
	}
	defer tx.Rollback()

	start := time.Now()
	defer func() {
		if err != nil {
			metrics.TaskDuration.WithLabelValues("services_exportFailedDuties").Observe(time.Since(start).Seconds())
			logger.WithFields(logrus.Fields{"duration": time.Since(start), "epoch": epoch}).Infof("exported failed duties")
		}
	}()

	_, err = tx.Exec(`
insert into validator_failed_duties (validatorindex, slot, duty)
(
	select validatorindex, attesterslot, 3 from attestation_assignments_p where week = $1/1575 and epoch = $1 and status = 0
)
on conflict (validatorindex, slot, duty) do nothing`, epoch)
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: inserting missed attestations: %w", err)
	}

	_, err = tx.Exec(`
insert into validator_failed_duties (validatorindex, slot, duty)
(
	select proposer, slot, 4 from blocks where slot >= $1*32 and slot < ($1+1)*32 and status = '2'
)
on conflict (validatorindex, slot, duty) do nothing`, epoch)
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: inserting missed proposals: %w", err)
	}

	_, err = tx.Exec(`
insert into validator_failed_duties (validatorindex, slot, duty)
(
	select validatorindex, slot, 5 from sync_assignments_p where week = $1/1575 and slot >= $1*32 and slot < ($1+1)*32 and status = 2
)
on conflict (validatorindex, slot, duty) do nothing`, epoch)
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: inserting missed sync: %w", err)
	}

	_, err = tx.Exec(`insert into validator_failed_duties_status (epoch, status) values ($1, true)`, epoch)
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: inserting into validator_failed_duties_status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error in services.exportFailedDuties: commiting db-tx: %w", err)
	}

	return nil
}
