package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"time"
)

func UpdatePubkeyTag() {
	logger.Infoln("Started Pubkey Tags Updater")
	for true {
		start := time.Now()

		tx, err := db.DB.Beginx()
		if err != nil {
			logger.WithError(err).Error("Error connecting to DB")
			// return err
		}
		_, err = tx.Exec(`INSERT INTO validator_tags (publickey, tag)
		SELECT publickey, FORMAT('pool:%s', sps.name) tag
		FROM eth1_deposits
		inner join stake_pools_stats as sps on ENCODE(from_address::bytea, 'hex')=sps.address
		WHERE sps.name NOT LIKE '%Rocketpool -%'
		ON CONFLICT (publickey, tag) DO NOTHING;`)
		if err != nil {
			logger.WithError(err).Error("Error updating validator_tags")
			// return err
		}

		err = tx.Commit()
		if err != nil {
			logger.WithError(err).Error("Error commiting transaction")
		}
		tx.Rollback()

		logger.Infof("Updating Pubkey Tags took %v sec.", time.Now().Sub(start).Seconds())
		metrics.TaskDuration.WithLabelValues("validator_pubkey_tag_updater").Observe(time.Since(start).Seconds())

		time.Sleep(time.Minute * 10)
	}
}
