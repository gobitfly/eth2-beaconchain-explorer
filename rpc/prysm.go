package rpc

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-bitfield"
	"google.golang.org/grpc"

	ptypes "github.com/gogo/protobuf/types"
)

// PrysmClient holds information about the Prysm Client
type PrysmClient struct {
	client              ethpb.BeaconChainClient
	nodeClient          ethpb.NodeClient
	conn                *grpc.ClientConn
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
	validatorsCache     *lru.Cache
	validatorsCacheMux  *sync.Mutex
}

// NewPrysmClient is used for a new Prysm client connection
func NewPrysmClient(endpoint string) (*PrysmClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		// Maximum receive value 128 MB
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(128 * 1024 * 1024)),
	}
	conn, err := grpc.Dial(endpoint, dialOpts...)

	if err != nil {
		return nil, err
	}

	chainClient := ethpb.NewBeaconChainClient(conn)
	nodeClient := ethpb.NewNodeClient(conn)

	logger.Printf("gRPC connection to backend node established")
	client := &PrysmClient{
		client:              chainClient,
		nodeClient:          nodeClient,
		conn:                conn,
		assignmentsCacheMux: &sync.Mutex{},
		validatorsCacheMux:  &sync.Mutex{},
	}
	client.assignmentsCache, _ = lru.New(10)
	client.validatorsCache, _ = lru.New(10)

	return client, nil
}

// Close will close a Prysm client connection
func (pc *PrysmClient) Close() {
	pc.conn.Close()
}

// GetGenesisTimestamp returns the genesis timestamp of the beacon chain
func (pc *PrysmClient) GetGenesisTimestamp() (int64, error) {
	genesis, err := pc.nodeClient.GetGenesis(context.Background(), &ptypes.Empty{})

	if err != nil {
		return 0, err
	}

	return genesis.GenesisTime.Seconds, nil
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
func (pc *PrysmClient) GetValidatorQueue() (*types.ValidatorQueue, error) {
	var err error

	validators, err := pc.client.GetValidatorQueue(context.Background(), &ptypes.Empty{})

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	return &types.ValidatorQueue{
		ChurnLimit:                 validators.ChurnLimit,
		ActivationPublicKeys:       validators.ActivationPublicKeys,
		ExitPublicKeys:             validators.ExitPublicKeys,
		ActivationValidatorIndices: validators.ActivationValidatorIndices,
		ExitValidatorIndices:       validators.ExitValidatorIndices,
	}, nil
}

// GetAttestationPool will get the attestation pool from a Prysm client
func (pc *PrysmClient) GetAttestationPool() ([]*types.Attestation, error) {
	var err error

	attestationPoolResponse := &ethpb.AttestationPoolResponse{}

	attestations := []*types.Attestation{}

	for {
		attestationPoolResponse, err = pc.client.AttestationPool(context.Background(), &ethpb.AttestationPoolRequest{PageSize: utils.Config.Indexer.Node.PageSize, PageToken: attestationPoolResponse.NextPageToken})
		if err != nil {
			return nil, err
		}
		if attestationPoolResponse.TotalSize == 0 {
			break
		}
		for _, attestation := range attestationPoolResponse.Attestations {
			attestations = append(attestations, &types.Attestation{
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
				Signature: attestation.Signature,
			})
		}
		if attestationPoolResponse.NextPageToken == "" {
			break
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

	logger.Infof("caching assignements for epoch %v", epoch)
	start := time.Now()
	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// Retrieve the validator assignments for the epoch
	validatorAssignmentes := make([]*ethpb.ValidatorAssignments_CommitteeAssignment, 0)
	validatorAssignmentResponse := &ethpb.ValidatorAssignments{}
	validatorAssignmentRequest := &ethpb.ListValidatorAssignmentsRequest{PageToken: validatorAssignmentResponse.NextPageToken, PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListValidatorAssignmentsRequest_Epoch{Epoch: epoch}}
	if epoch == 0 {
		validatorAssignmentRequest.QueryFilter = &ethpb.ListValidatorAssignmentsRequest_Genesis{Genesis: true}
	}
	for {
		validatorAssignmentRequest.PageToken = validatorAssignmentResponse.NextPageToken
		validatorAssignmentResponse, err = pc.client.ListValidatorAssignments(context.Background(), validatorAssignmentRequest)
		if err != nil {
			return nil, fmt.Errorf("error retrieving validator assignment response for caching: %v", err)
		}

		validatorAssignmentes = append(validatorAssignmentes, validatorAssignmentResponse.Assignments...)
		//logger.Printf("retrieved %v assignments of %v for epoch %v", len(validatorAssignmentes), validatorAssignmentResponse.TotalSize, epoch)

		if validatorAssignmentResponse.NextPageToken == "" || validatorAssignmentResponse.TotalSize == 0 || len(validatorAssignmentes) == int(validatorAssignmentResponse.TotalSize) {
			break
		}
	}

	// Extract the proposer & attestation assignments from the response and cache them for later use
	// Proposer assignments are cached by the proposer slot
	// Attestation assignments are cached by the slot & committee key
	for _, assignment := range validatorAssignmentes {
		for _, slot := range assignment.ProposerSlots {
			assignments.ProposerAssignments[slot] = assignment.ValidatorIndex
		}

		for memberIndex, validatorIndex := range assignment.BeaconCommittees {
			assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(assignment.AttesterSlot, assignment.CommitteeIndex, uint64(memberIndex))] = validatorIndex
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		pc.assignmentsCache.Add(epoch, assignments)
	}

	logger.Infof("cached assignements for epoch %v took %v", epoch, time.Since(start))
	return assignments, nil
}

// GetEpochData will get the epoch data from a Prysm client
func (pc *PrysmClient) GetEpochData(epoch uint64) (*types.EpochData, error) {
	var err error

	data := &types.EpochData{}
	data.Epoch = epoch

	data.ValidatorIndices = make(map[string]uint64)

	// Retrieve the validator balances for the epoch (NOTE: Currently the API call is broken and allows only to retrieve the balances for the current epoch
	validatorBalancesByPubkey := make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	validatorBalancesRequest := &ethpb.ListValidatorBalancesRequest{PageSize: utils.Config.Indexer.Node.PageSize, PageToken: validatorBalancesResponse.NextPageToken, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: epoch}}
	if epoch == 0 {
		validatorBalancesRequest.QueryFilter = &ethpb.ListValidatorBalancesRequest_Genesis{Genesis: true}
	}
	for {
		validatorBalancesRequest.PageToken = validatorBalancesResponse.NextPageToken
		validatorBalancesResponse, err = pc.client.ListValidatorBalances(context.Background(), validatorBalancesRequest)
		if err != nil {
			logger.Printf("error retrieving validator balances for epoch %v: %v", epoch, err)
			break
		}
		if validatorBalancesResponse.TotalSize == 0 {
			break
		}

		for _, balance := range validatorBalancesResponse.Balances {
			data.ValidatorIndices[fmt.Sprintf("%x", balance.PublicKey)] = balance.Index
			validatorBalancesByPubkey[fmt.Sprintf("%x", balance.PublicKey)] = balance.Balance
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}
	logger.Printf("retrieved data for %v validator balances for epoch %v", len(validatorBalancesByPubkey), epoch)

	data.ValidatorAssignmentes, err = pc.GetEpochAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving assignments for epoch %v: %v", epoch, err)
	}
	logger.Printf("retrieved validator assignment data for epoch %v", epoch)

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.Config.Chain.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.SlotsPerEpoch-1; slot++ {
		blocks, err := pc.GetBlocksBySlot(slot)

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
	logger.Printf("retrieved %v blocks for epoch %v", len(data.Blocks), epoch)

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

	cachedValue, found := pc.validatorsCache.Get(epoch)
	if found {
		data.Validators = cachedValue.([]*types.Validator)
	} else {
		// Retrieve the validator set for the epoch
		data.Validators = make([]*types.Validator, 0)
		validatorResponse := &ethpb.Validators{}
		validatorRequest := &ethpb.ListValidatorsRequest{PageToken: validatorResponse.NextPageToken, PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListValidatorsRequest_Epoch{Epoch: epoch}}
		if epoch == 0 {
			validatorRequest.QueryFilter = &ethpb.ListValidatorsRequest_Genesis{Genesis: true}
		}
		for {
			validatorRequest.PageToken = validatorResponse.NextPageToken
			validatorResponse, err = pc.client.ListValidators(context.Background(), validatorRequest)
			if err != nil {
				logger.Errorf("error retrieving validator response: %v", err)
				break
			}
			if validatorResponse.TotalSize == 0 {
				break
			}

			for _, validator := range validatorResponse.ValidatorList {
				balance, exists := validatorBalancesByPubkey[fmt.Sprintf("%x", validator.Validator.PublicKey)]
				if !exists {
					logger.WithField("pubkey", fmt.Sprintf("%x", validator.Validator.PublicKey)).WithField("epoch", epoch).Errorf("error retrieving validator balance")
					continue
				}
				data.Validators = append(data.Validators, &types.Validator{
					Index:                      validator.Index,
					PublicKey:                  validator.Validator.PublicKey,
					WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
					Balance:                    balance,
					EffectiveBalance:           validator.Validator.EffectiveBalance,
					Slashed:                    validator.Validator.Slashed,
					ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
					ActivationEpoch:            validator.Validator.ActivationEpoch,
					ExitEpoch:                  validator.Validator.ExitEpoch,
					WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
				})
			}

			if validatorResponse.NextPageToken == "" {
				break
			}
		}
		pc.validatorsCache.Add(epoch, data.Validators)
		logger.Printf("retrieved data for %v validators for epoch %v", len(data.Validators), epoch)
	}

	// Retrieve the beacon committees for the epoch
	data.BeaconCommittees = make(map[uint64][]*types.BeaconCommitteItem)
	beaconCommitteesResponse := &ethpb.BeaconCommittees{}
	beaconCommitteesRequest := &ethpb.ListCommitteesRequest{QueryFilter: &ethpb.ListCommitteesRequest_Epoch{Epoch: epoch}}
	if epoch == 0 {
		beaconCommitteesRequest.QueryFilter = &ethpb.ListCommitteesRequest_Genesis{Genesis: true}
	}
	beaconCommitteesResponse, err = pc.client.ListBeaconCommittees(context.Background(), beaconCommitteesRequest)
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
	logger.Infof("Retrieving block at slot %v", slot)

	blocks := make([]*types.Block, 0)

	blocksRequest := &ethpb.ListBlocksRequest{PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Slot{Slot: slot}}
	if slot == 0 {
		blocksRequest.QueryFilter = &ethpb.ListBlocksRequest_Genesis{Genesis: true}
	}
	blocksResponse, err := pc.client.ListBlocks(context.Background(), blocksRequest)
	if err != nil {
		return nil, err
	}

	if blocksResponse.TotalSize == 0 {
		return blocks, nil
	}

	for _, block := range blocksResponse.BlockContainers {

		// Make sure that blocks from the genesis epoch have their Eth1Data field set
		if block.Block.Block.Body.Eth1Data == nil {
			block.Block.Block.Body.Eth1Data = &ethpb.Eth1Data{
				DepositRoot:  []byte{},
				DepositCount: 0,
				BlockHash:    []byte{},
			}
		}

		b := &types.Block{
			Status:       1,
			BlockRoot:    block.BlockRoot,
			Slot:         block.Block.Block.Slot,
			ParentRoot:   block.Block.Block.ParentRoot,
			StateRoot:    block.Block.Block.StateRoot,
			Signature:    block.Block.Signature,
			RandaoReveal: block.Block.Block.Body.RandaoReveal,
			Graffiti:     block.Block.Block.Body.Graffiti,
			Eth1Data: &types.Eth1Data{
				DepositRoot:  block.Block.Block.Body.Eth1Data.DepositRoot,
				DepositCount: block.Block.Block.Body.Eth1Data.DepositCount,
				BlockHash:    block.Block.Block.Body.Eth1Data.BlockHash,
			},
			ProposerSlashings: make([]*types.ProposerSlashing, len(block.Block.Block.Body.ProposerSlashings)),
			AttesterSlashings: make([]*types.AttesterSlashing, len(block.Block.Block.Body.AttesterSlashings)),
			Attestations:      make([]*types.Attestation, len(block.Block.Block.Body.Attestations)),
			Deposits:          make([]*types.Deposit, len(block.Block.Block.Body.Deposits)),
			VoluntaryExits:    make([]*types.VoluntaryExit, len(block.Block.Block.Body.VoluntaryExits)),
			Proposer:          block.Block.Block.ProposerIndex,
		}

		for i, proposerSlashing := range block.Block.Block.Body.ProposerSlashings {
			b.ProposerSlashings[i] = &types.ProposerSlashing{
				ProposerIndex: proposerSlashing.Header_1.Header.ProposerIndex,
				Header1: &types.Block{
					Slot:       proposerSlashing.Header_1.Header.Slot,
					ParentRoot: proposerSlashing.Header_1.Header.ParentRoot,
					StateRoot:  proposerSlashing.Header_1.Header.StateRoot,
					Signature:  proposerSlashing.Header_1.Signature,
					BodyRoot:   proposerSlashing.Header_1.Header.BodyRoot,
				},
				Header2: &types.Block{
					Slot:       proposerSlashing.Header_2.Header.Slot,
					ParentRoot: proposerSlashing.Header_2.Header.ParentRoot,
					StateRoot:  proposerSlashing.Header_2.Header.StateRoot,
					Signature:  proposerSlashing.Header_2.Signature,
					BodyRoot:   proposerSlashing.Header_2.Header.BodyRoot,
				},
			}
		}

		for i, attesterSlashing := range block.Block.Block.Body.AttesterSlashings {
			b.AttesterSlashings[i] = &types.AttesterSlashing{
				Attestation1: &types.IndexedAttestation{
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
					Signature:        attesterSlashing.Attestation_1.Signature,
					AttestingIndices: attesterSlashing.Attestation_1.AttestingIndices,
				},
				Attestation2: &types.IndexedAttestation{
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
					Signature:        attesterSlashing.Attestation_2.Signature,
					AttestingIndices: attesterSlashing.Attestation_2.AttestingIndices,
				},
			}
		}

		for i, attestation := range block.Block.Block.Body.Attestations {
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
				Signature: attestation.Signature,
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
		for i, deposit := range block.Block.Block.Body.Deposits {
			b.Deposits[i] = &types.Deposit{
				Proof:                 deposit.Proof,
				PublicKey:             deposit.Data.PublicKey,
				WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
				Amount:                deposit.Data.Amount,
				Signature:             deposit.Data.Signature,
			}
		}

		for i, voluntaryExit := range block.Block.Block.Body.VoluntaryExits {
			b.VoluntaryExits[i] = &types.VoluntaryExit{
				Epoch:          voluntaryExit.Exit.Epoch,
				ValidatorIndex: voluntaryExit.Exit.ValidatorIndex,
				Signature:      voluntaryExit.Signature,
			}
		}

		blocks = append(blocks, b)
	}

	return blocks, nil
}

// GetValidatorParticipation will get the validator participation from Prysm client
func (pc *PrysmClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	validatorParticipationRequest := &ethpb.GetValidatorParticipationRequest{QueryFilter: &ethpb.GetValidatorParticipationRequest_Epoch{Epoch: epoch}}
	if epoch == 0 {
		validatorParticipationRequest.QueryFilter = &ethpb.GetValidatorParticipationRequest_Genesis{Genesis: true}
	}
	epochParticipationStatistics, err := pc.client.GetValidatorParticipation(context.Background(), validatorParticipationRequest)
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
