package rpc

import (
	"encoding/json"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/protolambda/zssz/bitfields"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prysmaticlabs/go-bitfield"
)

// LighthouseClient holds the Lighthouse client info
type LighthouseClient struct {
	endpoint            string
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
}

// NewLighthouseClient is used to create a new Lighthouse client
func NewLighthouseClient(endpoint string) (*LighthouseClient, error) {
	client := &LighthouseClient{
		endpoint:            endpoint,
		assignmentsCacheMux: &sync.Mutex{},
	}
	client.assignmentsCache, _ = lru.New(10)

	return client, nil
}

func (lc *LighthouseClient) GetNewBlockChan() chan *types.Block {
	blkCh := make(chan *types.Block, 10)
	go func() {
		// poll sync status 2 times per slot
		t := time.NewTicker(time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot) / 2)

		lastHeadSlot := uint64(0)
		headResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/head", lc.endpoint))
		if err == nil {
			var parsedHead StandardBeaconHeaderResponse
			err = json.Unmarshal(headResp, &parsedHead)
			if err != nil {
				logger.Warnf("failed to decode head, starting blocks channel at slot 0")
			} else {
				lastHeadSlot = uint64(parsedHead.Data.Header.Message.Slot)
			}
		} else {
			logger.Warnf("failed to fetch head, starting blocks channel at slot 0")
		}

		for {
			<-t.C
			syncingResp, err := lc.get(fmt.Sprintf("%s/eth/v1/node/syncing", lc.endpoint))
			if err != nil {
				logger.Warnf("failed to retrieve syncing status: %v", err)
				continue
			}
			var parsedSyncing StandardSyncingResponse
			err = json.Unmarshal(syncingResp, &parsedSyncing)
			if err != nil {
				logger.Warnf("failed to decode syncing status: %v", err)
				continue
			}
			for slot := lastHeadSlot; slot < uint64(parsedSyncing.Data.HeadSlot); slot++ {
				blks, err := lc.GetBlocksBySlot(slot)
				if err != nil {
					logger.Warnf("failed to fetch block(s) for slot %d: %v", slot, err)
					continue
				}
				for _, blk := range blks {
					blkCh <- blk
				}
				lastHeadSlot = slot
			}
		}
	}()
	return blkCh
}

// GetChainHead gets the chain head from Lighthouse
func (lc *LighthouseClient) GetChainHead() (*types.ChainHead, error) {
	headResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/head", lc.endpoint))
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain head: %v", err)
	}

	var parsedHead StandardBeaconHeaderResponse
	err = json.Unmarshal(headResp, &parsedHead)
	if err != nil {
		return nil, fmt.Errorf("error parsing chain head: %v", err)
	}

	id := parsedHead.Data.Header.Message.StateRoot
	if parsedHead.Data.Header.Message.Slot == 0 {
		id = "genesis"
	}
	finalityResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", lc.endpoint, id))
	if err != nil {
		return nil, fmt.Errorf("error retrieving finality checkpoints of head: %v", err)
	}

	var parsedFinality StandardFinalityCheckpointsResponse
	err = json.Unmarshal(finalityResp, &parsedFinality)
	if err != nil {
		return nil, fmt.Errorf("error parsing finality checkpoints of head: %v", err)
	}

	return &types.ChainHead{
		HeadSlot:                   uint64(parsedHead.Data.Header.Message.Slot),
		HeadEpoch:                  uint64(parsedHead.Data.Header.Message.Slot) / utils.Config.Chain.SlotsPerEpoch,
		HeadBlockRoot:              utils.MustParseHex(parsedHead.Data.Root),
		FinalizedSlot:              uint64(parsedFinality.Data.Finalized.Epoch) * utils.Config.Chain.SlotsPerEpoch,
		FinalizedEpoch:             uint64(parsedFinality.Data.Finalized.Epoch),
		FinalizedBlockRoot:         utils.MustParseHex(parsedFinality.Data.Finalized.Root),
		JustifiedSlot:              uint64(parsedFinality.Data.CurrentJustified.Epoch) * utils.Config.Chain.SlotsPerEpoch,
		JustifiedEpoch:             uint64(parsedFinality.Data.CurrentJustified.Epoch),
		JustifiedBlockRoot:         utils.MustParseHex(parsedFinality.Data.CurrentJustified.Root),
		PreviousJustifiedSlot:      uint64(parsedFinality.Data.PreviousJustified.Epoch) * utils.Config.Chain.SlotsPerEpoch,
		PreviousJustifiedEpoch:     uint64(parsedFinality.Data.PreviousJustified.Epoch),
		PreviousJustifiedBlockRoot: utils.MustParseHex(parsedFinality.Data.PreviousJustified.Root),
	}, nil
}

func (lc *LighthouseClient) GetValidatorQueue() (*types.ValidatorQueue, error) {
	// pre-filter the status, to return much less validators, thus much faster!
	validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/head/validators?status=pending_queued,exited", lc.endpoint))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator for head valiqdator queue check: %v", err)
	}

	var parsedValidators StandardValidatorsResponse
	err = json.Unmarshal(validatorsResp, &parsedValidators)
	if err != nil {
		return nil, fmt.Errorf("error parsing queue validators: %v", err)
	}
	// TODO: maybe track more status counts in the future?
	activatingValidatorCount := uint64(0)
	exitingValidatorCount := uint64(0)
	for _, validator := range parsedValidators.Data {
		switch validator.Status {
		case "pending_initialized":
			break
		case "pending_queued":
			activatingValidatorCount += 1
			break
		case "active_ongoing", "active_exiting", "active_slashed":
			break
		case "exited_unslashed", "exited_slashed":
			exitingValidatorCount += exitingValidatorCount
			break
		case "withdrawal_possible", "withdrawal_done":
			break
		default:
			return nil, fmt.Errorf("unrecognized validator status (validator %d): %s", validator.Index, validator.Status)
		}
	}
	return &types.ValidatorQueue{
		Activating: activatingValidatorCount,
		Exititing:  exitingValidatorCount,
	}, nil
}

// GetEpochAssignments will get the epoch assignments from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	lc.assignmentsCacheMux.Lock()
	defer lc.assignmentsCacheMux.Unlock()

	var err error

	cachedValue, found := lc.assignmentsCache.Get(epoch)
	if found {
		return cachedValue.(*types.EpochAssignments), nil
	}

	proposerResp, err := lc.get(fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", lc.endpoint, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving proposer duties: %v", err)
	}
	var parsedProposerResponse StandardProposerDutiesResponse
	err = json.Unmarshal(proposerResp, &parsedProposerResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing proposer duties: %v", err)
	}

	// fetch the block root that the proposer data is dependent on
	headerResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/headers/%s", lc.endpoint, parsedProposerResponse.DependentRoot))
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain header: %v", err)
	}
	var parsedHeader StandardBeaconHeaderResponse
	err = json.Unmarshal(headerResp, &parsedHeader)
	if err != nil {
		return nil, fmt.Errorf("error parsing chain header: %v", err)
	}
	depStateRoot := parsedHeader.Data.Header.Message.StateRoot

	// Now use the state root to make a consistent committee query
	committeesResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/committees?epoch=%d", lc.endpoint, depStateRoot, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain header: %v", err)
	}
	var parsedCommittees StandardCommitteesResponse
	err = json.Unmarshal(committeesResp, &parsedCommittees)
	if err != nil {
		return nil, fmt.Errorf("error parsing committees data: %v", err)
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}
	for _, duty := range parsedProposerResponse.Data {
		assignments.ProposerAssignments[uint64(duty.Slot)] = uint64(duty.ValidatorIndex)
	}

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

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		lc.assignmentsCache.Add(epoch, assignments)
	}

	return assignments, nil
}

// GetEpochData will get the epoch data from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochData(epoch uint64) (*types.EpochData, error) {
	var err error

	data := &types.EpochData{}
	data.Epoch = epoch

	validatorsResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", lc.endpoint, epoch*utils.Config.Chain.SlotsPerEpoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validators for epoch %v: %v", epoch, err)
	}

	var parsedValidators StandardValidatorsResponse
	err = json.Unmarshal(validatorsResp, &parsedValidators)

	if err != nil {
		return nil, fmt.Errorf("error parsing epoch validators: %v", err)
	}

	epoch1d := int64(epoch) - 225
	if epoch1d < 0 {
		epoch1d = 0
	}
	start := time.Now()
	validatorBalances1d, err := lc.getBalancesForEpoch(epoch1d)
	if err != nil {
		return nil, err
	}
	logger.Printf("retrieved data for %v validator balances for epoch %v (1d) took %v", len(parsedValidators.Data), epoch1d, time.Since(start))
	epoch7d := int64(epoch) - 225*7
	if epoch7d < 0 {
		epoch7d = 0
	}
	start = time.Now()
	validatorBalances7d, err := lc.getBalancesForEpoch(epoch7d)
	if err != nil {
		return nil, err
	}
	logger.Printf("retrieved data for %v validator balances for epoch %v (7d) took %v", len(parsedValidators.Data), epoch7d, time.Since(start))
	start = time.Now()
	epoch31d := int64(epoch) - 225*31
	if epoch31d < 0 {
		epoch31d = 0
	}
	validatorBalances31d, err := lc.getBalancesForEpoch(epoch31d)
	if err != nil {
		return nil, err
	}
	logger.Printf("retrieved data for %v validator balances for epoch %v (31d) took %v", len(parsedValidators.Data), epoch31d, time.Since(start))

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
			Balance1d:                  validatorBalances1d[uint64(validator.Index)],
			Balance7d:                  validatorBalances7d[uint64(validator.Index)],
			Balance31d:                 validatorBalances31d[uint64(validator.Index)],
		})
	}

	logger.Printf("retrieved data for %v validators for epoch %v", len(data.Validators), epoch)

	data.ValidatorAssignmentes, err = lc.GetEpochAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving assignments for epoch %v: %v", epoch, err)
	}
	logger.Printf("retrieved validator assignment data for epoch %v", epoch)

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.Config.Chain.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.SlotsPerEpoch-1; slot++ {

		if slot == 0 || utils.SlotToTime(slot).After(time.Now()) { // Currently slot 0 returns all blocks, also skip asking for future blocks
			continue
		}

		blocks, err := lc.GetBlocksBySlot(slot)

		if err != nil {
			logger.Errorf("error retrieving blocks for slot %v: %v", slot, err)
			continue
		}

		for _, block := range blocks {
			if data.Blocks[block.Slot] == nil {
				data.Blocks[block.Slot] = make(map[string]*types.Block)
			}
			data.Blocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block
		}
	}
	logger.Printf("retrieved %v blocks for epoch %v", len(data.Blocks), epoch)

	// Fill up missed and scheduled blocks
	for slot, proposer := range data.ValidatorAssignmentes.ProposerAssignments {
		_, found := data.Blocks[slot]
		if !found {
			// Proposer was assigned but did not yet propose a block
			data.Blocks[slot] = make(map[string]*types.Block)
			data.Blocks[slot]["0x0"] = &types.Block{
				Status:            0,
				Canonical:         true,
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

			if utils.SlotToTime(slot).After(time.Now().Add(time.Second * -60)) {
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

	data.EpochParticipationStats, err = lc.GetValidatorParticipation(epoch)
	if err != nil {
		logger.Errorf("error retrieving epoch participation statistics for epoch %v: %v", epoch, err)

		data.EpochParticipationStats = &types.ValidatorParticipation{
			Epoch:                   epoch,
			Finalized:               false,
			GlobalParticipationRate: 0,
			VotedEther:              0,
			EligibleEther:           0,
		}
	}

	return data, nil
}

func bitlistParticipation(bits []byte) float64 {
	bitLen := bitfields.BitlistLen(bits)
	participating := uint64(0)
	for i := uint64(0); i < bitLen; i++ {
		if bitfields.GetBit(bits, i) {
			participating += 1
		}
	}
	return float64(participating) / float64(bitLen)
}

func uint64List(li []uint64Str) []uint64 {
	out := make([]uint64, len(li), len(li))
	for i, v := range li {
		out[i] = uint64(v)
	}
	return out
}

func (lc *LighthouseClient) getBalancesForEpoch(epoch int64) (map[uint64]uint64, error) {

	if epoch < 0 {
		epoch = 0
	}

	var err error

	validatorBalances := make(map[uint64]uint64)

	resp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%d/validator_balances", lc.endpoint, epoch*int64(utils.Config.Chain.SlotsPerEpoch)))
	if err != nil {
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

// GetBlocksBySlot will get the blocks by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {
	respRoot, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/blocks/%d/root", lc.endpoint, slot))
	if err != nil {
		if err == notFoundErr {
			// no block in this slot
			return []*types.Block{}, nil
		}
		return nil, fmt.Errorf("error retrieving block root at slot %v: %v", slot, err)
	}
	var parsedRoot StandardV1BlockRootResponse
	err = json.Unmarshal(respRoot, &parsedRoot)
	if err != nil {
		logger.Errorf("error parsing block root at slot %v: %v", slot, err)
		return []*types.Block{}, nil
	}

	resp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/blocks/%s", lc.endpoint, parsedRoot.Data.Root))
	if err != nil {
		return nil, fmt.Errorf("error retrieving block data at slot %v: %v", slot, err)
	}

	var parsedResponse StandardV2BlockResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		logger.Errorf("error parsing block data at slot %v: %v", slot, err)
		return []*types.Block{}, nil
	}

	parsedBlock := parsedResponse.Data

	block := &types.Block{
		Status:       1,
		Canonical:    true,
		Proposer:     uint64(parsedBlock.Message.ProposerIndex),
		BlockRoot:    utils.MustParseHex(parsedRoot.Data.Root),
		Slot:         slot,
		ParentRoot:   utils.MustParseHex(parsedBlock.Message.ParentRoot),
		StateRoot:    utils.MustParseHex(parsedBlock.Message.StateRoot),
		Signature:    utils.MustParseHex(parsedBlock.Signature),
		RandaoReveal: utils.MustParseHex(parsedBlock.Message.Body.RandaoReveal),
		Graffiti:     utils.MustParseHex(parsedBlock.Message.Body.Graffiti),
		Eth1Data: &types.Eth1Data{
			DepositRoot:  utils.MustParseHex(parsedBlock.Message.Body.Eth1Data.DepositRoot),
			DepositCount: uint64(parsedBlock.Message.Body.Eth1Data.DepositCount),
			BlockHash:    utils.MustParseHex(parsedBlock.Message.Body.Eth1Data.BlockHash),
		},
		ProposerSlashings: make([]*types.ProposerSlashing, len(parsedBlock.Message.Body.ProposerSlashings)),
		AttesterSlashings: make([]*types.AttesterSlashing, len(parsedBlock.Message.Body.AttesterSlashings)),
		Attestations:      make([]*types.Attestation, len(parsedBlock.Message.Body.Attestations)),
		Deposits:          make([]*types.Deposit, len(parsedBlock.Message.Body.Deposits)),
		VoluntaryExits:    make([]*types.VoluntaryExit, len(parsedBlock.Message.Body.VoluntaryExits)),
	}

	if agg := parsedBlock.Message.Body.SyncAggregate; agg != nil {
		bits := utils.MustParseHex(agg.SyncCommitteeBits)
		block.SyncAggregate = &types.SyncAggregate{
			SyncCommitteeBits:          bits,
			SyncAggregateParticipation: bitlistParticipation(bits),
			SyncCommitteeSignature:     utils.MustParseHex(agg.SyncCommitteeSignature),
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
		assignments, err := lc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.SlotsPerEpoch, err)
		}

		for i := uint64(0); i < aggregationBits.Len(); i++ {
			if aggregationBits.BitAt(i) {
				validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, a.Data.CommitteeIndex, i)]
				if !found { // This should never happen!
					validator = 0
					logger.Errorf("error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v", i, block.Slot, a.Data.Slot, a.Data.CommitteeIndex, i)
				}
				a.Attesters = append(a.Attesters, validator)
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

	return []*types.Block{block}, nil
}

// GetValidatorParticipation will get the validator participation from the Lighthouse RPC api
func (lc *LighthouseClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	resp, err := lc.get(fmt.Sprintf("%s/lighthouse/validator_inclusion/%d/global", lc.endpoint, epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator participation data for epoch %v: %v", epoch, err)
	}

	var parsedResponse LighthouseValidatorParticipationResponse
	err = json.Unmarshal(resp, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing validator participation data for epoch %v: %v", epoch, err)
	}

	return &types.ValidatorParticipation{
		Epoch: epoch,
		// technically there are rules for delayed finalization through previous epoch, applied only later. But good enough, this matches previous wonky behavior
		Finalized:               float32(parsedResponse.Data.CurrentEpochTargetAttestingGwei)/float32(parsedResponse.Data.CurrentEpochActiveGwei) > (float32(2) / float32(3)),
		GlobalParticipationRate: float32(parsedResponse.Data.PreviousEpochTargetAttestingGwei) / float32(parsedResponse.Data.CurrentEpochActiveGwei),
		VotedEther:              uint64(parsedResponse.Data.PreviousEpochTargetAttestingGwei),
		EligibleEther:           uint64(parsedResponse.Data.CurrentEpochActiveGwei),
	}, nil
}

func (lc *LighthouseClient) GetFinalityCheckpoints(epoch uint64) (*types.FinalityCheckpoints, error) {
	// finalityResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", lc.endpoint, id))
	// if err != nil {
	// 	return nil, fmt.Errorf("error retrieving finality checkpoints of head: %v", err)
	// }
	return &types.FinalityCheckpoints{}, nil
}

var notFoundErr = errors.New("not found 404")

func (lc *LighthouseClient) get(url string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * 120}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, notFoundErr
		}
		return nil, fmt.Errorf("error-response: %s", data)
	}

	return data, err
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
		Root      string `json:"root"`
		Canonical bool   `json:"canonical"`
		Header    struct {
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
		} `json:"body"`
	} `json:"message"`
	Signature string `json:"signature"`
}

type StandardV2BlockResponse struct {
	Version string         `json:"version"`
	Data    AnySignedBlock `json:"data"`
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

func (pc *LighthouseClient) GetBlockStatusByEpoch(epoch uint64) ([]*types.CanonBlock, error) {
	blocks := make([]*types.CanonBlock, 0)
	return blocks, nil
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
