package exporter

import (
	"eth2-exporter/db"
	"time"
)

func genesisDepositsExporter() {
	for {
		// check if the beaconchain has started
		var latestEpoch uint64
		err := db.DB.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if latestEpoch == 0 {
			time.Sleep(time.Second * 60)
			continue
		}

		// check if genesis-deposits have already been exported
		var genesisDepositsCount uint64
		err = db.DB.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
		if err != nil {
			logger.Errorf("error retrieving genesis-deposits-count when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// if genesis-deposits have already been exported exit this go-routine
		if genesisDepositsCount > 0 {
			return
		}

		// get genesis-validators-count
		var genesisValidatorsCount uint64
		err = db.DB.Get(&genesisValidatorsCount, "SELECT validatorscount FROM epochs WHERE epoch=0")
		if err != nil {
			logger.Errorf("error retrieving validatorscount for genesis-epoch when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// check if eth1-deposits have already been exported
		var missingEth1Deposits uint64
		err = db.DB.Get(&missingEth1Deposits, `
			SELECT COUNT(*)
			FROM validators v
			LEFT JOIN ( 
				SELECT DISTINCT ON (publickey) publickey, signature FROM eth1_deposits 
			) d ON d.publickey = v.pubkey
			WHERE d.publickey IS NULL AND v.validatorindex < $1`, genesisValidatorsCount)
		if err != nil {
			logger.Errorf("error retrieving missing-eth1-deposits-count when exporting genesis-deposits")
			time.Sleep(time.Second * 60)
			continue
		}

		if missingEth1Deposits > 0 {
			logger.Infof("delaying export of genesis-deposits until eth1-deposits have been exported")
			time.Sleep(time.Second * 60)
			continue
		}

		tx, err := db.DB.Beginx()
		if err != nil {
			logger.Errorf("error beginning db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// export genesis-deposits from eth1-deposits and data already gathered from the eth2-client
		_, err = tx.Exec(`
			INSERT INTO blocks_deposits (block_slot, block_index, publickey, withdrawalcredentials, amount, signature)
			SELECT
				0 as block_slot,
				v.validatorindex as block_index,
				v.pubkey as publickey,
				v.withdrawalcredentials,
				b.balance as amount,
				d.signature as signature
			FROM validators v
			LEFT JOIN validator_balances b 
				ON v.validatorindex = b.validatorindex
				AND b.epoch = 0
			LEFT JOIN ( 
				SELECT DISTINCT ON (publickey) publickey, signature FROM eth1_deposits 
			) d ON d.publickey = v.pubkey
			WHERE v.validatorindex < $1`, genesisValidatorsCount)
		if err != nil {
			tx.Rollback()
			logger.Errorf("error exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// update deposits-count
		_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", genesisValidatorsCount)
		if err != nil {
			tx.Rollback()
			logger.Errorf("error exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			logger.Errorf("error committing db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		logger.Infof("exported genesis-deposits for %v genesis-validators", genesisValidatorsCount)
		return
	}
}
