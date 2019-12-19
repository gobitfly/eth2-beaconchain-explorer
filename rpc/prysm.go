package rpc

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-bitfield"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"

	ptypes "github.com/golang/protobuf/ptypes/empty"
)

// PrysmClient holds information about the Prysm Client
type PrysmClient struct {
	client              ethpb.BeaconChainClient
	conn                *grpc.ClientConn
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
}

// NewPrysmClient is used for a new Prysm client connection
func NewPrysmClient(endpoint string) (*PrysmClient, error) {
	dialOpt := grpc.WithInsecure()
	conn, err := grpc.Dial(endpoint, dialOpt)

	if err != nil {
		return nil, err
	}

	chainClient := ethpb.NewBeaconChainClient(conn)

	log.Printf("gRPC connection to backend node established")
	client := &PrysmClient{
		client:              chainClient,
		conn:                conn,
		assignmentsCacheMux: &sync.Mutex{},
	}
	client.assignmentsCache, _ = lru.New(128)

	return client, nil
}

// Close will close a Prysm client connection
func (pc *PrysmClient) Close() {
	pc.conn.Close()
}

// GetChainHead will get the chain head from a Prysm client
func (pc *PrysmClient) GetChainHead() (*types.ChainHead, error) {
	headResponse, err := pc.client.GetChainHead(context.Background(), &ptypes.Empty{})

	if err != nil {
		return nil, err
	}

	return &types.ChainHead{
		HeadSlot:                   headResponse.HeadSlot,
		HeadEpoch:                  headResponse.HeadEpoch,
		HeadBlockRoot:              headResponse.HeadBlockRoot,
		FinalizedSlot:              headResponse.FinalizedSlot,
		FinalizedEpoch:             headResponse.FinalizedEpoch,
		FinalizedBlockRoot:         headResponse.FinalizedBlockRoot,
		JustifiedSlot:              headResponse.JustifiedSlot,
		JustifiedEpoch:             headResponse.JustifiedEpoch,
		JustifiedBlockRoot:         headResponse.JustifiedBlockRoot,
		PreviousJustifiedSlot:      headResponse.PreviousJustifiedSlot,
		PreviousJustifiedEpoch:     headResponse.PreviousJustifiedEpoch,
		PreviousJustifiedBlockRoot: headResponse.PreviousJustifiedBlockRoot,
	}, nil
}

// GetValidatorQueue will get the validator queue from a Prysm client
func (pc *PrysmClient) GetValidatorQueue() (*types.ValidatorQueue, map[string]uint64, error) {
	var err error

	validatorIndices := make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = pc.client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize})
		if err != nil {
			return nil, nil, fmt.Errorf("error retrieving validator balances response: %v", err)
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

	validators, err := pc.client.GetValidatorQueue(context.Background(), &ptypes.Empty{})

	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	return &types.ValidatorQueue{
		ChurnLimit:           validators.ChurnLimit,
		ActivationPublicKeys: validators.ActivationPublicKeys,
		ExitPublicKeys:       validators.ExitPublicKeys,
	}, validatorIndices, nil
}

// GetAttestationPool will get the attestation pool from a Prysm client
func (pc *PrysmClient) GetAttestationPool() ([]*types.Attestation, error) {
	attestationsResponse, err := pc.client.AttestationPool(context.Background(), &ptypes.Empty{})

	if err != nil {
		return nil, fmt.Errorf("error retrieving attestation pool data: %v", err)
	}

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

	return attestations, nil
}

// GetEpochAssignments will get the epoch assignments from a Prysm client
func (pc *PrysmClient) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {

	pc.assignmentsCacheMux.Lock()
	defer pc.assignmentsCacheMux.Unlock()

	var err error

	cachedValue, found := pc.assignmentsCache.Get(epoch)
	if found {
		return cachedValue.(*types.EpochAssignments), nil
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// Retrieve the currently active validator set in order to map public keys to indexes
	validators := make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = pc.client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: epoch}})
		if err != nil {
			logger.Printf("error retrieving validator balances response: %v", err)
			break
		}
		if validatorBalancesResponse.TotalSize == 0 {
			break
		}

		for _, balance := range validatorBalancesResponse.Balances {
			logger.Debugf("%x - %v", balance.PublicKey, balance.Index)
			validators[utils.FormatPublicKey(balance.PublicKey)] = balance.Index
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}

	// Retrieve the validator assignments for the epoch
	validatorAssignmentes := make([]*ethpb.ValidatorAssignments_CommitteeAssignment, 0)
	validatorAssignmentResponse := &ethpb.ValidatorAssignments{}

	for validatorAssignmentResponse.NextPageToken == "" || len(validatorAssignmentes) < int(validatorAssignmentResponse.TotalSize) {
		validatorAssignmentResponse, err = pc.client.ListValidatorAssignments(context.Background(), &ethpb.ListValidatorAssignmentsRequest{PageToken: validatorAssignmentResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorAssignmentsRequest_Epoch{Epoch: epoch}})
		if err != nil {
			return nil, fmt.Errorf("error retrieving validator assignment response for caching: %v", err)
		}
		if validatorAssignmentResponse.TotalSize == 0 || len(validatorAssignmentes) == int(validatorAssignmentResponse.TotalSize) {
			break
		}
		validatorAssignmentes = append(validatorAssignmentes, validatorAssignmentResponse.Assignments...)
		logger.Printf("Retrieved %v assignments of %v for epoch %v", len(validatorAssignmentes), validatorAssignmentResponse.TotalSize, epoch)
	}

	// Extract the proposer & attestation assignments from the response and cache them for later use
	// Proposer assignments are cached by the proposer slot
	// Attestation assignments are cached by the slot & committee key
	for _, assignment := range validatorAssignmentes {
		if assignment.ProposerSlot > 0 {
			assignments.ProposerAssignments[assignment.ProposerSlot] = validators[utils.FormatPublicKey(assignment.PublicKey)]
		}

		if assignment.AttesterSlot > 0 {
			for memberIndex, validatorIndex := range assignment.BeaconCommittees {
				assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(assignment.AttesterSlot, assignment.CommitteeIndex, uint64(memberIndex))] = validatorIndex
			}
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		pc.assignmentsCache.Add(epoch, assignments)
	}

	return assignments, nil
}

// GetEpochData will get the epoch data from a Prysm client
func (pc *PrysmClient) GetEpochData(epoch uint64) (*types.EpochData, error) {
	var err error

	data := &types.EpochData{}
	data.Epoch = epoch

	// Retrieve the validator balances for the epoch (NOTE: Currently the API call is broken and allows only to retrieve the balances for the current epoch
	data.ValidatorBalances = make([]*types.ValidatorBalance, 0)
	data.ValidatorIndices = make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = pc.client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: epoch}})
		if err != nil {
			return nil, err
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

	data.ValidatorAssignmentes, err = pc.GetEpochAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving assignments for epoch %v: %v", epoch, err)
	}
	logger.Printf("Retrieved validator assignment data for epoch %v", epoch)

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.Config.Chain.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.SlotsPerEpoch-1; slot++ {

		if slot == 0 { // Currently slot 0 returns all blocks
			continue
		}

		blocks, err := pc.GetBlocksBySlot(slot)

		if err != nil {
			return nil, err
		}

		for _, block := range blocks {
			if data.Blocks[block.Slot] == nil {
				data.Blocks[block.Slot] = make(map[string]*types.Block)
			}

			block.Proposer = data.ValidatorAssignmentes.ProposerAssignments[block.Slot]
			data.Blocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block
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
		validatorResponse, err = pc.client.ListValidators(context.Background(), &ethpb.ListValidatorsRequest{PageToken: validatorResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorsRequest_Epoch{Epoch: epoch}})
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
	beaconCommitteesResponse, err = pc.client.ListBeaconCommittees(context.Background(), &ethpb.ListCommitteesRequest{QueryFilter: &ethpb.ListCommitteesRequest_Epoch{Epoch: epoch}})
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

	data.EpochParticipationStats, err = pc.GetValidatorParticipation(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %v", epoch, err)
	}

	return data, nil
}

// GetBlocksBySlot will get blocks by slot from a Prysm client
func (pc *PrysmClient) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {

	blocks := make([]*types.Block, 0)
	blocksResponse, err := pc.client.ListBlocks(context.Background(), &ethpb.ListBlocksRequest{PageSize: utils.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Slot{Slot: slot}, IncludeNoncanonical: true})
	if err != nil {
		return nil, err
	}

	if blocksResponse.TotalSize == 0 {
		return blocks, nil
	}

	for _, block := range blocksResponse.BlockContainers {

		// Make sure that blocks from the genesis epoch have their Eth1Data field set
		if block.Block.Body.Eth1Data == nil {
			block.Block.Body.Eth1Data = &ethpb.Eth1Data{
				DepositRoot:  []byte{},
				DepositCount: 0,
				BlockHash:    []byte{},
			}
		}

		b := &types.Block{
			Status:       1,
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
				Header1: &types.Block{
					Slot:       proposerSlashing.Header_1.Slot,
					ParentRoot: proposerSlashing.Header_1.ParentRoot,
					StateRoot:  proposerSlashing.Header_1.StateRoot,
					Signature:  proposerSlashing.Header_1.Signature,
					BodyRoot:   proposerSlashing.Header_1.BodyRoot,
				},
				Header2: &types.Block{
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
				Attestation1: &types.IndexedAttestation{
					Custodybit0indices: attesterSlashing.Attestation_1.CustodyBit_0Indices,
					Custodybit1indices: attesterSlashing.Attestation_1.CustodyBit_1Indices,
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
				Attestation2: &types.IndexedAttestation{
					Custodybit0indices: attesterSlashing.Attestation_2.CustodyBit_0Indices,
					Custodybit1indices: attesterSlashing.Attestation_2.CustodyBit_1Indices,
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
			a := &types.Attestation{
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

			aggregationBits := bitfield.Bitlist(a.AggregationBits)
			assignments, err := pc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.SlotsPerEpoch)
			if err != nil {
				return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.SlotsPerEpoch, err)
			}

			a.Attesters = make([]uint64, 0)
			for i := uint64(0); i < aggregationBits.Len(); i++ {
				if aggregationBits.BitAt(i) {
					validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, a.Data.CommitteeIndex, i)]
					if !found { // This should never happen!
						validator = 0
						logger.Errorf("error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v", i, b.Slot, a.Data.Slot, a.Data.CommitteeIndex, i)
					}
					a.Attesters = append(a.Attesters, validator)
				}
			}

			b.Attestations[i] = a
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

		blocks = append(blocks, b)
	}

	return blocks, nil
}

// GetValidatorParticipation will get the validator participation from Prysm client
func (pc *PrysmClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	epochParticipationStatistics, err := pc.client.GetValidatorParticipation(context.Background(), &ethpb.GetValidatorParticipationRequest{QueryFilter: &ethpb.GetValidatorParticipationRequest_Epoch{Epoch: epoch}})
	if err != nil {
		logger.Printf("error retrieving epoch participation statistics: %v", err)
		return &types.ValidatorParticipation{
			Epoch:                   epoch,
			Finalized:               false,
			GlobalParticipationRate: 0,
			VotedEther:              0,
			EligibleEther:           0,
		}, nil
	}
	return &types.ValidatorParticipation{
		Epoch:                   epoch,
		Finalized:               epochParticipationStatistics.Finalized,
		GlobalParticipationRate: epochParticipationStatistics.Participation.GlobalParticipationRate,
		VotedEther:              epochParticipationStatistics.Participation.VotedEther,
		EligibleEther:           epochParticipationStatistics.Participation.EligibleEther,
	}, nil
}
