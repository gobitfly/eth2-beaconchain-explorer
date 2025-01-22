package exporter

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func RunSlotExporter(client rpc.Client, firstRun bool) error {
	// get the current chain head
	head, err := client.GetChainHead()

	if err != nil {
		return fmt.Errorf("error retrieving chain head: %w", err)
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting tx: %w", err)
	}
	defer tx.Rollback()

	if firstRun {
		// get all slots we currently have in the database
		dbSlots, err := db.GetAllSlots(tx)
		if err != nil {
			return fmt.Errorf("error retrieving all db slots: %w", err)
		}

		if len(dbSlots) > 0 {
			if dbSlots[0] != 0 {
				logger.Infof("exporting genesis slot as it is missing in the database")
				err := ExportSlot(client, 0, utils.EpochOfSlot(0) == head.HeadEpoch, tx)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", 0, err)
				}
				dbSlots, err = db.GetAllSlots(tx)
				if err != nil {
					return fmt.Errorf("error retrieving all db slots: %w", err)
				}
			}
		}

		if len(dbSlots) > 1 {
			// export any gaps we might have (for whatever reason)
			for slotIndex := 1; slotIndex < len(dbSlots); slotIndex++ {
				previousSlot := dbSlots[slotIndex-1]
				currentSlot := dbSlots[slotIndex]

				if previousSlot != currentSlot-1 {
					logger.Infof("slots between %v and %v are missing, exporting them", previousSlot, currentSlot)
					for slot := previousSlot + 1; slot <= currentSlot-1; slot++ {
						err := ExportSlot(client, slot, false, tx)

						if err != nil {
							return fmt.Errorf("error exporting slot %v: %w", slot, err)
						}
					}
				}
			}
		}
	}

	// at this point we know that we have a coherent list of slots in the database without any gaps
	lastDbSlot := uint64(0)
	err = tx.Get(&lastDbSlot, "SELECT slot FROM blocks ORDER BY slot DESC limit 1")

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Infof("db is empty, export genesis slot")
			err := ExportSlot(client, 0, utils.EpochOfSlot(0) == head.HeadEpoch, tx)
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", 0, err)
			}
			lastDbSlot = 0
		} else {
			return fmt.Errorf("error retrieving last slot from the db: %w", err)
		}
	}

	// check if any new slots have been added to the chain
	if lastDbSlot != head.HeadSlot {
		slotsExported := 0
		for slot := lastDbSlot + 1; slot <= head.HeadSlot; slot++ { // export any new slots
			err := ExportSlot(client, slot, utils.EpochOfSlot(slot) == head.HeadEpoch, tx)
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", slot, err)
			}
			slotsExported++

			// in case of large export runs, export at most 10 epochs per tx
			if slotsExported == int(utils.Config.Chain.ClConfig.SlotsPerEpoch)*10 {
				err := tx.Commit()

				if err != nil {
					return fmt.Errorf("error committing tx: %w", err)
				}

				return nil
			}
		}
	}

	// at this point we have all all data up to the current chain head in the database

	// check if any non-finalized slot has changed by comparing it with the node
	dbNonFinalSlots, err := db.GetAllNonFinalizedSlots()
	if err != nil {
		return fmt.Errorf("error retrieving all non finalized slots from the db: %w", err)
	}
	for _, dbSlot := range dbNonFinalSlots {
		header, err := client.GetBlockHeader(dbSlot.Slot)

		if err != nil {
			return fmt.Errorf("error retrieving block root for slot %v: %w", dbSlot.Slot, err)
		}

		nodeSlotFinalized := dbSlot.Slot <= head.FinalizedSlot

		if nodeSlotFinalized != dbSlot.Finalized {
			// slot has finalized, mark it in the db
			if header != nil && bytes.Equal(dbSlot.BlockRoot, utils.MustParseHex(header.Data.Root)) {
				// no reorg happened, simply mark the slot as final
				logger.Infof("setting slot %v as finalized (proposed)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, dbSlot.Status, tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (proposed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) < 32 {
				// no reorg happened, mark the slot as missed
				logger.Infof("setting slot %v as finalized (missed)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "2", tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (missed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) == 32 {
				// slot has been orphaned, mark the slot as orphaned
				logger.Infof("setting slot %v as finalized (orphaned)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
			} else if header != nil && !bytes.Equal(utils.MustParseHex(header.Data.Root), dbSlot.BlockRoot) {
				// we have a different block root for the slot in the db, mark the currently present one as orphaned and write the new one
				logger.Infof("setting slot %v as orphaned and exporting new slot", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
				err = ExportSlot(client, dbSlot.Slot, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, tx)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}

			// epoch transition slot has finalized, update epoch status
			if dbSlot.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 && dbSlot.Slot > utils.Config.Chain.ClConfig.SlotsPerEpoch-1 {
				epoch := utils.EpochOfSlot(dbSlot.Slot)
				epochParticipationStats, err := client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				} else {
					logger.Printf("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
					err := db.UpdateEpochStatus(epochParticipationStats, tx)

					if err != nil {
						return err
					}

					logger.Infof("exporting validation queue")
					queue, err := client.GetValidatorQueue()
					if err != nil {
						return fmt.Errorf("error retrieving validator queue data: %w", err)
					}

					err = db.SaveValidatorQueue(queue, tx)
					if err != nil {
						return fmt.Errorf("error saving validator queue data: %w", err)
					}
				}

			}
		} else { // check if a late slot has been proposed in the meantime
			if len(dbSlot.BlockRoot) < 32 && header != nil { // we have no slot in the db, but the node has a slot, export it
				logger.Infof("exporting new slot %v", dbSlot.Slot)
				err := ExportSlot(client, dbSlot.Slot, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, tx)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx: %w", err)
	}

	return nil

}

func ExportSlot(client rpc.Client, slot uint64, isHeadEpoch bool, tx *sqlx.Tx) error {

	isFirstSlotOfEpoch := slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0
	epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch

	if isFirstSlotOfEpoch {
		logger.Infof("exporting slot %v (epoch transition into epoch %v)", slot, epoch)
	} else {
		logger.Infof("exporting slot %v", slot)
	}
	start := time.Now()

	// retrieve the data for the slot from the node
	// the first slot of an epoch will also contain all validator duties for the whole epoch
	block, err := client.GetBlockBySlot(slot)
	if err != nil {
		return fmt.Errorf("error retrieving data for slot %v: %w", slot, err)
	}

	if block.EpochAssignments != nil { // export the epoch assignments as they are included in the first slot of an epoch
		logger.Infof("exporting duties & balances for epoch %v", utils.EpochOfSlot(slot))

		// prepare the duties for export to bigtable
		syncDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex]bool)
		attDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if syncDutiesEpoch[types.Slot(slot)] == nil {
				syncDutiesEpoch[types.Slot(slot)] = make(map[types.ValidatorIndex]bool)
			}
			for _, validatorIndex := range block.EpochAssignments.SyncAssignments {
				syncDutiesEpoch[types.Slot(slot)][types.ValidatorIndex(validatorIndex)] = false
			}
		}

		for key, validatorIndex := range block.EpochAssignments.AttestorAssignments {
			keySplit := strings.Split(key, "-")
			attestedSlot, err := strconv.ParseUint(keySplit[0], 10, 64)

			if err != nil {
				return fmt.Errorf("error parsing attested slot from attestation key: %w", err)
			}

			if attDutiesEpoch[types.Slot(attestedSlot)] == nil {
				attDutiesEpoch[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}

			attDutiesEpoch[types.Slot(attestedSlot)][types.ValidatorIndex(validatorIndex)] = []types.Slot{}
		}

		g := errgroup.Group{}

		// save all duties to bigtable
		g.Go(func() error {
			err := db.BigtableClient.SaveAttestationDuties(attDutiesEpoch)
			if err != nil {
				return fmt.Errorf("error exporting attestation assignments to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})
		g.Go(func() error {
			err := db.BigtableClient.SaveSyncComitteeDuties(syncDutiesEpoch)
			if err != nil {
				return fmt.Errorf("error exporting sync committee assignments to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})

		// save the validator balances to bigtable
		g.Go(func() error {
			err := db.BigtableClient.SaveValidatorBalances(epoch, block.Validators)
			if err != nil {
				return fmt.Errorf("error exporting validator balances to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})
		// if we are exporting the head epoch, update the validator db table
		if isHeadEpoch {
			g.Go(func() error {
				err := db.SaveValidators(epoch, block.Validators, client, 10000, tx)
				if err != nil {
					return fmt.Errorf("error saving validators for epoch %v: %w", epoch, err)
				}

				// also update the queue deposit table once every epoch
				err = db.UpdateQueueDeposits(tx)
				if err != nil {
					return fmt.Errorf("error updating queue deposits cache: %w", err)
				}
				return nil
			})
		}

		var epochParticipationStats *types.ValidatorParticipation
		if epoch > 0 {
			g.Go(func() error {
				// retrieve the epoch participation stats
				var err error
				epochParticipationStats, err = client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					return fmt.Errorf("error retrieving epoch participation statistics: %w", err)
				}
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			return err
		}

		// save the epoch metadata to the database
		err = db.SaveEpoch(epoch, block.Validators, client, tx)
		if err != nil {
			return fmt.Errorf("error saving epoch data: %w", err)
		}

		if epoch > 0 && epochParticipationStats != nil {
			logger.Printf("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
			err := db.UpdateEpochStatus(epochParticipationStats, tx)

			if err != nil {
				return err
			}
		}

		// time.Sleep(time.Minute)
	}

	// for the slot itself start by preparing the duties for export to bigtable
	syncDuties := make(map[types.Slot]map[types.ValidatorIndex]bool)
	syncDuties[types.Slot(block.Slot)] = make(map[types.ValidatorIndex]bool)

	for validator, duty := range block.SyncDuties {
		syncDuties[types.Slot(block.Slot)][types.ValidatorIndex(validator)] = duty
	}

	attDuties := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
	for validator, attestedSlots := range block.AttestationDuties {

		for _, attestedSlot := range attestedSlots {
			if attDuties[types.Slot(attestedSlot)] == nil {
				attDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}
			if attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] == nil {
				attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = make([]types.Slot, 0, 10)
			}
			attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = append(attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)], types.Slot(block.Slot))
		}
	}

	// save sync & attestation duties to bigtable
	err = db.BigtableClient.SaveAttestationDuties(attDuties)
	if err != nil {
		return fmt.Errorf("error exporting attestations to bigtable for slot %v: %w", block.Slot, err)
	}
	err = db.BigtableClient.SaveSyncComitteeDuties(syncDuties)
	if err != nil {
		return fmt.Errorf("error exporting sync committee duties to bigtable for slot %v: %w", block.Slot, err)
	}

	// save the block data to the db
	err = db.SaveBlock(block, false, tx)
	if err != nil {
		return fmt.Errorf("error saving slot to the db: %w", err)
	}
	// time.Sleep(time.Second)

	logger.WithFields(
		logrus.Fields{
			"slot":      block.Slot,
			"blockRoot": fmt.Sprintf("%x", block.BlockRoot),
		},
	).Infof("! export of slot completed, took %v", time.Since(start))

	return nil
}
