package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func syncCommitteesCountExporter() {
	for {
		err := exportSyncCommitteesCount()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Errorf("error exporting sync_committees_count_per_validator")
		}
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommitteesCount() error {
	rowCount := uint64(0)
	err := db.WriterDb.Get(&rowCount, `SELECT COUNT(*) FROM sync_committees_count_per_validator`)
	if err != nil {
		return err
	}

	latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
	if err != nil {
		logger.Errorf("error retrieving latest exported finalized epoch from the database: %v", err)
	}

	currentPeriod := utils.SyncPeriodOfEpoch(latestFinalizedEpoch)
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.Config.AltairForkEpoch)

	dbPeriod := uint64(0)
	countSoFar := float64(0)
	if rowCount > 0 {
		err = db.WriterDb.Get(&dbPeriod, `SELECT MAX(period) FROM sync_committees_count_per_validator`)
		if err != nil {
			return err
		}

		if firstPeriod <= dbPeriod {
			// continue where we left off last time
			firstPeriod = dbPeriod + 1
		}

		err = db.WriterDb.Get(&countSoFar, `SELECT count_so_far FROM sync_committees_count_per_validator WHERE period = $1`, dbPeriod)
		if err != nil {
			return err
		}
	}

	for period := firstPeriod; period <= currentPeriod; period++ {
		t := time.Now()
		countSoFar, err = exportSyncCommitteesCountAtPeriod(period, countSoFar)
		if err != nil {
			return fmt.Errorf("error exporting sync-committee count at period %v: %w", period, err)
		}
		logrus.WithFields(logrus.Fields{
			"period":   period,
			"duration": time.Since(t),
		}).Infof("exported sync_committees_count_per_validator")
	}

	return nil
}

func exportSyncCommitteesCountAtPeriod(period uint64, countSoFar float64) (float64, error) {
	logger.Infof("exporting sync committee count for period %v", period)

	count := 0.0
	if period > 0 {
		e := utils.FirstEpochOfSyncPeriod(period - 1)
		totalValidatorsCount := uint64(0)
		err := db.WriterDb.Get(&totalValidatorsCount, "SELECT validatorscount FROM epochs WHERE epoch = $1", e)
		if err != nil {
			return 0, fmt.Errorf("error retrieving validatorscount for epoch %v: %v", e, err)
		}
		count = countSoFar + (float64(utils.Config.Chain.Config.SyncCommitteeSize) / float64(totalValidatorsCount))
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees_count_per_validator (period, count_so_far) 
			VALUES (%d, %f)
			ON CONFLICT (period) DO UPDATE SET
				period = excluded.period,
				count_so_far = excluded.count_so_far`,
			period, count))
	if err != nil {
		return 0, err
	}

	return count, tx.Commit()
}
