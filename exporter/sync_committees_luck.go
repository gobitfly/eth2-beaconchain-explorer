package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func syncCommitteesLuckExporter() {
	for {
		err := exportSyncCommitteesLuck()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Errorf("error exporting sync_committees_luck")
		}
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommitteesLuck() error {
	rowCount := uint64(0)
	err := db.WriterDb.Get(&rowCount, `SELECT COUNT(*) as count FROM sync_committees_luck`)
	if err != nil {
		return err
	}

	currEpoch := utils.TimeToEpoch(time.Now())
	currentPeriod := utils.SyncPeriodOfEpoch(uint64(currEpoch))
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.Config.AltairForkEpoch)

	dbPeriod := uint64(0)
	prevLuck := float64(0)
	if rowCount > 0 {
		err = db.WriterDb.Get(&dbPeriod, `SELECT MAX(period) FROM sync_committees_luck`)
		if err != nil {
			return err
		}

		if firstPeriod <= dbPeriod {
			// continue where we left off last time
			firstPeriod = dbPeriod + 1
		}

		err = db.WriterDb.Get(&prevLuck, `SELECT luck_agg FROM sync_committees_luck WHERE period = $1`, dbPeriod)
		if err != nil {
			return err
		}
	}

	for p := firstPeriod; p <= currentPeriod; p++ {
		t := time.Now()
		prevLuck, err = exportSyncCommitteesLuckAtPeriod(p, prevLuck)
		if err != nil {
			return fmt.Errorf("error exporting snyc-committee luck at period %v: %w", p, err)
		}
		logrus.WithFields(logrus.Fields{
			"period":   p,
			"duration": time.Since(t),
		}).Infof("exported sync_committees_luck")
	}

	return nil
}

func exportSyncCommitteesLuckAtPeriod(p uint64, pl float64) (float64, error) {
	logger.Infof("exporting sync committee luck for period %v", p)

	e := utils.FirstEpochOfSyncPeriod(p)
	totalValidatorsCount := uint64(0)
	err := db.WriterDb.Get(&totalValidatorsCount, "SELECT validatorscount FROM epochs WHERE epoch = $1", e)
	if err != nil {
		return 0, err
	}
	l := pl + (float64(utils.Config.Chain.Config.SyncCommitteeSize) / float64(totalValidatorsCount))

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees_luck (period, luck_agg) 
			VALUES (%d, %f)
			ON CONFLICT (period) DO UPDATE SET
				period = excluded.period,
				luck_agg = excluded.luck_agg`,
			p, l))
	if err != nil {
		return 0, err
	}

	return l, tx.Commit()
}
