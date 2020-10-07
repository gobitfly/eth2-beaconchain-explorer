package rpc

import (
	"errors"
	"eth2-exporter/eth2api"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
)

type Eth2ApiV1Client struct {
	client                      *eth2api.Client
	proposerAssignmentsCache    *lru.Cache
	proposerAssignmentsCacheMux *sync.Mutex
	attesterAssignmentsCache    *lru.Cache
	attesterAssignmentsCacheMux *sync.Mutex
	validatorsCache             *lru.Cache
	validatorsCacheMux          *sync.Mutex
}

func NewEth2ApiV1Client(endpoint string) (*Eth2ApiV1Client, error) {
	c, err := eth2api.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	client := &Eth2ApiV1Client{
		client:                      c,
		proposerAssignmentsCacheMux: &sync.Mutex{},
		attesterAssignmentsCacheMux: &sync.Mutex{},
		validatorsCacheMux:          &sync.Mutex{},
	}

	client.proposerAssignmentsCache, _ = lru.New(10)
	client.attesterAssignmentsCache, _ = lru.New(10)
	client.validatorsCache, _ = lru.New(10)

	return client, nil
}

func (c *Eth2ApiV1Client) GetChainHead() (*types.ChainHead, error) {
	headers, err := c.client.GetHeaders()
	if err != nil {
		return nil, err
	}
	var headSlot uint64
	var headBlockRoot []byte
	var hasCanonical bool
	for _, h := range headers {
		if !h.Canonical {
			continue
		}
		hasCanonical = true
		headSlot = h.Header.Message.Slot
		headBlockRoot = h.Root
	}
	if !hasCanonical {
		return nil, fmt.Errorf("error getting chainhead: no canonical header found")
	}

	finalityCheckpoints, err := c.client.GetFinalityCheckpoints(fmt.Sprintf("%d", headSlot))
	if err != nil {
		return nil, err
	}

	// note: {Finalized,Justified,PriviousJustified}Slot is missing, it is currently not used anywhere anyway
	return &types.ChainHead{
		HeadEpoch:                  headSlot / utils.Config.Chain.SlotsPerEpoch,
		HeadBlockRoot:              headBlockRoot,
		FinalizedEpoch:             finalityCheckpoints.Finalized.Epoch,
		FinalizedBlockRoot:         finalityCheckpoints.Finalized.Root,
		JustifiedEpoch:             finalityCheckpoints.CurrentJustified.Epoch,
		JustifiedBlockRoot:         finalityCheckpoints.CurrentJustified.Root,
		PreviousJustifiedEpoch:     finalityCheckpoints.PreviousJustified.Epoch,
		PreviousJustifiedBlockRoot: finalityCheckpoints.PreviousJustified.Root,
	}, nil
}

func (c *Eth2ApiV1Client) GetEpochData(epoch uint64) (*types.EpochData, error) {
	var err error

	t0 := time.Now()
	data := &types.EpochData{}
	data.Epoch = epoch

	slotStr := fmt.Sprintf("%d", epoch*utils.Config.Chain.SlotsPerEpoch)

	data.ValidatorAssignmentes, err = c.GetEpochAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving assignments for epoch %v: %w", epoch, err)
	}

	t1 := time.Now()

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.Config.Chain.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.SlotsPerEpoch-1; slot++ {
		blocks, err := c.GetBlocksBySlot(slot)
		if err != nil {
			return nil, err
		}
		for _, block := range blocks {
			if data.Blocks[block.Slot] == nil {
				data.Blocks[block.Slot] = make(map[string]*types.Block)
			}
			data.Blocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block
		}
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

	t2 := time.Now()

	// Retrieve the validator set for the epoch

	validators, err := c.client.GetValidators(slotStr)
	if err != nil {
		return nil, err
	}

	data.Validators = make([]*types.Validator, len(validators))

	for i, validator := range validators {
		data.Validators[i] = &types.Validator{
			Index:                      validator.Index,
			PublicKey:                  validator.Validator.Pubkey,
			WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
			Balance:                    validator.Balance,
			EffectiveBalance:           validator.Validator.EffectiveBalance,
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.Validator.ActivationEpoch,
			ExitEpoch:                  validator.Validator.ExitEpoch,
			WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
		}
	}

	t3 := time.Now()

	logger.WithFields(logrus.Fields{
		"validators":     len(data.Validators),
		"blocks":         len(data.Blocks),
		"dur":            time.Since(t0),
		"durValidators":  t3.Sub(t2),
		"durBlocks":      t2.Sub(t1),
		"durAssignments": t1.Sub(t0),
	}).Info("GetEpochData")

	return data, nil
}

func (c *Eth2ApiV1Client) GetValidators(epoch uint64) ([]*eth2api.Validator, error) {
	c.validatorsCacheMux.Lock()
	defer c.validatorsCacheMux.Unlock()

	cachedValue, found := c.validatorsCache.Get(epoch)
	if found {
		return cachedValue.([]*eth2api.Validator), nil
	}

	validators, err := c.client.GetValidators(fmt.Sprintf("%d", epoch))
	if err != nil {
		return nil, err
	}

	logger.Infof("got %v validators for epoch %v", len(validators), epoch)

	if len(validators) > 0 {
		c.validatorsCache.Add(epoch, validators)
	}

	return validators, nil
}

func (c *Eth2ApiV1Client) GetValidatorQueue() (*types.ValidatorQueue, error) {
	return nil, nil
}

// GetAttestationPool is not needed
func (c *Eth2ApiV1Client) GetAttestationPool() ([]*types.Attestation, error) {
	return []*types.Attestation{}, nil
}

func (c *Eth2ApiV1Client) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	c.proposerAssignmentsCacheMux.Lock()
	defer c.proposerAssignmentsCacheMux.Unlock()

	c.attesterAssignmentsCacheMux.Lock()
	defer c.attesterAssignmentsCacheMux.Unlock()

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	cachedProposerAssignments, found := c.proposerAssignmentsCache.Get(epoch)
	if found {
		assignments.ProposerAssignments = cachedProposerAssignments.(map[uint64]uint64)
	} else {
		start := time.Now()
		proposerDuties, err := c.client.GetProposerDuties(epoch)
		if err != nil {
			var apiErr *eth2api.APIError
			if errors.As(err, &apiErr) && err.(*eth2api.APIError).Code == 400 {
				// logger.Info(err.(*eth2api.APIError).Message)
				proposerDuties = []*eth2api.ProposerDuty{}
			} else {
				return nil, err
			}
		}
		for _, proposerDuty := range proposerDuties {
			assignments.ProposerAssignments[proposerDuty.Slot] = proposerDuty.ValidatorIndex
		}
		if len(assignments.ProposerAssignments) > 0 {
			c.proposerAssignmentsCache.Add(epoch, assignments.ProposerAssignments)
		}
		logger.WithFields(logrus.Fields{
			"dur":                 time.Since(start),
			"epoch":               epoch,
			"proposerAssignments": len(assignments.ProposerAssignments),
		}).Info("cached proposerAssignments")
	}

	cachedAttesterAssignments, found := c.attesterAssignmentsCache.Get(epoch)
	if found {
		assignments.AttestorAssignments = cachedAttesterAssignments.(map[string]uint64)
	} else {
		start := time.Now()
		slotStr := fmt.Sprintf("%d", epoch*utils.Config.Chain.SlotsPerEpoch)
		committees, err := c.client.GetCommittees(slotStr, epoch)
		if err != nil {
			return nil, err
		}
		for _, committee := range committees {
			for i, v := range committee.Validators {
				assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(uint64(committee.Slot), uint64(committee.Index), uint64(i))] = uint64(v)
			}
		}
		if len(assignments.AttestorAssignments) > 0 {
			c.attesterAssignmentsCache.Add(epoch, assignments.AttestorAssignments)
		}
		logger.WithFields(logrus.Fields{
			"dur":                 time.Since(start),
			"epoch":               epoch,
			"attesterAssignments": len(assignments.AttestorAssignments),
		}).Info("cached attesterAssignments")
	}

	return assignments, nil
}

func (c *Eth2ApiV1Client) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {
	t0 := time.Now()
	slotStr := fmt.Sprintf("%v", slot)

	b1, err := c.client.GetBlock(slotStr)
	if err != nil {
		return nil, err
	}
	t1 := time.Now()

	if b1 == nil || slot != b1.Message.Slot {
		return []*types.Block{}, nil
	}

	b1Root, err := c.client.GetBlockRoot(slotStr)
	if err != nil {
		return nil, err
	}
	t2 := time.Now()

	b1Header, err := c.client.GetHeader(slotStr)
	if err != nil {
		return nil, err
	}
	t3 := time.Now()

	b2 := types.Block{
		Status:            1,
		Proposer:          b1.Message.ProposerIndex,
		BlockRoot:         b1Root.Root,
		Slot:              b1.Message.Slot,
		ParentRoot:        b1.Message.ParentRoot,
		StateRoot:         b1.Message.StateRoot,
		Signature:         b1.Signature,
		RandaoReveal:      b1.Message.Body.RandaoReveal,
		Graffiti:          b1.Message.Body.Graffiti,
		Eth1Data:          &types.Eth1Data{},
		BodyRoot:          b1Header.Header.Message.BodyRoot,
		ProposerSlashings: make([]*types.ProposerSlashing, len(b1.Message.Body.ProposerSlashings)),
		AttesterSlashings: make([]*types.AttesterSlashing, len(b1.Message.Body.AttesterSlashings)),
		Attestations:      make([]*types.Attestation, len(b1.Message.Body.Attestations)),
		Deposits:          make([]*types.Deposit, len(b1.Message.Body.Deposits)),
		VoluntaryExits:    make([]*types.VoluntaryExit, len(b1.Message.Body.VoluntaryExits)),
	}

	for i, v := range b1.Message.Body.ProposerSlashings {
		b2.ProposerSlashings[i] = &types.ProposerSlashing{
			ProposerIndex: v.SignedHeader1.Message.ProposerIndex,
			Header1: &types.Block{
				Slot:       v.SignedHeader1.Message.Slot,
				ParentRoot: v.SignedHeader1.Message.ParentRoot,
				StateRoot:  v.SignedHeader1.Message.StateRoot,
				BodyRoot:   v.SignedHeader1.Message.BodyRoot,
				Signature:  v.SignedHeader1.Signature,
			},
			Header2: &types.Block{
				Slot:       v.SignedHeader2.Message.Slot,
				ParentRoot: v.SignedHeader2.Message.ParentRoot,
				StateRoot:  v.SignedHeader2.Message.StateRoot,
				BodyRoot:   v.SignedHeader2.Message.BodyRoot,
				Signature:  v.SignedHeader2.Signature,
			},
		}
	}

	for i, v := range b1.Message.Body.AttesterSlashings {
		b2.AttesterSlashings[i] = &types.AttesterSlashing{
			Attestation1: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            v.Attestation1.Data.Slot,
					CommitteeIndex:  v.Attestation1.Data.Index,
					BeaconBlockRoot: v.Attestation1.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: v.Attestation1.Data.Source.Epoch,
						Root:  v.Attestation1.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: v.Attestation1.Data.Target.Epoch,
						Root:  v.Attestation1.Data.Target.Root,
					},
				},
				Signature:        v.Attestation1.Signature,
				AttestingIndices: v.Attestation1.AttestingIndices,
			},
			Attestation2: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            v.Attestation2.Data.Slot,
					CommitteeIndex:  v.Attestation2.Data.Index,
					BeaconBlockRoot: v.Attestation2.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: v.Attestation2.Data.Source.Epoch,
						Root:  v.Attestation2.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: v.Attestation2.Data.Target.Epoch,
						Root:  v.Attestation2.Data.Target.Root,
					},
				},
				Signature:        v.Attestation2.Signature,
				AttestingIndices: v.Attestation2.AttestingIndices,
			},
		}
	}

	for i, attestation := range b1.Message.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: attestation.AggregationBits,
			Data: &types.AttestationData{
				Slot:            attestation.Data.Slot,
				CommitteeIndex:  attestation.Data.Index,
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
			Signature: attestation.Signature,
		}

		aggregationBits := bitfield.Bitlist(a.AggregationBits)
		assignments, err := c.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.SlotsPerEpoch, err)
		}

		a.Attesters = make([]uint64, 0)
		for i := uint64(0); i < aggregationBits.Len(); i++ {
			if aggregationBits.BitAt(i) {
				validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, a.Data.CommitteeIndex, i)]
				if !found { // This should never happen!
					validator = 0
					logger.Errorf(
						"error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v",
						i, b2.Slot, a.Data.Slot, a.Data.CommitteeIndex, i,
					)
				}
				a.Attesters = append(a.Attesters, validator)
			}
		}

		b2.Attestations[i] = a
	}

	for i, deposit := range b1.Message.Body.Deposits {
		b2.Deposits[i] = &types.Deposit{
			Proof:                 deposit.Proof,
			PublicKey:             deposit.Data.Pubkey,
			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
			Amount:                deposit.Data.Amount,
			Signature:             deposit.Data.Signature,
		}
	}

	for i, voluntaryExit := range b1.Message.Body.VoluntaryExits {
		b2.VoluntaryExits[i] = &types.VoluntaryExit{
			Epoch:          voluntaryExit.Message.Epoch,
			ValidatorIndex: voluntaryExit.Message.ValidatorIndex,
			Signature:      voluntaryExit.Signature,
		}
	}

	logger.WithFields(logrus.Fields{
		"slot":            slot,
		"epoch":           slot / utils.Config.Chain.SlotsPerEpoch,
		"proposer":        b1.Message.ProposerIndex,
		"blockroot":       fmt.Sprintf("%x", b1Root.Root),
		"attestations":    len(b2.Attestations),
		"deposits":        len(b2.Deposits),
		"voluntaryExits":  len(b2.VoluntaryExits),
		"durGetHeader":    t3.Sub(t2),
		"durGetBlockRoot": t2.Sub(t1),
		"durGetBlock":     t1.Sub(t0),
		"dur":             time.Since(t0),
	}).Info("GetBlocksBySlot")

	return []*types.Block{&b2}, nil
}

func (c *Eth2ApiV1Client) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	// if err != nil {
	// 	logger.Printf("error retrieving epoch participation statistics: %v", err)
	// 	return &types.ValidatorParticipation{
	// 		Epoch:                   epoch,
	// 		Finalized:               false,
	// 		GlobalParticipationRate: 0,
	// 		VotedEther:              0,
	// 		EligibleEther:           0,
	// 	}, nil
	// }
	return &types.ValidatorParticipation{
		Epoch:                   epoch,
		Finalized:               false,
		GlobalParticipationRate: 0,
		VotedEther:              0,
		EligibleEther:           0,
	}, nil
}
