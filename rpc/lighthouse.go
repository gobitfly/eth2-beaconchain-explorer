package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/donovanhide/eventsource"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/errgroup"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prysmaticlabs/go-bitfield"
)

// LighthouseLatestHeadEpoch is used to cache the latest head epoch for participation requests
var LighthouseLatestHeadEpoch uint64 = 0

// LighthouseClient holds the Lighthouse client info
type LighthouseClient struct {
	endpoint            string
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
	slotsCache          *lru.Cache
	slotsCacheMux       *sync.Mutex
	signer              gtypes.Signer
}

// NewLighthouseClient is used to create a new Lighthouse client
func NewLighthouseClient(endpoint string, chainID *big.Int) (*LighthouseClient, error) {
	signer := gtypes.NewCancunSigner(chainID)
	client := &LighthouseClient{
		endpoint:            endpoint,
		assignmentsCacheMux: &sync.Mutex{},
		slotsCacheMux:       &sync.Mutex{},
		signer:              signer,
	}
	client.assignmentsCache, _ = lru.New(10)
	client.slotsCache, _ = lru.New(128) // cache at most 128 slots

	return client, nil
}

func (lc *LighthouseClient) GetNewBlockChan() chan *types.Block {
	blkCh := make(chan *types.Block, 10)
	go func() {
		stream, err := eventsource.Subscribe(fmt.Sprintf("%s/eth/v1/events?topics=head", lc.endpoint), "")

		if err != nil {
			utils.LogFatal(err, "getting eventsource stream error", 0)
		}
		defer stream.Close()

		for {
			select {
			// It is important to register to Errors, otherwise the stream does not reconnect if the connection was lost
			case err := <-stream.Errors:
				utils.LogError(err, "Lighthouse connection error (will automatically retry to connect)", 0)
			case e := <-stream.Events:
				// logger.Infof("retrieved %v via event stream", e.Data())
				var parsed StreamedBlockEventData
				err = json.Unmarshal([]byte(e.Data()), &parsed)
				if err != nil {
					logger.Warnf("failed to decode block event: %v", err)
					continue
				}

				logger.Infof("retrieving data for slot %v", parsed.Slot)
				block, err := lc.GetBlockBySlot(uint64(parsed.Slot))
				if err != nil {
					logger.Warnf("failed to fetch block for slot %d: %v", uint64(parsed.Slot), err)
					continue
				}
				logger.Infof("retrieved block for slot %v", parsed.Slot)
				// logger.Infof("pushing block %v", blk.Slot)
				blkCh <- block
			}
		}
	}()
	return blkCh
}

// GetChainHead gets the chain head from Lighthouse
func (lc *LighthouseClient) GetChainHead() (*types.ChainHead, error) {
	headResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/head", lc.endpoint))
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain head: %w", err)
	}

	var parsedHead StandardBeaconHeaderResponse
	err = json.Unmarshal(headResp, &parsedHead)
	if err != nil {
		return nil, fmt.Errorf("error parsing chain head: %w", err)
	}

	id := parsedHead.Data.Header.Message.StateRoot
	if parsedHead.Data.Header.Message.Slot == 0 {
		id = "genesis"
	}
	finalityResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", lc.endpoint, id))
	if err != nil {
		return nil, fmt.Errorf("error retrieving finality checkpoints of head: %w", err)
	}

	var parsedFinality StandardFinalityCheckpointsResponse
	err = json.Unmarshal(finalityResp, &parsedFinality)
	if err != nil {
		return nil, fmt.Errorf("error parsing finality checkpoints of head: %w", err)
	}

	// The epoch in the Finalized Object is not the finalized epoch, but the epoch for the checkpoint - the 'real' finalized epoch is the one before
	var finalizedEpoch = uint64(parsedFinality.Data.Finalized.Epoch)
	if finalizedEpoch > 0 {
		finalizedEpoch--
	}

	finalizedSlot := (finalizedEpoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch // The first Slot of the next epoch is finalized.
	if finalizedEpoch == 0 && parsedFinality.Data.Finalized.Root == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		finalizedSlot = 0
	}
	return &types.ChainHead{
		HeadSlot:                   uint64(parsedHead.Data.Header.Message.Slot),
		HeadEpoch:                  uint64(parsedHead.Data.Header.Message.Slot) / utils.Config.Chain.ClConfig.SlotsPerEpoch,
		HeadBlockRoot:              utils.MustParseHex(parsedHead.Data.Root),
		FinalizedSlot:              finalizedSlot,
		FinalizedEpoch:             finalizedEpoch,
		FinalizedBlockRoot:         utils.MustParseHex(parsedFinality.Data.Finalized.Root),
		JustifiedSlot:              uint64(parsedFinality.Data.CurrentJustified.Epoch) * utils.Config.Chain.ClConfig.SlotsPerEpoch,
		JustifiedEpoch:             uint64(parsedFinality.Data.CurrentJustified.Epoch),
		JustifiedBlockRoot:         utils.MustParseHex(parsedFinality.Data.CurrentJustified.Root),
		PreviousJustifiedSlot:      uint64(parsedFinality.Data.PreviousJustified.Epoch) * utils.Config.Chain.ClConfig.SlotsPerEpoch,
		PreviousJustifiedEpoch:     uint64(parsedFinality.Data.PreviousJustified.Epoch),
		PreviousJustifiedBlockRoot: utils.MustParseHex(parsedFinality.Data.PreviousJustified.Root),
	}, nil
}

func (lc *LighthouseClient) GetValidatorQueue() (*types.ValidatorQueue, error) {
	// pre-filter the status, to return much less validators, thus much faster!
	validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/head/validators?status=pending_queued,active_exiting,active_slashed", lc.endpoint))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator for head valiqdator queue check: %w", err)
	}

	var parsedValidators StandardValidatorsResponse
	err = json.Unmarshal(validatorsResp, &parsedValidators)
	if err != nil {
		return nil, fmt.Errorf("error parsing queue validators: %w", err)
	}
	// TODO: maybe track more status counts in the future?
	statusMap := make(map[string]uint64)

	for _, validator := range parsedValidators.Data {
		statusMap[validator.Status] += 1
	}
	return &types.ValidatorQueue{
		Activating: statusMap["pending_queued"],
		Exiting:    statusMap["active_exiting"] + statusMap["active_slashed"],
	}, nil
}

// GetEpochAssignments will get the epoch assignments from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {

	var err error

	lc.assignmentsCacheMux.Lock()
	cachedValue, found := lc.assignmentsCache.Get(epoch)
	if found {
		lc.assignmentsCacheMux.Unlock()
		return cachedValue.(*types.EpochAssignments), nil
	}
	lc.assignmentsCacheMux.Unlock()

	proposerResp, err := lc.get(fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", lc.endpoint, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving proposer duties for epoch %v: %w", epoch, err)
	}
	var parsedProposerResponse StandardProposerDutiesResponse
	err = json.Unmarshal(proposerResp, &parsedProposerResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing proposer duties: %w", err)
	}

	// fetch the block root that the proposer data is dependent on
	headerResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/%s", lc.endpoint, parsedProposerResponse.DependentRoot))
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain header: %w", err)
	}
	var parsedHeader StandardBeaconHeaderResponse
	err = json.Unmarshal(headerResp, &parsedHeader)
	if err != nil {
		return nil, fmt.Errorf("error parsing chain header: %w", err)
	}
	depStateRoot := parsedHeader.Data.Header.Message.StateRoot

	// Now use the state root to make a consistent committee query
	committeesResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/committees?epoch=%d", lc.endpoint, depStateRoot, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving committees data: %w", err)
	}
	var parsedCommittees StandardCommitteesResponse
	err = json.Unmarshal(committeesResp, &parsedCommittees)
	if err != nil {
		return nil, fmt.Errorf("error parsing committees data: %w", err)
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// propose
	for _, duty := range parsedProposerResponse.Data {
		assignments.ProposerAssignments[uint64(duty.Slot)] = uint64(duty.ValidatorIndex)
	}

	// attest
	for _, committee := range parsedCommittees.Data {
		for i, valIndex := range committee.Validators {
			valIndexU64, err := strconv.ParseUint(valIndex, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("epoch %d committee %d index %d has bad validator index %q", epoch, committee.Index, i, valIndex)
			}
			k := utils.FormatAttestorAssignmentKey(uint64(committee.Slot), uint64(committee.Index), uint64(i))
			assignments.AttestorAssignments[k] = valIndexU64
		}
	}

	if epoch >= utils.Config.Chain.ClConfig.AltairForkEpoch {
		syncCommitteeState := depStateRoot
		if epoch == utils.Config.Chain.ClConfig.AltairForkEpoch {
			syncCommitteeState = fmt.Sprintf("%d", utils.Config.Chain.ClConfig.AltairForkEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
		parsedSyncCommittees, err := lc.GetSyncCommittee(syncCommitteeState, epoch)
		if err != nil {
			return nil, err
		}
		assignments.SyncAssignments = make([]uint64, len(parsedSyncCommittees.Validators))

		// sync
		for i, valIndexStr := range parsedSyncCommittees.Validators {
			valIndexU64, err := strconv.ParseUint(valIndexStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("in sync_committee for epoch %d validator %d has bad validator index: %q", epoch, i, valIndexStr)
			}
			assignments.SyncAssignments[i] = valIndexU64
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		lc.assignmentsCacheMux.Lock()
		lc.assignmentsCache.Add(epoch, assignments)
		lc.assignmentsCacheMux.Unlock()
	}

	return assignments, nil
}

// GetEpochProposerAssignments will get the epoch proposer assignments from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochProposerAssignments(epoch uint64) (*StandardProposerDutiesResponse, error) {
	var err error

	proposerResp, err := lc.get(fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", lc.endpoint, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving proposer duties for epoch %v: %w", epoch, err)
	}
	parsedProposerResponse := &StandardProposerDutiesResponse{}
	err = json.Unmarshal(proposerResp, parsedProposerResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing proposer duties: %w", err)
	}

	return parsedProposerResponse, nil
}

func (lc *LighthouseClient) GetValidatorState(epoch uint64) (*StandardValidatorsResponse, error) {
	validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", lc.endpoint, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch))
	if err != nil && epoch == 0 {
		validatorsResp, err = lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", lc.endpoint, "genesis"))
		if err != nil {
			return nil, fmt.Errorf("error retrieving validators for genesis: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
	}

	parsedValidators := &StandardValidatorsResponse{}
	err = json.Unmarshal(validatorsResp, parsedValidators)
	if err != nil {
		return nil, fmt.Errorf("error parsing epoch validators: %w", err)
	}

	return parsedValidators, nil
}

// GetEpochData will get the epoch data from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error) {
	wg := &errgroup.Group{}
	mux := &sync.Mutex{}

	head, err := lc.GetChainHead()
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain head: %w", err)
	}
	data := &types.EpochData{
		SyncDuties:        make(map[types.Slot]map[types.ValidatorIndex]bool),
		AttestationDuties: make(map[types.Slot]map[types.ValidatorIndex][]types.Slot),
	}
	data.Epoch = epoch
	if head.FinalizedEpoch >= epoch {
		data.Finalized = true
	}

	if head.FinalizedEpoch == 0 && epoch == 0 {
		data.Finalized = false
	}

	validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", lc.endpoint, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch))
	if err != nil && epoch == 0 {
		validatorsResp, err = lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", lc.endpoint, "genesis"))
		if err != nil {
			return nil, fmt.Errorf("error retrieving validators for genesis: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
	}

	var parsedValidators StandardValidatorsResponse
	err = json.Unmarshal(validatorsResp, &parsedValidators)
	if err != nil {
		return nil, fmt.Errorf("error parsing epoch validators: %w", err)
	}

	for _, validator := range parsedValidators.Data {
		data.Validators = append(data.Validators, &types.Validator{
			Index:                      uint64(validator.Index),
			PublicKey:                  utils.MustParseHex(validator.Validator.Pubkey),
			WithdrawalCredentials:      utils.MustParseHex(validator.Validator.WithdrawalCredentials),
			Balance:                    uint64(validator.Balance),
			EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
			ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
			ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
			WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
			Status:                     validator.Status,
		})
	}

	logger.Printf("retrieved data for %v validators for epoch %v", len(data.Validators), epoch)

	wg.Go(func() error {
		var err error
		data.ValidatorAssignmentes, err = lc.GetEpochAssignments(epoch)
		if err != nil {
			return fmt.Errorf("error retrieving assignments for epoch %v: %w", epoch, err)
		}

		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if data.SyncDuties[types.Slot(slot)] == nil {
				data.SyncDuties[types.Slot(slot)] = make(map[types.ValidatorIndex]bool)
			}
			for _, validatorIndex := range data.ValidatorAssignmentes.SyncAssignments {
				data.SyncDuties[types.Slot(slot)][types.ValidatorIndex(validatorIndex)] = false
			}
		}

		for key, validatorIndex := range data.ValidatorAssignmentes.AttestorAssignments {
			keySplit := strings.Split(key, "-")
			attestedSlot, err := strconv.ParseUint(keySplit[0], 10, 64)

			if err != nil {
				return fmt.Errorf("error parsing attested slot from attestation key: %w", err)
			}

			if data.AttestationDuties[types.Slot(attestedSlot)] == nil {
				data.AttestationDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}

			data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validatorIndex)] = []types.Slot{}

		}
		logger.Printf("retrieved validator assignment data for epoch %v", epoch)
		return nil
	})

	if epoch < head.HeadEpoch {
		wg.Go(func() error {
			var err error
			data.EpochParticipationStats, err = lc.GetValidatorParticipation(epoch)
			if err != nil {
				if strings.HasSuffix(err.Error(), "can't be retrieved as it hasn't finished yet") { // should no longer happen
					logger.Warnf("error retrieving epoch participation statistics for epoch %v: %v", epoch, err)
				} else {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				}
				data.EpochParticipationStats = &types.ValidatorParticipation{
					Epoch:                   epoch,
					GlobalParticipationRate: 0.0,
					VotedEther:              0,
					EligibleEther:           0,
				}
			}
			return nil
		})
	} else {
		data.EpochParticipationStats = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: 0.0,
			VotedEther:              0,
			EligibleEther:           0,
		}
	}

	err = wg.Wait()
	if err != nil {
		return nil, err
	}
	wg = &errgroup.Group{}
	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	wg.Go(func() error {
		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if slot != 0 && slot > head.HeadSlot { // don't export slots that have not occured yet
				continue
			}
			start := time.Now()
			block, err := lc.GetBlockBySlot(slot)

			if err != nil {
				return fmt.Errorf("error retrieving block for slot %v: %w", slot, err)
			}

			mux.Lock()
			if data.Blocks[block.Slot] == nil {
				data.Blocks[block.Slot] = make(map[string]*types.Block)
			}
			data.Blocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

			for validator, duty := range block.SyncDuties {
				data.SyncDuties[types.Slot(block.Slot)][types.ValidatorIndex(validator)] = duty
			}
			for validator, attestedSlots := range block.AttestationDuties {
				for _, attestedSlot := range attestedSlots {
					if data.AttestationDuties[types.Slot(attestedSlot)] == nil {
						data.AttestationDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
					}
					if data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] == nil {
						data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = make([]types.Slot, 0, 10)
					}
					data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = append(data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)], types.Slot(block.Slot))
				}
			}
			mux.Unlock()
			logger.Infof("processed data for current epoch slot %v in %v", slot, time.Since(start))

		}
		return nil
	})

	// we need future blocks to properly tracke fullfilled attestation duties
	data.FutureBlocks = make(map[uint64]map[string]*types.Block)
	wg.Go(func() error {
		for slot := (epoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+2)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if slot != 0 && slot > head.HeadSlot { // don't export slots that have not occured yet
				continue
			}
			start := time.Now()
			block, err := lc.GetBlockBySlot(slot)

			if err != nil {
				return fmt.Errorf("error retrieving block for slot %v: %w", slot, err)
			}

			mux.Lock()
			if data.FutureBlocks[block.Slot] == nil {
				data.FutureBlocks[block.Slot] = make(map[string]*types.Block)
			}
			data.FutureBlocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

			// fill out performed attestation duties
			for validator, attestedSlots := range block.AttestationDuties {
				for _, attestedSlot := range attestedSlots {
					if attestedSlot < types.Slot((epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch) {
						data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = append(data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)], types.Slot(block.Slot))
					}
				}
			}
			mux.Unlock()
			logger.Infof("processed data for next epoch slot %v in %v", slot, time.Since(start))

		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	logger.Printf("retrieved %v blocks for epoch %v", len(data.Blocks), epoch)

	if data.ValidatorAssignmentes == nil {
		return data, fmt.Errorf("no assignments for epoch %v", epoch)
	}

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
				SyncAggregate:     nil,
			}

			if utils.SlotToTime(slot).After(time.Now().Add(time.Second * -4)) {
				// Block is in the future, set status to scheduled
				data.Blocks[slot]["0x0"].Status = 0
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x0}
			} else {
				// Block is in the past, set status to missed
				data.Blocks[slot]["0x0"].Status = 2
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x1}
			}
		}
	}

	// for validator, duty := range data.AttestationDuties {
	// 	for slot, inclusions := range duty {
	// 		logger.Infof("validator %v has attested for slot %v in slots %v", validator, slot, inclusions)
	// 	}
	// }

	return data, nil
}

func uint64List(li []uint64Str) []uint64 {
	out := make([]uint64, len(li))
	for i, v := range li {
		out[i] = uint64(v)
	}
	return out
}

func (lc *LighthouseClient) GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error) {

	if epoch < 0 {
		epoch = 0
	}

	var err error

	validatorBalances := make(map[uint64]uint64)

	resp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validator_balances", lc.endpoint, epoch*int64(utils.Config.Chain.ClConfig.SlotsPerEpoch)))
	if err != nil && epoch == 0 {
		resp, err = lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/genesis/validator_balances", lc.endpoint))
		if err != nil {
			return validatorBalances, err
		}
	} else if err != nil {
		return validatorBalances, err
	}

	var parsedResponse StandardValidatorBalancesResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing response for validator_balances")
	}

	for _, b := range parsedResponse.Data {
		validatorBalances[uint64(b.Index)] = uint64(b.Balance)
	}

	return validatorBalances, nil
}

func (lc *LighthouseClient) GetBlockByBlockroot(blockroot []byte) (*types.Block, error) {
	resHeaders, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/0x%x", lc.endpoint, blockroot))
	if err != nil {
		if err == errNotFound {
			// no block found
			return &types.Block{}, nil
		}
		return nil, fmt.Errorf("error retrieving headers for blockroot 0x%x: %w", blockroot, err)
	}
	var parsedHeaders StandardBeaconHeaderResponse
	err = json.Unmarshal(resHeaders, &parsedHeaders)
	if err != nil {
		return nil, fmt.Errorf("error parsing header-response for blockroot 0x%x: %w", blockroot, err)
	}

	slot := uint64(parsedHeaders.Data.Header.Message.Slot)

	resp, err := lc.get(fmt.Sprintf("%s/eth/v2/beacon/blocks/%s", lc.endpoint, parsedHeaders.Data.Root))
	if err != nil {
		return nil, fmt.Errorf("error retrieving block data at slot %v: %w", slot, err)
	}

	var parsedResponse StandardV2BlockResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		logger.Errorf("error parsing block data at slot %v: %v", parsedHeaders.Data.Header.Message.Slot, err)
		return nil, fmt.Errorf("error parsing block-response at slot %v: %w", slot, err)
	}

	return lc.blockFromResponse(&parsedHeaders, &parsedResponse)
}

// GetBlockHeader will get the block header by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlockHeader(slot uint64) (*StandardBeaconHeaderResponse, error) {
	var parsedHeaders *StandardBeaconHeaderResponse

	resHeaders, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/%d", lc.endpoint, slot))
	if err != nil && slot == 0 {
		headResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers", lc.endpoint))
		if err != nil {
			return nil, fmt.Errorf("error retrieving chain head for slot %v: %w", slot, err)
		}

		var parsedHeader StandardBeaconHeadersResponse
		err = json.Unmarshal(headResp, &parsedHeader)
		if err != nil {
			return nil, fmt.Errorf("error parsing chain head for slot %v: %w", slot, err)
		}

		if len(parsedHeader.Data) == 0 {
			return nil, fmt.Errorf("error no headers available for slot %v", slot)
		}

		parsedHeaders = &StandardBeaconHeaderResponse{
			Data: parsedHeader.Data[len(parsedHeader.Data)-1],
		}

	} else if err != nil {
		if err == errNotFound { // return dummy block for missed slots
			// no block found
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving headers at slot %v: %w", slot, err)
	}

	if parsedHeaders == nil {
		err = json.Unmarshal(resHeaders, &parsedHeaders)
		if err != nil {
			return nil, fmt.Errorf("error parsing header-response at slot %v: %w", slot, err)
		}
	}

	return parsedHeaders, nil
}

// GetBlocksBySlot will get the blocks by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlockBySlot(slot uint64) (*types.Block, error) {
	epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	isFirstSlotOfEpoch := slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0

	var parsedHeaders *StandardBeaconHeaderResponse

	resHeaders, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/%d", lc.endpoint, slot))
	if err != nil && slot == 0 {
		headResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers", lc.endpoint))
		if err != nil {
			return nil, fmt.Errorf("error retrieving chain head for slot %v: %w", slot, err)
		}

		var parsedHeader StandardBeaconHeadersResponse
		err = json.Unmarshal(headResp, &parsedHeader)
		if err != nil {
			return nil, fmt.Errorf("error parsing chain head for slot %v: %w", slot, err)
		}

		if len(parsedHeader.Data) == 0 {
			return nil, fmt.Errorf("error no headers available for slot %v", slot)
		}

		parsedHeaders = &StandardBeaconHeaderResponse{
			Data: parsedHeader.Data[len(parsedHeader.Data)-1],
		}

	} else if err != nil {
		if err == errNotFound { // return dummy block for missed slots
			proposerAssignments, err := lc.GetEpochProposerAssignments(epoch)
			if err != nil {
				return nil, err
			}

			proposer := uint64(math.MaxUint64)
			for _, pa := range proposerAssignments.Data {
				if uint64(pa.Slot) == slot {
					proposer = uint64(pa.ValidatorIndex)
				}
			}

			block := &types.Block{
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
				SyncAggregate:     nil,
			}

			if isFirstSlotOfEpoch {
				assignments, err := lc.GetEpochAssignments(epoch)
				if err != nil {
					return nil, err
				}

				block.EpochAssignments = assignments

				validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", lc.endpoint, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch))
				if err != nil && epoch == 0 {
					validatorsResp, err = lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", lc.endpoint, "genesis"))
					if err != nil {
						return nil, fmt.Errorf("error retrieving validators for genesis: %w", err)
					}
				} else if err != nil {
					return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
				}

				var parsedValidators StandardValidatorsResponse
				err = json.Unmarshal(validatorsResp, &parsedValidators)
				if err != nil {
					return nil, fmt.Errorf("error parsing epoch validators: %w", err)
				}

				block.Validators = make([]*types.Validator, 0, len(parsedValidators.Data))
				for _, validator := range parsedValidators.Data {
					block.Validators = append(block.Validators, &types.Validator{
						Index:                      uint64(validator.Index),
						PublicKey:                  utils.MustParseHex(validator.Validator.Pubkey),
						WithdrawalCredentials:      utils.MustParseHex(validator.Validator.WithdrawalCredentials),
						Balance:                    uint64(validator.Balance),
						EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
						Slashed:                    validator.Validator.Slashed,
						ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
						ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
						ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
						WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
						Status:                     validator.Status,
					})
				}
			}

			return block, nil
		}
		return nil, fmt.Errorf("error retrieving headers at slot %v: %w", slot, err)
	}

	if parsedHeaders == nil {
		err = json.Unmarshal(resHeaders, &parsedHeaders)
		if err != nil {
			return nil, fmt.Errorf("error parsing header-response at slot %v: %w", slot, err)
		}
	}

	lc.slotsCacheMux.Lock()
	cachedBlock, ok := lc.slotsCache.Get(parsedHeaders.Data.Root)
	if ok {
		lc.slotsCacheMux.Unlock()
		block, ok := cachedBlock.(*types.Block)

		if ok {
			logger.Infof("retrieved slot %v (0x%x) from in memory cache", block.Slot, block.BlockRoot)
			return block, nil
		} else {
			logger.Errorf("unable to convert cached block to block type")
		}
	}
	lc.slotsCacheMux.Unlock()

	resp, err := lc.get(fmt.Sprintf("%s/eth/v2/beacon/blocks/%s", lc.endpoint, parsedHeaders.Data.Root))
	if err != nil && slot == 0 {
		return nil, fmt.Errorf("error retrieving block data at slot %v: %w", slot, err)
	}

	var parsedResponse StandardV2BlockResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		logger.Errorf("error parsing block data at slot %v: %v", slot, err)
		return nil, fmt.Errorf("error parsing block-response at slot %v: %w", slot, err)
	}

	block, err := lc.blockFromResponse(parsedHeaders, &parsedResponse)
	if err != nil {
		return nil, err
	}

	// for the first slot of an epoch, also retrieve the epoch assignments
	if block.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 {
		var err error
		block.EpochAssignments, err = lc.GetEpochAssignments(block.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
		if err != nil {
			return nil, err
		}

		validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", lc.endpoint, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch))
		if err != nil && epoch == 0 {
			validatorsResp, err = lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", lc.endpoint, "genesis"))
			if err != nil {
				return nil, fmt.Errorf("error retrieving validators for genesis: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
		}

		var parsedValidators StandardValidatorsResponse
		err = json.Unmarshal(validatorsResp, &parsedValidators)
		if err != nil {
			return nil, fmt.Errorf("error parsing epoch validators: %w", err)
		}

		block.Validators = make([]*types.Validator, 0, len(parsedValidators.Data))
		for _, validator := range parsedValidators.Data {
			block.Validators = append(block.Validators, &types.Validator{
				Index:                      uint64(validator.Index),
				PublicKey:                  utils.MustParseHex(validator.Validator.Pubkey),
				WithdrawalCredentials:      utils.MustParseHex(validator.Validator.WithdrawalCredentials),
				Balance:                    uint64(validator.Balance),
				EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
				Slashed:                    validator.Validator.Slashed,
				ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
				ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
				ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
				WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
				Status:                     validator.Status,
			})
		}
	}

	lc.slotsCacheMux.Lock()
	lc.slotsCache.Add(parsedHeaders.Data.Root, block)
	lc.slotsCacheMux.Unlock()

	return block, nil
}

func (lc *LighthouseClient) blockFromResponse(parsedHeaders *StandardBeaconHeaderResponse, parsedResponse *StandardV2BlockResponse) (*types.Block, error) {
	parsedBlock := parsedResponse.Data
	slot := uint64(parsedHeaders.Data.Header.Message.Slot)
	block := &types.Block{
		Status:       1,
		Finalized:    parsedHeaders.Finalized,
		Proposer:     uint64(parsedBlock.Message.ProposerIndex),
		BlockRoot:    utils.MustParseHex(parsedHeaders.Data.Root),
		Slot:         slot,
		ParentRoot:   utils.MustParseHex(parsedBlock.Message.ParentRoot),
		StateRoot:    utils.MustParseHex(parsedBlock.Message.StateRoot),
		Signature:    parsedBlock.Signature,
		RandaoReveal: utils.MustParseHex(parsedBlock.Message.Body.RandaoReveal),
		Graffiti:     utils.MustParseHex(parsedBlock.Message.Body.Graffiti),
		Eth1Data: &types.Eth1Data{
			DepositRoot:  utils.MustParseHex(parsedBlock.Message.Body.Eth1Data.DepositRoot),
			DepositCount: uint64(parsedBlock.Message.Body.Eth1Data.DepositCount),
			BlockHash:    utils.MustParseHex(parsedBlock.Message.Body.Eth1Data.BlockHash),
		},
		ProposerSlashings:          make([]*types.ProposerSlashing, len(parsedBlock.Message.Body.ProposerSlashings)),
		AttesterSlashings:          make([]*types.AttesterSlashing, len(parsedBlock.Message.Body.AttesterSlashings)),
		Attestations:               make([]*types.Attestation, len(parsedBlock.Message.Body.Attestations)),
		Deposits:                   make([]*types.Deposit, len(parsedBlock.Message.Body.Deposits)),
		VoluntaryExits:             make([]*types.VoluntaryExit, len(parsedBlock.Message.Body.VoluntaryExits)),
		SignedBLSToExecutionChange: make([]*types.SignedBLSToExecutionChange, len(parsedBlock.Message.Body.SignedBLSToExecutionChange)),
		BlobKZGCommitments:         make([][]byte, len(parsedBlock.Message.Body.BlobKZGCommitments)),
		BlobKZGProofs:              make([][]byte, len(parsedBlock.Message.Body.BlobKZGCommitments)),
		AttestationDuties:          make(map[types.ValidatorIndex][]types.Slot),
		SyncDuties:                 make(map[types.ValidatorIndex]bool),
	}

	for i, c := range parsedBlock.Message.Body.BlobKZGCommitments {
		block.BlobKZGCommitments[i] = c
	}

	if len(parsedBlock.Message.Body.BlobKZGCommitments) > 0 {
		res, err := lc.GetBlobSidecars(fmt.Sprintf("%#x", block.BlockRoot))
		if err != nil {
			return nil, err
		}
		if len(res.Data) != len(parsedBlock.Message.Body.BlobKZGCommitments) {
			return nil, fmt.Errorf("error constructing block at slot %v: len(blob_sidecars) != len(block.blob_kzg_commitments): %v != %v", block.Slot, len(res.Data), len(parsedBlock.Message.Body.BlobKZGCommitments))
		}
		for i, d := range res.Data {
			if !bytes.Equal(d.KzgCommitment, block.BlobKZGCommitments[i]) {
				return nil, fmt.Errorf("error constructing block at slot %v: unequal kzg_commitments at index %v: %#x != %#x", block.Slot, i, d.KzgCommitment, block.BlobKZGCommitments[i])
			}
			block.BlobKZGProofs[i] = d.KzgProof
		}
	}

	epochAssignments, err := lc.GetEpochAssignments(slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return nil, err
	}

	if agg := parsedBlock.Message.Body.SyncAggregate; agg != nil {
		bits := utils.MustParseHex(agg.SyncCommitteeBits)

		if utils.Config.Chain.ClConfig.SyncCommitteeSize != uint64(len(bits)*8) {
			return nil, fmt.Errorf("sync-aggregate-bits-size does not match sync-committee-size: %v != %v", len(bits)*8, utils.Config.Chain.ClConfig.SyncCommitteeSize)
		}

		block.SyncAggregate = &types.SyncAggregate{
			SyncCommitteeValidators:    epochAssignments.SyncAssignments,
			SyncCommitteeBits:          bits,
			SyncAggregateParticipation: syncCommitteeParticipation(bits),
			SyncCommitteeSignature:     utils.MustParseHex(agg.SyncCommitteeSignature),
		}

		// fill out performed sync duties
		bitLen := len(block.SyncAggregate.SyncCommitteeBits) * 8
		valLen := len(block.SyncAggregate.SyncCommitteeValidators)
		if bitLen < valLen {
			return nil, fmt.Errorf("error getting sync_committee participants: bitLen != valLen: %v != %v", bitLen, valLen)
		}
		for i, valIndex := range block.SyncAggregate.SyncCommitteeValidators {
			block.SyncDuties[types.ValidatorIndex(valIndex)] = utils.BitAtVector(block.SyncAggregate.SyncCommitteeBits, i)
		}
	}

	if payload := parsedBlock.Message.Body.ExecutionPayload; payload != nil && !bytes.Equal(payload.ParentHash, make([]byte, 32)) {
		txs := make([]*types.Transaction, 0, len(payload.Transactions))
		for i, rawTx := range payload.Transactions {
			tx := &types.Transaction{Raw: rawTx}
			var decTx gtypes.Transaction
			if err := decTx.UnmarshalBinary(rawTx); err != nil {
				return nil, fmt.Errorf("error parsing tx %d block %x: %w", i, payload.BlockHash, err)
			} else {
				h := decTx.Hash()
				tx.TxHash = h[:]
				tx.AccountNonce = decTx.Nonce()
				// big endian
				tx.Price = decTx.GasPrice().Bytes()
				tx.GasLimit = decTx.Gas()
				sender, err := lc.signer.Sender(&decTx)
				if err != nil {
					return nil, fmt.Errorf("transaction with invalid sender (slot: %v, tx-hash: %x): %w", slot, h, err)
				}
				tx.Sender = sender.Bytes()
				if v := decTx.To(); v != nil {
					tx.Recipient = v.Bytes()
				} else {
					tx.Recipient = []byte{}
				}
				tx.Amount = decTx.Value().Bytes()
				tx.Payload = decTx.Data()
				tx.MaxPriorityFeePerGas = decTx.GasTipCap().Uint64()
				tx.MaxFeePerGas = decTx.GasFeeCap().Uint64()

				if decTx.BlobGasFeeCap() != nil {
					tx.MaxFeePerBlobGas = decTx.BlobGasFeeCap().Uint64()
				}
				for _, h := range decTx.BlobHashes() {
					tx.BlobVersionedHashes = append(tx.BlobVersionedHashes, h.Bytes())
				}
			}
			txs = append(txs, tx)
		}
		withdrawals := make([]*types.Withdrawals, 0, len(payload.Withdrawals))
		for _, w := range payload.Withdrawals {
			withdrawals = append(withdrawals, &types.Withdrawals{
				Index:          uint64(w.Index),
				ValidatorIndex: uint64(w.ValidatorIndex),
				Address:        w.Address,
				Amount:         uint64(w.Amount),
			})
		}

		block.ExecutionPayload = &types.ExecutionPayload{
			ParentHash:    payload.ParentHash,
			FeeRecipient:  payload.FeeRecipient,
			StateRoot:     payload.StateRoot,
			ReceiptsRoot:  payload.ReceiptsRoot,
			LogsBloom:     payload.LogsBloom,
			Random:        payload.PrevRandao,
			BlockNumber:   uint64(payload.BlockNumber),
			GasLimit:      uint64(payload.GasLimit),
			GasUsed:       uint64(payload.GasUsed),
			Timestamp:     uint64(payload.Timestamp),
			ExtraData:     payload.ExtraData,
			BaseFeePerGas: uint64(payload.BaseFeePerGas),
			BlockHash:     payload.BlockHash,
			Transactions:  txs,
			Withdrawals:   withdrawals,
			BlobGasUsed:   uint64(payload.BlobGasUsed),
			ExcessBlobGas: uint64(payload.ExcessBlobGas),
		}
	}

	// TODO: this is legacy from old lighthouse API. Does it even still apply?
	if block.Eth1Data.DepositCount > 2147483647 { // Sometimes the lighthouse node does return bogus data for the DepositCount value
		block.Eth1Data.DepositCount = 0
	}

	for i, proposerSlashing := range parsedBlock.Message.Body.ProposerSlashings {
		block.ProposerSlashings[i] = &types.ProposerSlashing{
			ProposerIndex: uint64(proposerSlashing.SignedHeader1.Message.ProposerIndex),
			Header1: &types.Block{
				Slot:       uint64(proposerSlashing.SignedHeader1.Message.Slot),
				ParentRoot: utils.MustParseHex(proposerSlashing.SignedHeader1.Message.ParentRoot),
				StateRoot:  utils.MustParseHex(proposerSlashing.SignedHeader1.Message.StateRoot),
				Signature:  utils.MustParseHex(proposerSlashing.SignedHeader1.Signature),
				BodyRoot:   utils.MustParseHex(proposerSlashing.SignedHeader1.Message.BodyRoot),
			},
			Header2: &types.Block{
				Slot:       uint64(proposerSlashing.SignedHeader2.Message.Slot),
				ParentRoot: utils.MustParseHex(proposerSlashing.SignedHeader2.Message.ParentRoot),
				StateRoot:  utils.MustParseHex(proposerSlashing.SignedHeader2.Message.StateRoot),
				Signature:  utils.MustParseHex(proposerSlashing.SignedHeader2.Signature),
				BodyRoot:   utils.MustParseHex(proposerSlashing.SignedHeader2.Message.BodyRoot),
			},
		}
	}

	for i, attesterSlashing := range parsedBlock.Message.Body.AttesterSlashings {
		block.AttesterSlashings[i] = &types.AttesterSlashing{
			Attestation1: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation1.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation1.Data.Index),
					BeaconBlockRoot: utils.MustParseHex(attesterSlashing.Attestation1.Data.BeaconBlockRoot),
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation1.Data.Source.Epoch),
						Root:  utils.MustParseHex(attesterSlashing.Attestation1.Data.Source.Root),
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation1.Data.Target.Epoch),
						Root:  utils.MustParseHex(attesterSlashing.Attestation1.Data.Target.Root),
					},
				},
				Signature:        utils.MustParseHex(attesterSlashing.Attestation1.Signature),
				AttestingIndices: uint64List(attesterSlashing.Attestation1.AttestingIndices),
			},
			Attestation2: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation2.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation2.Data.Index),
					BeaconBlockRoot: utils.MustParseHex(attesterSlashing.Attestation2.Data.BeaconBlockRoot),
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation2.Data.Source.Epoch),
						Root:  utils.MustParseHex(attesterSlashing.Attestation2.Data.Source.Root),
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation2.Data.Target.Epoch),
						Root:  utils.MustParseHex(attesterSlashing.Attestation2.Data.Target.Root),
					},
				},
				Signature:        utils.MustParseHex(attesterSlashing.Attestation2.Signature),
				AttestingIndices: uint64List(attesterSlashing.Attestation2.AttestingIndices),
			},
		}
	}

	for i, attestation := range parsedBlock.Message.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: utils.MustParseHex(attestation.AggregationBits),
			Attesters:       []uint64{},
			Data: &types.AttestationData{
				Slot:            uint64(attestation.Data.Slot),
				CommitteeIndex:  uint64(attestation.Data.Index),
				BeaconBlockRoot: utils.MustParseHex(attestation.Data.BeaconBlockRoot),
				Source: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Source.Epoch),
					Root:  utils.MustParseHex(attestation.Data.Source.Root),
				},
				Target: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Target.Epoch),
					Root:  utils.MustParseHex(attestation.Data.Target.Root),
				},
			},
			Signature: utils.MustParseHex(attestation.Signature),
		}

		aggregationBits := bitfield.Bitlist(a.AggregationBits)
		assignments, err := lc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %w", a.Data.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch, err)
		}

		for i := uint64(0); i < aggregationBits.Len(); i++ {
			if aggregationBits.BitAt(i) {
				validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, a.Data.CommitteeIndex, i)]
				if !found { // This should never happen!
					validator = 0
					logger.Errorf("error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v", i, block.Slot, a.Data.Slot, a.Data.CommitteeIndex, i)
				}
				a.Attesters = append(a.Attesters, validator)

				if block.AttestationDuties[types.ValidatorIndex(validator)] == nil {
					block.AttestationDuties[types.ValidatorIndex(validator)] = []types.Slot{types.Slot(a.Data.Slot)}
				} else {
					block.AttestationDuties[types.ValidatorIndex(validator)] = append(block.AttestationDuties[types.ValidatorIndex(validator)], types.Slot(a.Data.Slot))
				}
			}
		}

		block.Attestations[i] = a
	}

	for i, deposit := range parsedBlock.Message.Body.Deposits {
		d := &types.Deposit{
			Proof:                 nil,
			PublicKey:             utils.MustParseHex(deposit.Data.Pubkey),
			WithdrawalCredentials: utils.MustParseHex(deposit.Data.WithdrawalCredentials),
			Amount:                uint64(deposit.Data.Amount),
			Signature:             utils.MustParseHex(deposit.Data.Signature),
		}

		block.Deposits[i] = d
	}

	for i, voluntaryExit := range parsedBlock.Message.Body.VoluntaryExits {
		block.VoluntaryExits[i] = &types.VoluntaryExit{
			Epoch:          uint64(voluntaryExit.Message.Epoch),
			ValidatorIndex: uint64(voluntaryExit.Message.ValidatorIndex),
			Signature:      utils.MustParseHex(voluntaryExit.Signature),
		}
	}

	for i, blsChange := range parsedBlock.Message.Body.SignedBLSToExecutionChange {
		block.SignedBLSToExecutionChange[i] = &types.SignedBLSToExecutionChange{
			Message: types.BLSToExecutionChange{
				Validatorindex: uint64(blsChange.Message.ValidatorIndex),
				BlsPubkey:      blsChange.Message.FromBlsPubkey,
				Address:        blsChange.Message.ToExecutionAddress,
			},
			Signature: blsChange.Signature,
		}
	}

	return block, nil
}

func syncCommitteeParticipation(bits []byte) float64 {
	participating := 0
	for i := 0; i < int(utils.Config.Chain.ClConfig.SyncCommitteeSize); i++ {
		if utils.BitAtVector(bits, i) {
			participating++
		}
	}
	return float64(participating) / float64(utils.Config.Chain.ClConfig.SyncCommitteeSize)
}

// GetValidatorParticipation will get the validator participation from the Lighthouse RPC api
func (lc *LighthouseClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {

	head, err := lc.GetChainHead()
	if err != nil {
		return nil, err
	}

	if epoch > head.HeadEpoch {
		return nil, fmt.Errorf("epoch %v is newer than the latest head %v", epoch, LighthouseLatestHeadEpoch)
	}
	if epoch == head.HeadEpoch {
		// participation stats are calculated at the end of an epoch,
		// making it impossible to retrieve stats of an currently ongoing epoch
		return nil, fmt.Errorf("epoch %v can't be retrieved as it hasn't finished yet", epoch)
	}

	request_epoch := epoch

	if epoch+1 < head.HeadEpoch {
		request_epoch += 1
	}

	logger.Infof("requesting validator inclusion data for epoch %v", request_epoch)

	resp, err := lc.get(fmt.Sprintf("%s/lighthouse/validator_inclusion/%d/global", lc.endpoint, request_epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator participation data for epoch %v: %w", request_epoch, err)
	}

	var parsedResponse LighthouseValidatorParticipationResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing validator participation data for epoch %v: %w", epoch, err)
	}

	var res *types.ValidatorParticipation
	if epoch < request_epoch {
		// we requested the next epoch, so we have to use the previous value for everything here
		res = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: float32(parsedResponse.Data.PreviousEpochTargetAttestingGwei) / float32(parsedResponse.Data.PreviousEpochActiveGwei),
			VotedEther:              uint64(parsedResponse.Data.PreviousEpochTargetAttestingGwei),
			EligibleEther:           uint64(parsedResponse.Data.PreviousEpochActiveGwei),
			Finalized:               epoch <= head.FinalizedEpoch && head.JustifiedEpoch > 0,
		}
	} else {
		res = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: float32(parsedResponse.Data.CurrentEpochTargetAttestingGwei) / float32(parsedResponse.Data.CurrentEpochActiveGwei),
			VotedEther:              uint64(parsedResponse.Data.CurrentEpochTargetAttestingGwei),
			EligibleEther:           uint64(parsedResponse.Data.CurrentEpochActiveGwei),
			Finalized:               epoch <= head.FinalizedEpoch && head.JustifiedEpoch > 0,
		}
	}
	return res, nil
}

func (lc *LighthouseClient) GetSyncCommittee(stateID string, epoch uint64) (*StandardSyncCommittee, error) {
	syncCommitteesResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/sync_committees?epoch=%d", lc.endpoint, stateID, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving sync_committees for epoch %v (state: %v): %w", epoch, stateID, err)
	}
	var parsedSyncCommittees StandardSyncCommitteesResponse
	err = json.Unmarshal(syncCommitteesResp, &parsedSyncCommittees)
	if err != nil {
		return nil, fmt.Errorf("error parsing sync_committees data for epoch %v (state: %v): %w", epoch, stateID, err)
	}
	return &parsedSyncCommittees.Data, nil
}

func (lc *LighthouseClient) GetBlobSidecars(stateID string) (*StandardBlobSidecarsResponse, error) {
	res, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/blob_sidecars/%s", lc.endpoint, stateID))
	if err != nil {
		return nil, fmt.Errorf("error retrieving blob_sidecars for %v: %w", stateID, err)
	}
	var parsed StandardBlobSidecarsResponse
	err = json.Unmarshal(res, &parsed)
	if err != nil {
		return nil, fmt.Errorf("error parsing blob_sidecars for %v: %w", stateID, err)
	}
	return &parsed, nil
}

var errNotFound = errors.New("not found 404")

func (lc *LighthouseClient) get(url string) ([]byte, error) {
	// t0 := time.Now()
	// defer func() { fmt.Println(url, time.Since(t0)) }()
	client := &http.Client{Timeout: time.Minute * 2}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("url: %v, error-response: %s", url, data)
	}

	return data, err
}

type bytesHexStr []byte

func (s *bytesHexStr) UnmarshalText(b []byte) error {
	if s == nil {
		return fmt.Errorf("cannot unmarshal bytes into nil")
	}
	if len(b) >= 2 && b[0] == '0' && b[1] == 'x' {
		b = b[2:]
	}
	out := make([]byte, len(b)/2)
	hex.Decode(out, b)
	*s = out
	return nil
}

type uint64Str uint64

func (s *uint64Str) UnmarshalJSON(b []byte) error {
	return Uint64Unmarshal((*uint64)(s), b)
}

// Parse a uint64, with or without quotes, in any base, with common prefixes accepted to change base.
func Uint64Unmarshal(v *uint64, b []byte) error {
	if v == nil {
		return errors.New("nil dest in uint64 decoding")
	}
	if len(b) == 0 {
		return errors.New("empty uint64 input")
	}
	if b[0] == '"' || b[0] == '\'' {
		if len(b) == 1 || b[len(b)-1] != b[0] {
			return errors.New("uneven/missing quotes")
		}
		b = b[1 : len(b)-1]
	}
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = n
	return nil
}

type StandardBeaconHeaderResponse struct {
	Data struct {
		Root   string `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64Str `json:"slot"`
				ProposerIndex uint64Str `json:"proposer_index"`
				ParentRoot    string    `json:"parent_root"`
				StateRoot     string    `json:"state_root"`
				BodyRoot      string    `json:"body_root"`
			} `json:"message"`
			Signature string `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}

type StandardBeaconHeadersResponse struct {
	Data []struct {
		Root   string `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64Str `json:"slot"`
				ProposerIndex uint64Str `json:"proposer_index"`
				ParentRoot    string    `json:"parent_root"`
				StateRoot     string    `json:"state_root"`
				BodyRoot      string    `json:"body_root"`
			} `json:"message"`
			Signature string `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}

type StandardFinalityCheckpointsResponse struct {
	Data struct {
		PreviousJustified struct {
			Epoch uint64Str `json:"epoch"`
			Root  string    `json:"root"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch uint64Str `json:"epoch"`
			Root  string    `json:"root"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch uint64Str `json:"epoch"`
			Root  string    `json:"root"`
		} `json:"finalized"`
	} `json:"data"`
}

type StreamedBlockEventData struct {
	Slot                uint64Str `json:"slot"`
	Block               string    `json:"block"`
	ExecutionOptimistic bool      `json:"execution_optimistic"`
}

type StandardProposerDuty struct {
	Pubkey         string    `json:"pubkey"`
	ValidatorIndex uint64Str `json:"validator_index"`
	Slot           uint64Str `json:"slot"`
}

type StandardProposerDutiesResponse struct {
	DependentRoot string                 `json:"dependent_root"`
	Data          []StandardProposerDuty `json:"data"`
}

type StandardCommitteeEntry struct {
	Index      uint64Str `json:"index"`
	Slot       uint64Str `json:"slot"`
	Validators []string  `json:"validators"`
}

type StandardCommitteesResponse struct {
	Data []StandardCommitteeEntry `json:"data"`
}

type StandardSyncCommittee struct {
	Validators          []string   `json:"validators"`
	ValidatorAggregates [][]string `json:"validator_aggregates"`
}

type StandardSyncCommitteesResponse struct {
	Data StandardSyncCommittee `json:"data"`
}

type LighthouseValidatorParticipationResponse struct {
	Data struct {
		CurrentEpochActiveGwei           uint64Str `json:"current_epoch_active_gwei"`
		PreviousEpochActiveGwei          uint64Str `json:"previous_epoch_active_gwei"`
		CurrentEpochTargetAttestingGwei  uint64Str `json:"current_epoch_target_attesting_gwei"`
		PreviousEpochTargetAttestingGwei uint64Str `json:"previous_epoch_target_attesting_gwei"`
		PreviousEpochHeadAttestingGwei   uint64Str `json:"previous_epoch_head_attesting_gwei"`
	} `json:"data"`
}

type ProposerSlashing struct {
	SignedHeader1 struct {
		Message struct {
			Slot          uint64Str `json:"slot"`
			ProposerIndex uint64Str `json:"proposer_index"`
			ParentRoot    string    `json:"parent_root"`
			StateRoot     string    `json:"state_root"`
			BodyRoot      string    `json:"body_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"signed_header_1"`
	SignedHeader2 struct {
		Message struct {
			Slot          uint64Str `json:"slot"`
			ProposerIndex uint64Str `json:"proposer_index"`
			ParentRoot    string    `json:"parent_root"`
			StateRoot     string    `json:"state_root"`
			BodyRoot      string    `json:"body_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"signed_header_2"`
}

type AttesterSlashing struct {
	Attestation1 struct {
		AttestingIndices []uint64Str `json:"attesting_indices"`
		Signature        string      `json:"signature"`
		Data             struct {
			Slot            uint64Str `json:"slot"`
			Index           uint64Str `json:"index"`
			BeaconBlockRoot string    `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64Str `json:"epoch"`
				Root  string    `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64Str `json:"epoch"`
				Root  string    `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_1"`
	Attestation2 struct {
		AttestingIndices []uint64Str `json:"attesting_indices"`
		Signature        string      `json:"signature"`
		Data             struct {
			Slot            uint64Str `json:"slot"`
			Index           uint64Str `json:"index"`
			BeaconBlockRoot string    `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64Str `json:"epoch"`
				Root  string    `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64Str `json:"epoch"`
				Root  string    `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_2"`
}

type Attestation struct {
	AggregationBits string `json:"aggregation_bits"`
	Signature       string `json:"signature"`
	Data            struct {
		Slot            uint64Str `json:"slot"`
		Index           uint64Str `json:"index"`
		BeaconBlockRoot string    `json:"beacon_block_root"`
		Source          struct {
			Epoch uint64Str `json:"epoch"`
			Root  string    `json:"root"`
		} `json:"source"`
		Target struct {
			Epoch uint64Str `json:"epoch"`
			Root  string    `json:"root"`
		} `json:"target"`
	} `json:"data"`
}

type Deposit struct {
	Proof []string `json:"proof"`
	Data  struct {
		Pubkey                string    `json:"pubkey"`
		WithdrawalCredentials string    `json:"withdrawal_credentials"`
		Amount                uint64Str `json:"amount"`
		Signature             string    `json:"signature"`
	} `json:"data"`
}

type VoluntaryExit struct {
	Message struct {
		Epoch          uint64Str `json:"epoch"`
		ValidatorIndex uint64Str `json:"validator_index"`
	} `json:"message"`
	Signature string `json:"signature"`
}

type Eth1Data struct {
	DepositRoot  string    `json:"deposit_root"`
	DepositCount uint64Str `json:"deposit_count"`
	BlockHash    string    `json:"block_hash"`
}

type SyncAggregate struct {
	SyncCommitteeBits      string `json:"sync_committee_bits"`
	SyncCommitteeSignature string `json:"sync_committee_signature"`
}

// https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
// https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
type ExecutionPayload struct {
	ParentHash    bytesHexStr   `json:"parent_hash"`
	FeeRecipient  bytesHexStr   `json:"fee_recipient"`
	StateRoot     bytesHexStr   `json:"state_root"`
	ReceiptsRoot  bytesHexStr   `json:"receipts_root"`
	LogsBloom     bytesHexStr   `json:"logs_bloom"`
	PrevRandao    bytesHexStr   `json:"prev_randao"`
	BlockNumber   uint64Str     `json:"block_number"`
	GasLimit      uint64Str     `json:"gas_limit"`
	GasUsed       uint64Str     `json:"gas_used"`
	Timestamp     uint64Str     `json:"timestamp"`
	ExtraData     bytesHexStr   `json:"extra_data"`
	BaseFeePerGas uint64Str     `json:"base_fee_per_gas"`
	BlockHash     bytesHexStr   `json:"block_hash"`
	Transactions  []bytesHexStr `json:"transactions"`
	// present only after capella
	Withdrawals []WithdrawalPayload `json:"withdrawals"`
	// present only after deneb
	BlobGasUsed   uint64Str `json:"blob_gas_used"`
	ExcessBlobGas uint64Str `json:"excess_blob_gas"`
}

type WithdrawalPayload struct {
	Index          uint64Str   `json:"index"`
	ValidatorIndex uint64Str   `json:"validator_index"`
	Address        bytesHexStr `json:"address"`
	Amount         uint64Str   `json:"amount"`
}

type SignedBLSToExecutionChange struct {
	Message struct {
		ValidatorIndex     uint64Str   `json:"validator_index"`
		FromBlsPubkey      bytesHexStr `json:"from_bls_pubkey"`
		ToExecutionAddress bytesHexStr `json:"to_execution_address"`
	} `json:"message"`
	Signature bytesHexStr `json:"signature"`
}

type AnySignedBlock struct {
	Message struct {
		Slot          uint64Str `json:"slot"`
		ProposerIndex uint64Str `json:"proposer_index"`
		ParentRoot    string    `json:"parent_root"`
		StateRoot     string    `json:"state_root"`
		Body          struct {
			RandaoReveal      string             `json:"randao_reveal"`
			Eth1Data          Eth1Data           `json:"eth1_data"`
			Graffiti          string             `json:"graffiti"`
			ProposerSlashings []ProposerSlashing `json:"proposer_slashings"`
			AttesterSlashings []AttesterSlashing `json:"attester_slashings"`
			Attestations      []Attestation      `json:"attestations"`
			Deposits          []Deposit          `json:"deposits"`
			VoluntaryExits    []VoluntaryExit    `json:"voluntary_exits"`

			// not present in phase0 blocks
			SyncAggregate *SyncAggregate `json:"sync_aggregate,omitempty"`

			// not present in phase0/altair blocks
			ExecutionPayload *ExecutionPayload `json:"execution_payload"`

			// present only after capella
			SignedBLSToExecutionChange []*SignedBLSToExecutionChange `json:"bls_to_execution_changes"`

			// present only after deneb
			BlobKZGCommitments []bytesHexStr `json:"blob_kzg_commitments"`
		} `json:"body"`
	} `json:"message"`
	Signature bytesHexStr `json:"signature"`
}

type StandardV2BlockResponse struct {
	Version             string         `json:"version"`
	ExecutionOptimistic bool           `json:"execution_optimistic"`
	Finalized           bool           `json:"finalized"`
	Data                AnySignedBlock `json:"data"`
}

type StandardV1BlockRootResponse struct {
	Data struct {
		Root string `json:"root"`
	} `json:"data"`
}

type StandardValidatorEntry struct {
	Index     uint64Str `json:"index"`
	Balance   uint64Str `json:"balance"`
	Status    string    `json:"status"`
	Validator struct {
		Pubkey                     string    `json:"pubkey"`
		WithdrawalCredentials      string    `json:"withdrawal_credentials"`
		EffectiveBalance           uint64Str `json:"effective_balance"`
		Slashed                    bool      `json:"slashed"`
		ActivationEligibilityEpoch uint64Str `json:"activation_eligibility_epoch"`
		ActivationEpoch            uint64Str `json:"activation_epoch"`
		ExitEpoch                  uint64Str `json:"exit_epoch"`
		WithdrawableEpoch          uint64Str `json:"withdrawable_epoch"`
	} `json:"validator"`
}

type StandardValidatorsResponse struct {
	Data []StandardValidatorEntry `json:"data"`
}

type StandardSyncingResponse struct {
	Data struct {
		IsSyncing    bool      `json:"is_syncing"`
		HeadSlot     uint64Str `json:"head_slot"`
		SyncDistance uint64Str `json:"sync_distance"`
	} `json:"data"`
}

type StandardValidatorBalancesResponse struct {
	Data []struct {
		Index   uint64Str `json:"index"`
		Balance uint64Str `json:"balance"`
	} `json:"data"`
}

type StandardBlobSidecarsResponse struct {
	Data []struct {
		BlockRoot       bytesHexStr `json:"block_root"`
		Index           uint64Str   `json:"index"`
		Slot            uint64Str   `json:"slot"`
		BlockParentRoot bytesHexStr `json:"block_parent_root"`
		ProposerIndex   uint64Str   `json:"proposer_index"`
		KzgCommitment   bytesHexStr `json:"kzg_commitment"`
		KzgProof        bytesHexStr `json:"kzg_proof"`
		// Blob            string `json:"blob"`
	}
}
