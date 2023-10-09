package exporter

import (
	"bytes"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strconv"
	"strings"
)

func StartSlotExporter(client rpc.Client) error {
	// next get all slots we currently have in the database
	dbSlots, err := db.GetAllSlots()
	if err != nil {
		return fmt.Errorf("error retrieving all db slots: %w", err)
	}

	lastDbSlot := uint64(0)
	for slotIndex := 1; slotIndex < len(dbSlots); slotIndex++ {
		previousSlot := dbSlots[slotIndex-1]
		currentSlot := dbSlots[slotIndex]

		if previousSlot.Slot != currentSlot.Slot-1 {
			logger.Info("slots between %v and %v are missing, exporting them", previousSlot.Slot, currentSlot.Slot)
			for slot := previousSlot.Slot + 1; slot <= currentSlot.Slot-1; slot++ {
				err := ExportSlot(client, slot, false)

				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", slot, err)
				}
			}
		}
	}

	if len(dbSlots) != 0 {
		lastDbSlot = dbSlots[len(dbSlots)-1].Slot
	}

	// at this point we know that we have a coherent list of slots in the database without any gaps
	// get the current chain head
	head, err := client.GetChainHead()

	if err != nil {
		return fmt.Errorf("error retrieving chain head: %w", err)
	}

	if len(dbSlots) == 0 {
		logger.Infof("db is empty, export genesis slot")
		err := ExportSlot(client, 0, utils.EpochOfSlot(0) == head.HeadEpoch)
		if err != nil {
			return fmt.Errorf("error exporting slot %v: %w", 0, err)
		}
	}

	// check if any new slots have been added to the chain
	if lastDbSlot != head.HeadSlot {
		for slot := lastDbSlot + 1; slot <= head.HeadSlot; slot++ { // export any new slots
			err := ExportSlot(client, slot, utils.EpochOfSlot(slot) == head.HeadEpoch)
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", slot, err)
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
			// block has finalized, mark it in the db
			if header != nil && bytes.Equal(dbSlot.BlockRoot, utils.MustParseHex(header.Data.Root)) {
				// no reorg happened, simply mark the block as final
				logger.Infof("setting block %v as finalized (proposed)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, dbSlot.Status)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (proposed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) < 32 {
				// no reorg happened, mark the block as missed
				logger.Infof("setting block %v as finalized (missed)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "2")
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (missed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) == 32 {
				// block has been orphaned, mark the block as orphaned
				logger.Infof("setting block %v as finalized (orphaned)", dbSlot.Slot)
				err := db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3")
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
			}

			// epoch transition slot has finalized, update epoch status
			if dbSlot.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 && dbSlot.Slot > utils.Config.Chain.ClConfig.SlotsPerEpoch-1 {
				epoch := utils.EpochOfSlot(dbSlot.Slot)
				epochParticipationStats, err := client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					logger.Printf("error retrieving epoch participation statistics: %v", err)
				} else {
					logger.Printf("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
					err := db.UpdateEpochStatus(epochParticipationStats)

					if err != nil {
						return err
					}
				}
			}
		} else { // check if the slot has been proposed in the meantime
			if len(dbSlot.BlockRoot) < 32 && header != nil { // we have no block in the db, but the node has a block, export it
				err := ExportSlot(client, dbSlot.Slot, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch)
				if err != nil {
					return fmt.Errorf("error exporting block %v: %w", dbSlot.Slot, err)
				}
			}
		}
	}

	return nil

}

func ExportSlot(client rpc.Client, slot uint64, isHeadEpoch bool) error {

	logger.Infof("exporting slot %v", slot)

	block, err := client.GetBlockBySlot(slot)
	if err != nil {
		return fmt.Errorf("error retrieving data for slot %v: %w", slot, err)
	}

	if block.EpochAssignments != nil { // export the epoch assignments as they are included in the first slot of an epoch
		epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
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
				return fmt.Errorf("error parsing attested slot from attestation key: %v", err)
			}

			if attDutiesEpoch[types.Slot(attestedSlot)] == nil {
				attDutiesEpoch[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}

			attDutiesEpoch[types.Slot(attestedSlot)][types.ValidatorIndex(validatorIndex)] = []types.Slot{}
		}

		err = db.BigtableClient.SaveAttestationDuties(attDutiesEpoch)
		if err != nil {
			return fmt.Errorf("error exporting attestation assignments to bigtable for block %v: %v", block.Slot, err)
		}
		err = db.BigtableClient.SaveSyncComitteeDuties(syncDutiesEpoch)
		if err != nil {
			return fmt.Errorf("error exporting sync committee assignments to bigtable for block %v: %v", block.Slot, err)
		}
		err = db.BigtableClient.SaveProposalAssignments(epoch, block.EpochAssignments.ProposerAssignments)
		if err != nil {
			return fmt.Errorf("error exporting proposal assignments to bigtable: %v", err)
		}
		err = db.BigtableClient.SaveValidatorBalances(epoch, block.Validators)
		if err != nil {
			return fmt.Errorf("error exporting validator balances to bigtable for block %v: %v", block.Slot, err)
		}

		err = db.SaveEpoch(epoch, block.Validators, client)
		if err != nil {
			return fmt.Errorf("error saving epoch data: %w", err)
		}

		if epoch > 0 {
			epochParticipationStats, err := client.GetValidatorParticipation(epoch - 1)
			if err != nil {
				logger.Printf("error retrieving epoch participation statistics: %v", err)
			} else {
				logger.Printf("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
				err := db.UpdateEpochStatus(epochParticipationStats)

				if err != nil {
					return err
				}
			}
		}

		if isHeadEpoch {
			err = db.SaveValidators(epoch, block.Validators, client, 10000)
			if err != nil {
				return fmt.Errorf("error saving validators for epoch %v: %w", epoch, err)
			}

			err = db.UpdateQueueDeposits()
			if err != nil {
				return fmt.Errorf("error updating queue deposits cache: %w", err)
			}
		}
	}

	// save the duties to bigtable
	syncDuties := make(map[types.Slot]map[types.ValidatorIndex]bool)
	syncDuties[types.Slot(block.Slot)] = make(map[types.ValidatorIndex]bool)

	for validator, duty := range block.SyncDuties {
		syncDuties[types.Slot(block.Slot)][types.ValidatorIndex(validator)] = duty
	}

	attDuties := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
	for validator, attestedSlot := range block.AttestationDuties {
		if attDuties[types.Slot(attestedSlot)] == nil {
			attDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
		}
		if attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] == nil {
			attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = []types.Slot{}
		}
		attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = append(attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)], types.Slot(block.Slot))
	}
	err = db.BigtableClient.SaveAttestationDuties(attDuties)
	if err != nil {
		return fmt.Errorf("error exporting attestations to bigtable for block %v: %v", block.Slot, err)
	}
	err = db.BigtableClient.SaveSyncComitteeDuties(syncDuties)
	if err != nil {
		return fmt.Errorf("error exporting sync committee duties to bigtable for block %v: %v", block.Slot, err)
	}

	// save the block data to the db
	err = db.SaveGraffitiwall(block)
	if err != nil {
		logger.Errorf("error saving block to the db: %v", err)
	}
	err = db.SaveBlock(block, false)
	if err != nil {
		logger.Errorf("error saving block to the db: %v", err)
	}

	return nil
}
