package exporter

import (
	"bytes"
	"context"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"sync"
	"time"

	ptypes "github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "exporter")

// If exporting an epoch fails for 10 consecutive times exporting this epoch will be disabled
// This is a workaround for a bug in the prysm archive node that causes epochs without blocks
// to not be archived properly (see https://github.com/prysmaticlabs/prysm/issues/4165)
var epochBlacklist = make(map[uint64]uint64)

func Start(client ethpb.BeaconChainClient) error {

	if utils.Config.Indexer.FullIndexOnStartup {
		logger.Printf("Performing one time full db reindex")
		head, err := client.GetChainHead(context.Background(), &ptypes.Empty{})
		if err != nil {
			logger.Fatal(err)
		}

		var wg sync.WaitGroup
		for epoch := uint64(0); epoch <= head.HeadEpoch; epoch++ {
			//err := exportEpoch(epoch, client)

			if err != nil {
				logger.Fatal(err)
			}
			wg.Add(1)

			logger.Printf("Exporting epoch %v of %v", epoch, head.HeadEpoch)
			go func(e uint64) {
				err := ExportEpoch(e, client)

				if err != nil {
					logger.Fatal(err)
				}
				logger.Printf("Finished export for epoch %v", e)
				wg.Done()
			}(epoch)

			if epoch%10 == 0 {
				logger.Printf("Waiting...")
				wg.Wait()
			}
		}
	}

	if utils.Config.Indexer.IndexMissingEpochsOnStartup {
		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		epochs, err := db.GetAllEpochs()
		if err != nil {
			logger.Fatal(err)
		}

		for i := 0; i < len(epochs)-1; i++ {
			if epochs[i] != epochs[i+1]-1 && epochs[i] != epochs[i+1] {
				logger.Println("Epochs between", epochs[i], "and", epochs[i+1], "are missing!")

				for epoch := epochs[i]; epoch <= epochs[i+1]; epoch++ {
					err := ExportEpoch(epoch, client)
					if err != nil {
						logger.Error(err)
					}
					logger.Printf("Finished export for epoch %v", epoch)
				}
			}
		}
	}

	if utils.Config.Indexer.CheckAllBlocksOnStartup {
		// Make sure that all blocks are correct by comparing all block hashes in the database to the ones we have in the node
		head, err := client.GetChainHead(context.Background(), &ptypes.Empty{})
		if err != nil {
			logger.Fatal(err)
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(1, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		nodeBlocks, err := GetLastBlocks(1, head.HeadEpoch, client)
		if err != nil {
			logger.Fatal(err)
		}

		blocksMap := make(map[string]*types.BlockComparisonContainer)

		for _, block := range dbBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Db = block
		}
		for _, block := range nodeBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Node = block
		}

		epochsToExport := make(map[uint64]bool)

		for key, block := range blocksMap {
			if block.Db == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
				logger.Printf("Queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		logger.Printf("Exporting %v epochs.", len(epochsToExport))

		keys := make([]uint64, 0)
		for k := range epochsToExport {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, epoch := range keys {
			logger.Printf("Exporting epoch %v", epoch)

			err = ExportEpoch(epoch, client)

			if err != nil {
				logger.Errorf("error exporting epoch: %v", err)
				if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
					epochBlacklist[epoch]++
				}
			}
			logger.Printf("Finished export for epoch %v", epoch)
		}
	}

	for true {

		head, err := client.GetChainHead(context.Background(), &ptypes.Empty{})
		if err != nil {
			logger.Fatal(err)
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(head.FinalizedEpoch-1, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		nodeBlocks, err := GetLastBlocks(head.FinalizedEpoch-1, head.HeadEpoch, client)
		if err != nil {
			logger.Fatal(err)
		}

		blocksMap := make(map[string]*types.BlockComparisonContainer)

		for _, block := range dbBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Db = block
		}
		for _, block := range nodeBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Node = block
		}

		epochsToExport := make(map[uint64]bool)

		for key, block := range blocksMap {
			if block.Db == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
				logger.Printf("Queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		epochs, err := db.GetAllEpochs()
		if err != nil {
			logger.Fatal(err)
		}

		// Add not yet exported epochs to the export set (for example during the initial sync)
		if len(epochs) > 0 && epochs[len(epochs)-1] < head.HeadEpoch {
			for i := epochs[len(epochs)-1]; i <= head.HeadEpoch; i++ {
				epochsToExport[i] = true
			}
		} else if len(epochs) > 0 && epochs[0] != 0 { // Export the genesis epoch if not yet present in the db
			epochsToExport[0] = true
		} else if len(epochs) == 0 { // No epochs are present int the db
			for i := uint64(0); i <= head.HeadEpoch; i++ {
				epochsToExport[i] = true
			}
		}

		logger.Printf("Exporting %v epochs.", len(epochsToExport))

		keys := make([]uint64, 0)
		for k := range epochsToExport {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, epoch := range keys {
			if epochBlacklist[epoch] > 3 {
				logger.Printf("Skipping export of epoch %v as it has errored %v times", epochBlacklist[epoch])
				continue
			}

			logger.Printf("Exporting epoch %v", epoch)

			err = ExportEpoch(epoch, client)

			if err != nil {
				logger.Errorf("error exporting epoch: %v", err)
				if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
					epochBlacklist[epoch]++
				}
			}
			logger.Printf("Finished export for epoch %v", epoch)
		}

		// Update epoch statistics up to 10 epochs after the last finalized epoch
		startEpoch := uint64(0)
		if head.FinalizedEpoch > 10 {
			startEpoch = head.FinalizedEpoch - 10
		}
		err = updateEpochStatus(client, startEpoch, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		err = exportAttestationPool(client)
		if err != nil {
			logger.Fatal(err)
		}

		err = exportValidatorQueue(client)
		if err != nil {
			logger.Fatal(err)
		}

		err = MarkOrphanedBlocks(head.FinalizedEpoch-1, head.HeadEpoch, nodeBlocks)
		if err != nil {
			logger.Fatal(err)
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}

func MarkOrphanedBlocks(startEpoch, endEpoch uint64, blocks []*types.MinimalBlock) error {
	blocksMap := make(map[string]bool)

	for _, block := range blocks {
		blocksMap[fmt.Sprintf("%x", block.BlockRoot)] = false
	}

	orphanedBlocks := make([][]byte, 0)
	parentRoot := ""
	for i := len(blocks) - 1; i >= 0; i-- {
		blockRoot := fmt.Sprintf("%x", blocks[i].BlockRoot)

		if i == len(blocks)-1 { // First block is always canon
			parentRoot = fmt.Sprintf("%x", blocks[i].ParentRoot)
			blocksMap[blockRoot] = true
			continue
		}
		if parentRoot != blockRoot { // Block is not part of the canonical chain
			logger.Errorf("Block %x at slot %v in epoch %v has been orphaned", blocks[i].BlockRoot, blocks[i].Slot, blocks[i].Epoch)
			orphanedBlocks = append(orphanedBlocks, blocks[i].BlockRoot)
			continue
		}
		blocksMap[blockRoot] = true
		parentRoot = fmt.Sprintf("%x", blocks[i].ParentRoot)
	}

	return db.UpdateCanonicalBlocks(startEpoch, endEpoch, orphanedBlocks)
}

func GetLastBlocks(startEpoch, endEpoch uint64, client ethpb.BeaconChainClient) ([]*types.MinimalBlock, error) {
	blocks := make([]*types.MinimalBlock, 0)

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		startSlot := epoch * utils.SlotsPerEpoch
		endSlot := (epoch+1)*utils.SlotsPerEpoch - 1
		for slot := startSlot; slot <= endSlot; slot++ {
			blocksResponse, err := client.ListBlocks(context.Background(), &ethpb.ListBlocksRequest{PageSize: utils.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Slot{Slot: slot}, IncludeNoncanonical: true})
			if err != nil {
				logger.Fatal(err)
			}

			for _, block := range blocksResponse.BlockContainers {
				blocks = append(blocks, &types.MinimalBlock{
					Epoch:      epoch,
					Slot:       block.Block.Slot,
					BlockRoot:  block.BlockRoot,
					ParentRoot: block.Block.ParentRoot,
				})
			}
		}

		logger.Printf("Retrieving all blocks for epoch %v. %v epochs remaining", epoch, endEpoch-epoch)
	}

	return blocks, nil
}

func ExportEpoch(epoch uint64, client ethpb.BeaconChainClient) error {
	var err error

	data := &types.EpochData{}
	data.Epoch = epoch

	// Retrieve the validator balances for the epoch (NOTE: Currently the API call is broken and allows only to retrieve the balances for the current epoch
	data.ValidatorBalances = make([]*types.ValidatorBalance, 0)
	data.ValidatorIndices = make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: epoch}})
		if err != nil {
			logger.Printf("error retrieving validator balances response: %v", err)
			break
		}
		if validatorBalancesResponse.TotalSize == 0 {
			break
		}

		for _, balance := range validatorBalancesResponse.Balances {
			data.ValidatorBalances = append(data.ValidatorBalances, &types.ValidatorBalance{
				PublicKey: balance.PublicKey,
				Index:     balance.Index,
				Balance:   balance.Balance,
			})
			data.ValidatorIndices[utils.FormatPublicKey(balance.PublicKey)] = balance.Index
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}
	logger.Printf("Retrieved data for %v validator balances for epoch %v", len(data.ValidatorBalances), epoch)

	data.ValidatorAssignmentes, err = cache.GetEpochAssignments(epoch)
	if err != nil {
		return fmt.Errorf("error retrieving assignments for epoch %v: %v", epoch, err)
	}
	logger.Printf("Retrieved validator assignment data for epoch %v", epoch)

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.SlotsPerEpoch; slot <= (epoch+1)*utils.SlotsPerEpoch-1; slot++ {

		if slot == 0 { // Currently slot 0 returns all blocks
			continue
		}

		blocksResponse, err := client.ListBlocks(context.Background(), &ethpb.ListBlocksRequest{PageSize: utils.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Slot{Slot: slot}, IncludeNoncanonical: true})
		if err != nil {
			logger.Fatal(err)
		}

		if blocksResponse.TotalSize == 0 {
			continue
		}

		for _, block := range blocksResponse.BlockContainers {

			// Make sure that blocks from the genesis epoch have their Eth1Data field set
			if epoch == 0 && block.Block.Body.Eth1Data == nil {
				block.Block.Body.Eth1Data = &ethpb.Eth1Data{
					DepositRoot:  []byte{},
					DepositCount: 0,
					BlockHash:    []byte{},
				}
			}

			if data.Blocks[block.Block.Slot] == nil {
				data.Blocks[block.Block.Slot] = make(map[string]*types.Block)
			}

			b := &types.Block{
				Status:       1,
				Proposer:     data.ValidatorAssignmentes.ProposerAssignments[block.Block.Slot],
				BlockRoot:    block.BlockRoot,
				Slot:         block.Block.Slot,
				ParentRoot:   block.Block.ParentRoot,
				StateRoot:    block.Block.StateRoot,
				Signature:    block.Block.Signature,
				RandaoReveal: block.Block.Body.RandaoReveal,
				Graffiti:     block.Block.Body.Graffiti,
				Eth1Data: &types.Eth1Data{
					DepositRoot:  block.Block.Body.Eth1Data.DepositRoot,
					DepositCount: block.Block.Body.Eth1Data.DepositCount,
					BlockHash:    block.Block.Body.Eth1Data.BlockHash,
				},
				ProposerSlashings: make([]*types.ProposerSlashing, len(block.Block.Body.ProposerSlashings)),
				AttesterSlashings: make([]*types.AttesterSlashing, len(block.Block.Body.AttesterSlashings)),
				Attestations:      make([]*types.Attestation, len(block.Block.Body.Attestations)),
				Deposits:          make([]*types.Deposit, len(block.Block.Body.Deposits)),
				VoluntaryExits:    make([]*types.VoluntaryExit, len(block.Block.Body.VoluntaryExits)),
			}

			for i, proposerSlashing := range block.Block.Body.ProposerSlashings {
				b.ProposerSlashings[i] = &types.ProposerSlashing{
					ProposerIndex: proposerSlashing.ProposerIndex,
					Header_1: &types.Block{
						Slot:       proposerSlashing.Header_1.Slot,
						ParentRoot: proposerSlashing.Header_1.ParentRoot,
						StateRoot:  proposerSlashing.Header_1.StateRoot,
						Signature:  proposerSlashing.Header_1.Signature,
						BodyRoot:   proposerSlashing.Header_1.BodyRoot,
					},
					Header_2: &types.Block{
						Slot:       proposerSlashing.Header_2.Slot,
						ParentRoot: proposerSlashing.Header_2.ParentRoot,
						StateRoot:  proposerSlashing.Header_2.StateRoot,
						Signature:  proposerSlashing.Header_2.Signature,
						BodyRoot:   proposerSlashing.Header_2.BodyRoot,
					},
				}
			}

			for i, attesterSlashing := range block.Block.Body.AttesterSlashings {
				b.AttesterSlashings[i] = &types.AttesterSlashing{
					Attestation_1: &types.IndexedAttestation{
						CustodyBit_0Indices: attesterSlashing.Attestation_1.CustodyBit_0Indices,
						CustodyBit_1Indices: attesterSlashing.Attestation_1.CustodyBit_1Indices,
						Data: &types.AttestationData{
							Slot:            attesterSlashing.Attestation_1.Data.Slot,
							CommitteeIndex:  attesterSlashing.Attestation_1.Data.CommitteeIndex,
							BeaconBlockRoot: attesterSlashing.Attestation_1.Data.BeaconBlockRoot,
							Source: &types.Checkpoint{
								Epoch: attesterSlashing.Attestation_1.Data.Source.Epoch,
								Root:  attesterSlashing.Attestation_1.Data.Source.Root,
							},
							Target: &types.Checkpoint{
								Epoch: attesterSlashing.Attestation_1.Data.Target.Epoch,
								Root:  attesterSlashing.Attestation_1.Data.Target.Root,
							},
						},
						Signature: attesterSlashing.Attestation_1.Signature,
					},
					Attestation_2: &types.IndexedAttestation{
						CustodyBit_0Indices: attesterSlashing.Attestation_2.CustodyBit_0Indices,
						CustodyBit_1Indices: attesterSlashing.Attestation_2.CustodyBit_1Indices,
						Data: &types.AttestationData{
							Slot:            attesterSlashing.Attestation_2.Data.Slot,
							CommitteeIndex:  attesterSlashing.Attestation_2.Data.CommitteeIndex,
							BeaconBlockRoot: attesterSlashing.Attestation_2.Data.BeaconBlockRoot,
							Source: &types.Checkpoint{
								Epoch: attesterSlashing.Attestation_2.Data.Source.Epoch,
								Root:  attesterSlashing.Attestation_2.Data.Source.Root,
							},
							Target: &types.Checkpoint{
								Epoch: attesterSlashing.Attestation_2.Data.Target.Epoch,
								Root:  attesterSlashing.Attestation_2.Data.Target.Root,
							},
						},
						Signature: attesterSlashing.Attestation_2.Signature,
					},
				}
			}

			for i, attestation := range block.Block.Body.Attestations {
				b.Attestations[i] = &types.Attestation{
					AggregationBits: attestation.AggregationBits,
					Data: &types.AttestationData{
						Slot:            attestation.Data.Slot,
						CommitteeIndex:  attestation.Data.CommitteeIndex,
						BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
						Source: &types.Checkpoint{
							Epoch: attestation.Data.Source.Epoch,
							Root:  attestation.Data.Source.Root,
						},
						Target: &types.Checkpoint{
							Epoch: attestation.Data.Target.Epoch,
							Root:  attestation.Data.Target.Root,
						},
					},
					CustodyBits: attestation.CustodyBits,
					Signature:   attestation.Signature,
				}
			}
			for i, deposit := range block.Block.Body.Deposits {
				b.Deposits[i] = &types.Deposit{
					Proof:                 deposit.Proof,
					PublicKey:             deposit.Data.PublicKey,
					WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
					Amount:                deposit.Data.Amount,
					Signature:             deposit.Data.Signature,
				}
			}

			for i, voluntaryExit := range block.Block.Body.VoluntaryExits {
				b.VoluntaryExits[i] = &types.VoluntaryExit{
					Epoch:          voluntaryExit.Epoch,
					ValidatorIndex: voluntaryExit.ValidatorIndex,
					Signature:      voluntaryExit.Signature,
				}
			}

			data.Blocks[block.Block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = b
		}
	}
	logger.Printf("Retrieved %v blocks for epoch %v", len(data.Blocks), epoch)

	// Fill up missed and scheduled blocks
	for slot, proposer := range data.ValidatorAssignmentes.ProposerAssignments {
		_, found := data.Blocks[slot]
		if !found {
			// Proposer was assigned but did not yet propose a block
			data.Blocks[slot] = make(map[string]*types.Block)
			data.Blocks[slot]["0x0"] = &types.Block{
				Status:            0,
				Proposer:          proposer,
				BlockRoot:         []byte{0x0},
				Slot:              slot,
				ParentRoot:        []byte{},
				StateRoot:         []byte{},
				Signature:         []byte{},
				RandaoReveal:      []byte{},
				Graffiti:          []byte{},
				BodyRoot:          []byte{},
				Eth1Data:          &types.Eth1Data{},
				ProposerSlashings: make([]*types.ProposerSlashing, 0),
				AttesterSlashings: make([]*types.AttesterSlashing, 0),
				Attestations:      make([]*types.Attestation, 0),
				Deposits:          make([]*types.Deposit, 0),
				VoluntaryExits:    make([]*types.VoluntaryExit, 0),
			}

			if utils.SlotToTime(slot).After(time.Now().Add(time.Second * -60)) {
				// Block is in the future, set status to scheduled
				data.Blocks[slot]["0x0"].Status = 0
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x0}
			} else {
				// Block is in the past, set status to missed
				data.Blocks[slot]["0x0"].Status = 2
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x1}
			}
		} else {
			for _, block := range data.Blocks[slot] {
				block.Proposer = proposer
			}
		}
	}

	// Retrieve the validator set for the epoch
	data.Validators = make([]*types.Validator, 0)
	validatorResponse := &ethpb.Validators{}
	for {
		validatorResponse, err = client.ListValidators(context.Background(), &ethpb.ListValidatorsRequest{PageToken: validatorResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorsRequest_Epoch{Epoch: epoch}})
		if err != nil {
			logger.Printf("error retrieving validator response: %v", err)
			break
		}
		if validatorResponse.TotalSize == 0 {
			break
		}

		for _, validator := range validatorResponse.Validators {
			data.Validators = append(data.Validators, &types.Validator{
				PublicKey:                  validator.PublicKey,
				WithdrawalCredentials:      validator.WithdrawalCredentials,
				EffectiveBalance:           validator.EffectiveBalance,
				Slashed:                    validator.Slashed,
				ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
				ActivationEpoch:            validator.ActivationEpoch,
				ExitEpoch:                  validator.ExitEpoch,
				WithdrawableEpoch:          validator.WithdrawableEpoch,
			})
		}

		if validatorResponse.NextPageToken == "" {
			break
		}
	}
	logger.Printf("Retrieved validator data for epoch %v", epoch)

	// Retrieve the beacon committees for the epoch
	data.BeaconCommittees = make(map[uint64][]*types.BeaconCommitteItem)
	beaconCommitteesResponse := &ethpb.BeaconCommittees{}
	beaconCommitteesResponse, err = client.ListBeaconCommittees(context.Background(), &ethpb.ListCommitteesRequest{QueryFilter: &ethpb.ListCommitteesRequest_Epoch{Epoch: epoch}})
	if err != nil {
		logger.Printf("error retrieving beacon committees response: %v", err)
	} else {
		for slot, committee := range beaconCommitteesResponse.Committees {
			if committee == nil {
				continue
			}
			if data.BeaconCommittees[slot] == nil {
				data.BeaconCommittees[slot] = make([]*types.BeaconCommitteItem, 0)
			}

			for _, beaconCommittee := range committee.Committees {
				data.BeaconCommittees[slot] = append(data.BeaconCommittees[slot], &types.BeaconCommitteItem{ValidatorIndices: beaconCommittee.ValidatorIndices})
			}
		}
	}

	epochParticipationStatistics, err := client.GetValidatorParticipation(context.Background(), &ethpb.GetValidatorParticipationRequest{QueryFilter: &ethpb.GetValidatorParticipationRequest_Epoch{Epoch: epoch}})
	if err != nil {
		logger.Printf("error retrieving epoch participation statistics: %v", err)
		data.EpochParticipationStats = &types.ValidatorParticipation{
			Epoch:                   epoch,
			Finalized:               false,
			GlobalParticipationRate: 0,
			VotedEther:              0,
			EligibleEther:           0,
		}
	} else {
		data.EpochParticipationStats = &types.ValidatorParticipation{
			Epoch:                   epoch,
			Finalized:               epochParticipationStatistics.Finalized,
			GlobalParticipationRate: epochParticipationStatistics.Participation.GlobalParticipationRate,
			VotedEther:              epochParticipationStatistics.Participation.VotedEther,
			EligibleEther:           epochParticipationStatistics.Participation.EligibleEther,
		}
	}

	return db.SaveEpoch(data)
}

func exportAttestationPool(client ethpb.BeaconChainClient) error {
	attestationsResponse, err := client.AttestationPool(context.Background(), &ptypes.Empty{})

	if err != nil {
		return fmt.Errorf("error retrieving attestation pool data: %v", err)
	}

	logger.Printf("Retrieved %v attestations from the attestation pool", len(attestationsResponse.Attestations))

	attestations := make([]*types.Attestation, len(attestationsResponse.Attestations))
	for i, attestation := range attestationsResponse.Attestations {
		attestations[i] = &types.Attestation{
			AggregationBits: attestation.AggregationBits,
			Data: &types.AttestationData{
				Slot:            attestation.Data.Slot,
				CommitteeIndex:  attestation.Data.CommitteeIndex,
				BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
				Source: &types.Checkpoint{
					Epoch: attestation.Data.Source.Epoch,
					Root:  attestation.Data.Source.Root,
				},
				Target: &types.Checkpoint{
					Epoch: attestation.Data.Target.Epoch,
					Root:  attestation.Data.Target.Root,
				},
			},
			CustodyBits: attestation.CustodyBits,
			Signature:   attestation.Signature,
		}
	}
	return db.SaveAttestationPool(attestations)
}

func exportValidatorQueue(client ethpb.BeaconChainClient) error {
	var err error

	validatorIndices := make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize})
		if err != nil {
			logger.Printf("error retrieving validator balances response: %v", err)
			break
		}
		if validatorBalancesResponse.TotalSize == 0 {
			break
		}

		for _, balance := range validatorBalancesResponse.Balances {
			validatorIndices[utils.FormatPublicKey(balance.PublicKey)] = balance.Index
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}

	validators, err := client.GetValidatorQueue(context.Background(), &ptypes.Empty{})

	if err != nil {
		return fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	logger.Printf("Retrieved %v validators to enter and %v validators to leave from the validator queue", len(validators.ActivationPublicKeys), len(validators.ExitPublicKeys))

	return db.SaveValidatorQueue(&types.ValidatorQueue{
		ChurnLimit:           validators.ChurnLimit,
		ActivationPublicKeys: validators.ActivationPublicKeys,
		ExitPublicKeys:       validators.ExitPublicKeys,
	}, validatorIndices)
}

func updateEpochStatus(client ethpb.BeaconChainClient, startEpoch, endEpoch uint64) error {
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		epochParticipationStats, err := client.GetValidatorParticipation(context.Background(), &ethpb.GetValidatorParticipationRequest{QueryFilter: &ethpb.GetValidatorParticipationRequest_Epoch{Epoch: epoch}})
		if err != nil {
			logger.Printf("error retrieving epoch participation statistics: %v", err)
		} else {
			logger.Printf("Updating epoch %v with status finalized = %v", epoch, epochParticipationStats.Finalized)
			err := db.UpdateEpochStatus(&types.ValidatorParticipation{
				Epoch:                   epoch,
				Finalized:               epochParticipationStats.Finalized,
				GlobalParticipationRate: epochParticipationStats.Participation.GlobalParticipationRate,
				VotedEther:              epochParticipationStats.Participation.VotedEther,
				EligibleEther:           epochParticipationStats.Participation.EligibleEther,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}
