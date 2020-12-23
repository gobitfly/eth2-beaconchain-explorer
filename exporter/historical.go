package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"time"
)

func historicalCollector() {
	for {
		start := time.Now()
		err := collectHistorical()
		if err != nil {
			logger.WithField("duration", time.Since(start)).WithError(err).Error("error collecting historical data")
		} else {
			logger.WithField("duration", time.Since(start)).Info("collected historical data")
		}
		time.Sleep(time.Minute * 10)
	}
}

func collectHistorical() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	keepHotEpochs := 7 * 24 * 3600 / (utils.Config.Chain.SlotsPerEpoch * utils.Config.Chain.SecondsPerSlot)
	// batchSize := uint64(255)
	batchSize := uint64(16) // not smaller than 4
	lastHistoricalEpoch := uint64(0)
	lastHotEpoch := uint64(0)

	// validator_balances
	err = tx.Get(&lastHistoricalEpoch, "select coalesce(max(epoch),0) from validator_balances_historical")
	if err != nil {
		return err
	}

	err = tx.Get(&lastHotEpoch, "select coalesce(max(epoch),0) as lastepoch from validator_balances")
	if err != nil {
		return err
	}

	if lastHistoricalEpoch > lastHotEpoch {
		return fmt.Errorf("lastHistoricalEpoch > lastHotEpoch: %v > %v", lastHistoricalEpoch, lastHotEpoch)
	}

	if lastHistoricalEpoch+batchSize < lastHotEpoch {
		res, err := tx.Exec("insert into validator_balances_historical select * from validator_balances where epoch > $1 and epoch <= $2", lastHistoricalEpoch, lastHistoricalEpoch+batchSize)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		logger.WithField("rowsAffected", rowsAffected).WithField("lowerLimit", lastHistoricalEpoch).WithField("upperLimit", lastHistoricalEpoch+batchSize).Info("collectHistorical: inserted into validator_balances_historical")
		lowerLimit := lastHotEpoch - keepHotEpochs
		if lastHistoricalEpoch+batchSize < lowerLimit {
			lowerLimit = lastHistoricalEpoch + batchSize
		}

		res, err = tx.Exec("delete from validator_balances where epoch < $1", lowerLimit)
		if err != nil {
			return err
		}
		rowsAffected, err = res.RowsAffected()
		if err != nil {
			return err
		}
		logger.WithField("rowsAffected", rowsAffected).WithField("lowerLimit", lowerLimit).Info("collectHistorical: deleted from validator_balances")
	}

	// attestation_assignments
	err = tx.Get(&lastHistoricalEpoch, "select coalesce(max(epoch),0) from attestation_assignments_historical")
	if err != nil {
		return err
	}

	err = tx.Get(&lastHotEpoch, "select coalesce(max(epoch),0) as lastepoch from attestation_assignments")
	if err != nil {
		return err
	}

	// do not collect younger epochs
	lastHotEpoch -= 4

	if lastHistoricalEpoch > lastHotEpoch {
		return fmt.Errorf("lastHistoricalEpoch > lastHotEpoch: %v > %v", lastHistoricalEpoch, lastHotEpoch)
	}

	if lastHistoricalEpoch+batchSize < lastHotEpoch {
		res, err := tx.Exec("insert into attestation_assignments_historical select * from attestation_assignments where epoch > $1 and epoch <= $2", lastHistoricalEpoch, lastHistoricalEpoch+batchSize)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		logger.WithField("rowsAffected", rowsAffected).WithField("lowerLimit", lastHistoricalEpoch).WithField("upperLimit", lastHistoricalEpoch+batchSize).Info("collectHistorical: inserted into attestation_assignments_historical")
		lowerLimit := lastHotEpoch - keepHotEpochs
		if lastHistoricalEpoch+batchSize < lowerLimit {
			lowerLimit = lastHistoricalEpoch + batchSize
		}
		_, err = tx.Exec("delete from attestation_assignments where epoch < $1", lowerLimit)
		if err != nil {
			return err
		}
		logger.WithField("rowsAffected", rowsAffected).WithField("lowerLimit", lowerLimit).Info("collectHistorical: deleted from attestation_assignments")
	}

	return tx.Commit()
}
