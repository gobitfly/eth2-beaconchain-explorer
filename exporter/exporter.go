package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/utils"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "exporter")

var Client *rpc.Client

// Start will start the export of data from rpc into the database
func Start(client rpc.Client) {
	go networkLivenessUpdater(client)
	go eth1DepositsExporter()
	go genesisDepositsExporter(client)
	go checkSubscriptions()
	go syncCommitteesExporter(client)
	go syncCommitteesCountExporter()
	if utils.Config.SSVExporter.Enabled {
		go ssvExporter()
	}
	if utils.Config.RocketpoolExporter.Enabled {
		go rocketpoolExporter()
	}

	if utils.Config.Indexer.PubKeyTagsExporter.Enabled {
		go UpdatePubkeyTag()
	}

	if utils.Config.MevBoostRelayExporter.Enabled {
		go mevBoostRelaysExporter()
	}
	// wait until the beacon-node is available
	for {
		head, err := client.GetChainHead()
		if err == nil {
			logger.Infof("beacon node is available with head slot: %v", head.HeadSlot)
			break
		}
		logger.Errorf("beacon-node seems to be unavailable: %v", err)
		time.Sleep(time.Second * 10)
	}

	firstRun := true

	minWaitTimeBetweenRuns := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)
	for {
		start := time.Now()
		err := RunSlotExporter(client, firstRun)
		if err != nil {
			logrus.Errorf("error during slot export run: %v", err)
		} else if err == nil && firstRun {
			firstRun = false
		}

		logrus.Info("update run completed")
		elapsed := time.Since(start)
		if elapsed < minWaitTimeBetweenRuns {
			time.Sleep(minWaitTimeBetweenRuns - elapsed)
		}

		services.ReportStatus("slotExporter", "Running", nil)
	}
}

func networkLivenessUpdater(client rpc.Client) {
	var prevHeadEpoch uint64
	err := db.WriterDb.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	if err != nil {
		utils.LogFatal(err, "getting previous head epoch from db error", 0)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)

	for {
		head, err := client.GetChainHead()
		if err != nil {
			logger.Errorf("error getting chainhead when exporting networkliveness: %v", err)
			time.Sleep(slotDuration)
			continue
		}

		if prevHeadEpoch == head.HeadEpoch {
			time.Sleep(slotDuration)
			continue
		}

		// wait for node to be synced
		if time.Now().Add(-epochDuration).After(utils.EpochToTime(head.HeadEpoch)) {
			time.Sleep(slotDuration)
			continue
		}

		_, err = db.WriterDb.Exec(`
			INSERT INTO network_liveness (ts, headepoch, finalizedepoch, justifiedepoch, previousjustifiedepoch)
			VALUES (NOW(), $1, $2, $3, $4)`,
			head.HeadEpoch, head.FinalizedEpoch, head.JustifiedEpoch, head.PreviousJustifiedEpoch)
		if err != nil {
			logger.Errorf("error saving networkliveness: %v", err)
		} else {
			logger.Printf("updated networkliveness for epoch %v", head.HeadEpoch)
			prevHeadEpoch = head.HeadEpoch
		}

		time.Sleep(slotDuration)
	}
}

func genesisDepositsExporter(client rpc.Client) {
	for {
		// check if the beaconchain has started
		var latestEpoch uint64
		err := db.WriterDb.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if latestEpoch == 0 {
			time.Sleep(time.Minute)
			continue
		}

		// check if genesis-deposits have already been exported
		var genesisDepositsCount uint64
		err = db.WriterDb.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
		if err != nil {
			logger.Errorf("error retrieving genesis-deposits-count when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		// if genesis-deposits have already been exported exit this go-routine
		if genesisDepositsCount > 0 {
			return
		}

		genesisValidators, err := client.GetValidatorState(0)
		if err != nil {
			logger.Errorf("error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			logger.Errorf("error beginning db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		logger.Infof("exporting deposit data for %v genesis validators", len(genesisValidators.Data))
		for i, validator := range genesisValidators.Data {
			if i%1000 == 0 {
				logger.Infof("exporting deposit data for genesis validator %v (%v/%v)", validator.Index, i, len(genesisValidators.Data))
			}
			_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
			VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
				validator.Index, utils.MustParseHex(validator.Validator.Pubkey), utils.MustParseHex(validator.Validator.WithdrawalCredentials), validator.Balance, []byte{0x0},
			)
			if err != nil {
				tx.Rollback()
				logger.Errorf("error exporting genesis-deposits: %v", err)
				time.Sleep(time.Minute)
				continue
			}
		}

		// hydrate the eth1 deposit signature for all genesis validators that have a corresponding eth1 deposit
		_, err = tx.Exec(`
			UPDATE blocks_deposits 
			SET signature = a.signature 
			FROM (
				SELECT DISTINCT ON(publickey) publickey, signature 
				FROM eth1_deposits 
				WHERE valid_signature = true) AS a 
			WHERE block_slot = 0 AND blocks_deposits.publickey = a.publickey AND blocks_deposits.signature = '\x'`)
		if err != nil {
			tx.Rollback()
			logger.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		// update deposits-count
		_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", len(genesisValidators.Data))
		if err != nil {
			tx.Rollback()
			logger.Errorf("error updating deposit count for the genesis slot: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			logger.Errorf("error committing db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		logger.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
		return
	}
}
