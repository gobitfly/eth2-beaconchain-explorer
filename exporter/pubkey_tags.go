package exporter

import (
	"eth2-exporter/db"
	"time"
)

func UpdatePubkeyTag() {
	logger.Infoln("Starting Pubkey Tag updater")
	for true {
		tx, err := db.DB.Beginx()
		if err != nil {
			logger.WithError(err).Error("Error connecting to DB")
			// return err
		}
		_, err = tx.Exec(`INSERT INTO validator_tags (publickey, tag)
		SELECT publickey, FORMAT('%s:%s', sps.name, sps.category) tag
		FROM eth1_deposits
		inner join stake_pools_stats as sps on ENCODE(from_address::bytea, 'hex')=sps.address
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
		time.Sleep(time.Minute * 10)
	}
}
