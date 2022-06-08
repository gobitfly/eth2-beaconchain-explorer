package rpc

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"sync"
	"time"

	gtypes "github.com/ethereum/go-ethereum/core/types"

	lru "github.com/hashicorp/golang-lru"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"

	"github.com/prysmaticlabs/go-bitfield"
	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	eth2types "github.com/prysmaticlabs/eth2-types"
)

// PrysmClient holds information about the Prysm Client
type PrysmClient struct {
	client              ethpb.BeaconChainClient
	nodeClient          ethpb.NodeClient
	conn                *grpc.ClientConn
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
	newBlockChan        chan *types.Block
	signer              gtypes.Signer
}

// NewPrysmClient is used for a new Prysm client connection
func NewPrysmClient(endpoint string, chainId *big.Int) (*PrysmClient, error) {
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
		newBlockChan:        make(chan *types.Block, 1000),
		signer:              gtypes.NewLondonSigner(chainId),
	}
	client.assignmentsCache, _ = lru.New(10)

	streamChainHeadClient, err := chainClient.StreamChainHead(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			head, err := streamChainHeadClient.Recv()

			if err != nil {
				logger.Errorf("error receiving from chain head stream: %v", err)

				// in order to recover from a stream error we wait for a second and then re-create the stream
				time.Sleep(time.Second)
				streamChainHeadClient, err = chainClient.StreamChainHead(context.Background(), &empty.Empty{})
				for err != nil {
					logger.Errorf("error initializing chain head stream: %v. retrying in 1s...", err)
					time.Sleep(time.Second)
					streamChainHeadClient, err = chainClient.StreamChainHead(context.Background(), &empty.Empty{})
				}
				continue
			}

			blocks, err := client.GetBlocksBySlot(uint64(head.HeadSlot))

			if err != nil {
				logger.Errorf("error receiving blocks via chain head stream: %v", err)
				continue
			}

			for _, b := range blocks {
				logger.Infof("received block at slot %v with hash %x via stream", blocks[0].Slot, blocks[0].BlockRoot)
				client.newBlockChan <- b
			}
		}
	}()
	return client, nil
}

// Close will close a Prysm client connection
func (pc *PrysmClient) Close() {
	pc.conn.Close()
}

func (pc *PrysmClient) GetNewBlockChan() chan *types.Block {
	return pc.newBlockChan
}

// GetGenesisTimestamp returns the genesis timestamp of the beacon chain
func (pc *PrysmClient) GetGenesisTimestamp() (int64, error) {
	genesis, err := pc.nodeClient.GetGenesis(context.Background(), &empty.Empty{})

	if err != nil {
		return 0, err
	}

	return genesis.GenesisTime.Seconds, nil
}

// GetChainHead will get the chain head from a Prysm client
func (pc *PrysmClient) GetChainHead() (*types.ChainHead, error) {
	headResponse, err := pc.client.GetChainHead(context.Background(), &empty.Empty{})

	if err != nil {
		return nil, err
	}

	return &types.ChainHead{
		HeadSlot:                   uint64(headResponse.HeadSlot),
		HeadEpoch:                  uint64(headResponse.HeadEpoch),
		HeadBlockRoot:              headResponse.HeadBlockRoot,
		FinalizedSlot:              uint64(headResponse.FinalizedSlot),
		FinalizedEpoch:             uint64(headResponse.FinalizedEpoch),
		FinalizedBlockRoot:         headResponse.FinalizedBlockRoot,
		JustifiedSlot:              uint64(headResponse.JustifiedSlot),
		JustifiedEpoch:             uint64(headResponse.JustifiedEpoch),
		JustifiedBlockRoot:         headResponse.JustifiedBlockRoot,
		PreviousJustifiedSlot:      uint64(headResponse.PreviousJustifiedSlot),
		PreviousJustifiedEpoch:     uint64(headResponse.PreviousJustifiedEpoch),
		PreviousJustifiedBlockRoot: headResponse.PreviousJustifiedBlockRoot,
	}, nil
}

// GetValidatorQueue will get the validator queue from a Prysm client
func (pc *PrysmClient) GetValidatorQueue() (*types.ValidatorQueue, error) {
	var err error

	validators, err := pc.client.GetValidatorQueue(context.Background(), &empty.Empty{})

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	return &types.ValidatorQueue{
		Activating: uint64(len(validators.ActivationPublicKeys)),
		Exititing:  uint64(len(validators.ExitPublicKeys)),
	}, nil
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

	logger.Infof("caching assignments for epoch %v", epoch)
	start := time.Now()
	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// Retrieve the validator assignments for the epoch
	validatorAssignmentes := make([]*ethpb.ValidatorAssignments_CommitteeAssignment, 0)
	validatorAssignmentResponse := &ethpb.ValidatorAssignments{}
	validatorAssignmentRequest := &ethpb.ListValidatorAssignmentsRequest{PageToken: validatorAssignmentResponse.NextPageToken, PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListValidatorAssignmentsRequest_Epoch{Epoch: eth2types.Epoch(epoch)}}
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
			assignments.ProposerAssignments[uint64(slot)] = uint64(assignment.ValidatorIndex)
		}

		for memberIndex, validatorIndex := range assignment.BeaconCommittees {
			assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(uint64(assignment.AttesterSlot), uint64(assignment.CommitteeIndex), uint64(memberIndex))] = uint64(validatorIndex)
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		pc.assignmentsCache.Add(epoch, assignments)
	}

	logger.Infof("cached assignments for epoch %v took %v", epoch, time.Since(start))
	return assignments, nil
}

// GetEpochData will get the epoch data from a Prysm client
func (pc *PrysmClient) GetEpochData(epoch uint64) (*types.EpochData, error) {
	var err error

	data := &types.EpochData{}
	data.Epoch = epoch

	// Retrieve the validator balances for the requested epoch
	start := time.Now()
	validatorBalances, err := pc.getBalancesForEpoch(int64(epoch))
	logger.Printf("retrieved data for %v validator balances for epoch %v took %v", len(validatorBalances), epoch, time.Since(start))

	// Retrieve the validator balances for the n-1d epoch
	start = time.Now()
	epoch1d := int64(epoch) - 225
	validatorBalances1d, err := pc.getBalancesForEpoch(epoch1d)
	logger.Printf("retrieved data for %v validator balances for 1d epoch %v took %v", len(validatorBalances), epoch1d, time.Since(start))

	// Retrieve the validator balances for the n-7d epoch
	start = time.Now()
	epoch7d := int64(epoch) - 225*7
	validatorBalances7d, err := pc.getBalancesForEpoch(epoch7d)
	logger.Printf("retrieved data for %v validator balances for 7d epoch %v took %v", len(validatorBalances), epoch7d, time.Since(start))

	// Retrieve the validator balances for the n-7d epoch
	start = time.Now()
	epoch31d := int64(epoch) - 225*31
	validatorBalances31d, err := pc.getBalancesForEpoch(epoch31d)
	logger.Printf("retrieved data for %v validator balances for 31d epoch %v took %v", len(validatorBalances), epoch31d, time.Since(start))

	data.ValidatorAssignmentes, err = pc.GetEpochAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving assignments for epoch %v: %v", epoch, err)
	}
	logger.Printf("retrieved validator assignment data for epoch %v", epoch)

	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	for slot := epoch * utils.Config.Chain.Config.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch-1; slot++ {
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

	// Retrieve the validator set for the epoch
	data.Validators = make([]*types.Validator, 0)
	validatorResponse := &ethpb.Validators{}
	validatorRequest := &ethpb.ListValidatorsRequest{PageToken: validatorResponse.NextPageToken, PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListValidatorsRequest_Epoch{Epoch: eth2types.Epoch(epoch)}}
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

			balance, exists := validatorBalances[uint64(validator.Index)]
			if !exists {
				logger.WithField("pubkey", fmt.Sprintf("%x", validator.Validator.PublicKey)).WithField("epoch", epoch).Errorf("error retrieving validator balance")
				continue
			}

			val := &types.Validator{
				Index:                      uint64(validator.Index),
				PublicKey:                  validator.Validator.PublicKey,
				WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
				Balance:                    balance,
				EffectiveBalance:           validator.Validator.EffectiveBalance,
				Slashed:                    validator.Validator.Slashed,
				ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
				ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
				ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
				WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
			}

			val.Balance1d = validatorBalances1d[uint64(validator.Index)]
			val.Balance7d = validatorBalances7d[uint64(validator.Index)]
			val.Balance31d = validatorBalances31d[uint64(validator.Index)]

			data.Validators = append(data.Validators, val)

		}

		if validatorResponse.NextPageToken == "" {
			break
		}
	}
	logger.Printf("retrieved data for %v validators for epoch %v", len(data.Validators), epoch)

	data.EpochParticipationStats, err = pc.GetValidatorParticipation(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %v", epoch, err)
	}

	return data, nil
}

func (pc *PrysmClient) getBalancesForEpoch(epoch int64) (map[uint64]uint64, error) {

	if epoch < 0 {
		epoch = 0
	}

	var err error

	validatorBalances := make(map[uint64]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	validatorBalancesRequest := &ethpb.ListValidatorBalancesRequest{PageSize: utils.Config.Indexer.Node.PageSize, PageToken: validatorBalancesResponse.NextPageToken, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: eth2types.Epoch(epoch)}}
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
			validatorBalances[uint64(balance.Index)] = balance.Balance
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}
	return validatorBalances, err
}

// GetBlocksBySlot will get blocks by slot from a Prysm client
func (pc *PrysmClient) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {
	logger.Infof("retrieving block at slot %v", slot)

	blocks := make([]*types.Block, 0)

	blocksRequest := &ethpb.ListBlocksRequest{PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Slot{Slot: eth2types.Slot(slot)}}
	if slot == 0 {
		blocksRequest.QueryFilter = &ethpb.ListBlocksRequest_Genesis{Genesis: true}
	}

	// blocksResponse, err := pc.client.ListBlocks(context.Background(), blocksRequest)
	blocksResponse, err := pc.client.ListBlocksAltair(context.Background(), blocksRequest)
	if err != nil {
		return nil, err
	}

	if blocksResponse.TotalSize == 0 {
		return blocks, nil
	}

	for _, block := range blocksResponse.BlockContainers {
		// Make sure that blocks from the genesis epoch have their Eth1Data field set
		blk := block.GetAltairBlock()
		if blk != nil && blk.Block.Body.Eth1Data == nil {
			blk.Block.Body.Eth1Data = &ethpb.Eth1Data{
				DepositRoot:  []byte{},
				DepositCount: 0,
				BlockHash:    []byte{},
			}
		}

		b, err := pc.parseRpcBlock(block)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, b)
	}

	return blocks, nil
}

// GetBlockStatusBySlot will get blocks by slot from a Prysm client
func (pc *PrysmClient) GetBlockStatusByEpoch(epoch uint64) ([]*types.CanonBlock, error) {
	logger.Infof("retrieving blocks for epoch %v", epoch)

	blocks := make([]*types.CanonBlock, 0)

	blocksRequest := &ethpb.ListBlocksRequest{PageSize: utils.Config.Indexer.Node.PageSize, QueryFilter: &ethpb.ListBlocksRequest_Epoch{Epoch: eth2types.Epoch(epoch)}}

	blocksResponse, err := pc.client.ListBlocks(context.Background(), blocksRequest)
	if err != nil {
		return nil, err
	}

	if blocksResponse.TotalSize == 0 {
		return blocks, nil
	}

	for _, block := range blocksResponse.BlockContainers {
		blocks = append(blocks, &types.CanonBlock{
			BlockRoot: block.BlockRoot,
			Slot:      uint64(block.Block.Block.Slot),
			Canonical: block.Canonical,
		})
	}

	return blocks, nil
}

func (pc *PrysmClient) parseRpcBlock(block *ethpb.BeaconBlockContainerAltair) (*types.Block, error) {
	phase0Block := block.GetPhase0Block()
	if phase0Block != nil {
		return pc.parsePhase0Block(block)
	}
	altairBlock := block.GetAltairBlock()
	if altairBlock != nil {
		return pc.parseAltairBlock(block)
	}
	// TODO mergeBlock
	return nil, fmt.Errorf("block is neither phase0 nor altair")
}

func (pc *PrysmClient) parsePhase0Block(block *ethpb.BeaconBlockContainerAltair) (*types.Block, error) {
	blk := block.GetPhase0Block()
	if blk == nil {
		return nil, fmt.Errorf("failed getting phase0 block")
	}
	b := &types.Block{
		Status:       1,
		Canonical:    block.Canonical,
		BlockRoot:    block.BlockRoot,
		Slot:         uint64(blk.Block.Slot),
		ParentRoot:   blk.Block.ParentRoot,
		StateRoot:    blk.Block.StateRoot,
		Signature:    blk.Signature,
		RandaoReveal: blk.Block.Body.RandaoReveal,
		Graffiti:     blk.Block.Body.Graffiti,
		Eth1Data: &types.Eth1Data{
			DepositRoot:  blk.Block.Body.Eth1Data.DepositRoot,
			DepositCount: blk.Block.Body.Eth1Data.DepositCount,
			BlockHash:    blk.Block.Body.Eth1Data.BlockHash,
		},
		ProposerSlashings: make([]*types.ProposerSlashing, len(blk.Block.Body.ProposerSlashings)),
		AttesterSlashings: make([]*types.AttesterSlashing, len(blk.Block.Body.AttesterSlashings)),
		Attestations:      make([]*types.Attestation, len(blk.Block.Body.Attestations)),
		Deposits:          make([]*types.Deposit, len(blk.Block.Body.Deposits)),
		VoluntaryExits:    make([]*types.VoluntaryExit, len(blk.Block.Body.VoluntaryExits)),
		Proposer:          uint64(blk.Block.ProposerIndex),
	}

	for i, proposerSlashing := range blk.Block.Body.ProposerSlashings {
		b.ProposerSlashings[i] = &types.ProposerSlashing{
			ProposerIndex: uint64(proposerSlashing.Header_1.Header.ProposerIndex),
			Header1: &types.Block{
				Slot:       uint64(proposerSlashing.Header_1.Header.Slot),
				ParentRoot: proposerSlashing.Header_1.Header.ParentRoot,
				StateRoot:  proposerSlashing.Header_1.Header.StateRoot,
				Signature:  proposerSlashing.Header_1.Signature,
				BodyRoot:   proposerSlashing.Header_1.Header.BodyRoot,
			},
			Header2: &types.Block{
				Slot:       uint64(proposerSlashing.Header_2.Header.Slot),
				ParentRoot: proposerSlashing.Header_2.Header.ParentRoot,
				StateRoot:  proposerSlashing.Header_2.Header.StateRoot,
				Signature:  proposerSlashing.Header_2.Signature,
				BodyRoot:   proposerSlashing.Header_2.Header.BodyRoot,
			},
		}
	}

	for i, attesterSlashing := range blk.Block.Body.AttesterSlashings {
		b.AttesterSlashings[i] = &types.AttesterSlashing{
			Attestation1: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation_1.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation_1.Data.CommitteeIndex),
					BeaconBlockRoot: attesterSlashing.Attestation_1.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_1.Data.Source.Epoch),
						Root:  attesterSlashing.Attestation_1.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_1.Data.Target.Epoch),
						Root:  attesterSlashing.Attestation_1.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation_1.Signature,
				AttestingIndices: attesterSlashing.Attestation_1.AttestingIndices,
			},
			Attestation2: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation_2.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation_2.Data.CommitteeIndex),
					BeaconBlockRoot: attesterSlashing.Attestation_2.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_2.Data.Source.Epoch),
						Root:  attesterSlashing.Attestation_2.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_2.Data.Target.Epoch),
						Root:  attesterSlashing.Attestation_2.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation_2.Signature,
				AttestingIndices: attesterSlashing.Attestation_2.AttestingIndices,
			},
		}
	}

	for i, attestation := range blk.Block.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: attestation.AggregationBits,
			Data: &types.AttestationData{
				Slot:            uint64(attestation.Data.Slot),
				CommitteeIndex:  uint64(attestation.Data.CommitteeIndex),
				BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
				Source: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Source.Epoch),
					Root:  attestation.Data.Source.Root,
				},
				Target: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Target.Epoch),
					Root:  attestation.Data.Target.Root,
				},
			},
			Signature: attestation.Signature,
		}

		aggregationBits := bitfield.Bitlist(a.AggregationBits)
		assignments, err := pc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.Config.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.Config.SlotsPerEpoch, err)
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
	for i, deposit := range blk.Block.Body.Deposits {
		b.Deposits[i] = &types.Deposit{
			Proof:                 deposit.Proof,
			PublicKey:             deposit.Data.PublicKey,
			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
			Amount:                deposit.Data.Amount,
			Signature:             deposit.Data.Signature,
		}
	}

	for i, voluntaryExit := range blk.Block.Body.VoluntaryExits {
		b.VoluntaryExits[i] = &types.VoluntaryExit{
			Epoch:          uint64(voluntaryExit.Exit.Epoch),
			ValidatorIndex: uint64(voluntaryExit.Exit.ValidatorIndex),
			Signature:      voluntaryExit.Signature,
		}
	}
	return b, nil
}

func (pc *PrysmClient) parseAltairBlock(block *ethpb.BeaconBlockContainerAltair) (*types.Block, error) {
	blk := block.GetAltairBlock()
	if blk == nil {
		return nil, fmt.Errorf("failed getting altair block")
	}
	b := &types.Block{
		Status:       1,
		Canonical:    block.Canonical,
		BlockRoot:    block.BlockRoot,
		Slot:         uint64(blk.Block.Slot),
		ParentRoot:   blk.Block.ParentRoot,
		StateRoot:    blk.Block.StateRoot,
		Signature:    blk.Signature,
		RandaoReveal: blk.Block.Body.RandaoReveal,
		Graffiti:     blk.Block.Body.Graffiti,
		Eth1Data: &types.Eth1Data{
			DepositRoot:  blk.Block.Body.Eth1Data.DepositRoot,
			DepositCount: blk.Block.Body.Eth1Data.DepositCount,
			BlockHash:    blk.Block.Body.Eth1Data.BlockHash,
		},
		ProposerSlashings: make([]*types.ProposerSlashing, len(blk.Block.Body.ProposerSlashings)),
		AttesterSlashings: make([]*types.AttesterSlashing, len(blk.Block.Body.AttesterSlashings)),
		Attestations:      make([]*types.Attestation, len(blk.Block.Body.Attestations)),
		Deposits:          make([]*types.Deposit, len(blk.Block.Body.Deposits)),
		VoluntaryExits:    make([]*types.VoluntaryExit, len(blk.Block.Body.VoluntaryExits)),
		Proposer:          uint64(blk.Block.ProposerIndex),
	}

	if blk.Block.Body.SyncAggregate != nil {
		bits := blk.Block.Body.SyncAggregate.SyncCommitteeBits.Bytes()
		b.SyncAggregate = &types.SyncAggregate{
			SyncCommitteeBits:          bits,
			SyncAggregateParticipation: syncCommitteeParticipation(bits),
			SyncCommitteeSignature:     blk.Block.Body.SyncAggregate.SyncCommitteeSignature,
		}
	}

	for i, proposerSlashing := range blk.Block.Body.ProposerSlashings {
		b.ProposerSlashings[i] = &types.ProposerSlashing{
			ProposerIndex: uint64(proposerSlashing.Header_1.Header.ProposerIndex),
			Header1: &types.Block{
				Slot:       uint64(proposerSlashing.Header_1.Header.Slot),
				ParentRoot: proposerSlashing.Header_1.Header.ParentRoot,
				StateRoot:  proposerSlashing.Header_1.Header.StateRoot,
				Signature:  proposerSlashing.Header_1.Signature,
				BodyRoot:   proposerSlashing.Header_1.Header.BodyRoot,
			},
			Header2: &types.Block{
				Slot:       uint64(proposerSlashing.Header_2.Header.Slot),
				ParentRoot: proposerSlashing.Header_2.Header.ParentRoot,
				StateRoot:  proposerSlashing.Header_2.Header.StateRoot,
				Signature:  proposerSlashing.Header_2.Signature,
				BodyRoot:   proposerSlashing.Header_2.Header.BodyRoot,
			},
		}
	}

	for i, attesterSlashing := range blk.Block.Body.AttesterSlashings {
		b.AttesterSlashings[i] = &types.AttesterSlashing{
			Attestation1: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation_1.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation_1.Data.CommitteeIndex),
					BeaconBlockRoot: attesterSlashing.Attestation_1.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_1.Data.Source.Epoch),
						Root:  attesterSlashing.Attestation_1.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_1.Data.Target.Epoch),
						Root:  attesterSlashing.Attestation_1.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation_1.Signature,
				AttestingIndices: attesterSlashing.Attestation_1.AttestingIndices,
			},
			Attestation2: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            uint64(attesterSlashing.Attestation_2.Data.Slot),
					CommitteeIndex:  uint64(attesterSlashing.Attestation_2.Data.CommitteeIndex),
					BeaconBlockRoot: attesterSlashing.Attestation_2.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_2.Data.Source.Epoch),
						Root:  attesterSlashing.Attestation_2.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: uint64(attesterSlashing.Attestation_2.Data.Target.Epoch),
						Root:  attesterSlashing.Attestation_2.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation_2.Signature,
				AttestingIndices: attesterSlashing.Attestation_2.AttestingIndices,
			},
		}
	}

	for i, attestation := range blk.Block.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: attestation.AggregationBits,
			Data: &types.AttestationData{
				Slot:            uint64(attestation.Data.Slot),
				CommitteeIndex:  uint64(attestation.Data.CommitteeIndex),
				BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
				Source: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Source.Epoch),
					Root:  attestation.Data.Source.Root,
				},
				Target: &types.Checkpoint{
					Epoch: uint64(attestation.Data.Target.Epoch),
					Root:  attestation.Data.Target.Root,
				},
			},
			Signature: attestation.Signature,
		}

		aggregationBits := bitfield.Bitlist(a.AggregationBits)
		assignments, err := pc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.Config.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.Config.SlotsPerEpoch, err)
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
	for i, deposit := range blk.Block.Body.Deposits {
		b.Deposits[i] = &types.Deposit{
			Proof:                 deposit.Proof,
			PublicKey:             deposit.Data.PublicKey,
			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
			Amount:                deposit.Data.Amount,
			Signature:             deposit.Data.Signature,
		}
	}

	for i, voluntaryExit := range blk.Block.Body.VoluntaryExits {
		b.VoluntaryExits[i] = &types.VoluntaryExit{
			Epoch:          uint64(voluntaryExit.Exit.Epoch),
			ValidatorIndex: uint64(voluntaryExit.Exit.ValidatorIndex),
			Signature:      voluntaryExit.Signature,
		}
	}
	return b, nil
}

//
//func (pc *PrysmClient) parseMergeBlock(block *ethpb.BeaconBlockContainerMerge) (*types.Block, error) {
//	blk := block.GetAltairBlock()
//	if blk == nil {
//		return nil, fmt.Errorf("failed getting altair block")
//	}
//	exec := &blk.Block.Body.ExecutionPayload
//	b := &types.Block{
//		Status:       1,
//		Canonical:    block.Canonical,
//		BlockRoot:    block.BlockRoot,
//		Slot:         uint64(blk.Block.Slot),
//		ParentRoot:   blk.Block.ParentRoot,
//		StateRoot:    blk.Block.StateRoot,
//		Signature:    blk.Signature,
//		RandaoReveal: blk.Block.Body.RandaoReveal,
//		Graffiti:     blk.Block.Body.Graffiti,
//		Eth1Data: &types.Eth1Data{
//			DepositRoot:  blk.Block.Body.Eth1Data.DepositRoot,
//			DepositCount: blk.Block.Body.Eth1Data.DepositCount,
//			BlockHash:    blk.Block.Body.Eth1Data.BlockHash,
//		},
//		ProposerSlashings: make([]*types.ProposerSlashing, len(blk.Block.Body.ProposerSlashings)),
//		AttesterSlashings: make([]*types.AttesterSlashing, len(blk.Block.Body.AttesterSlashings)),
//		Attestations:      make([]*types.Attestation, len(blk.Block.Body.Attestations)),
//		Deposits:          make([]*types.Deposit, len(blk.Block.Body.Deposits)),
//		VoluntaryExits:    make([]*types.VoluntaryExit, len(blk.Block.Body.VoluntaryExits)),
//		Proposer:          uint64(blk.Block.ProposerIndex),
//		ExecutionPayload: &types.ExecutionPayload{
//			BlockHash:    exec.BlockHash,
//			ParentHash:   exec.ParentHash,
//			Coinbase:     exec.Coinbase,
//			StateRoot:    exec.StateRoot,
//			Number:       exec.Number,
//			GasLimit:     exec.GasLimit,
//			GasUsed:      exec.GasUsed,
//			Timestamp:    exec.Timestamp,
//			ReceiptRoot:  exec.ReceiptRoot,
//			LogsBloom:    exec.LogsBloom,
//			Transactions: make([]*types.Transaction, len(exec.Transactions)),
//		},
//	}
//
//	for i, rawTx := range exec.Transactions {
//		tx := &types.Transaction{Raw: rawTx}
//		var decTx gtypes.Transaction
//		if err := decTx.DecodeRLP(rlp.NewStream(bytes.NewReader(rawTx), uint64(len(rawTx)))); err != nil {
//h := decTx.Hash()
//tx.TxHash = h[:]
//tx.AccountNonce = decTx.Nonce()
//// big endian
//tx.Price = decTx.GasPrice().Bytes()
//tx.GasLimit = decTx.Gas()
//msg, err := decTx.AsMessage(pc.signer)
//if err != nil {
//tx.Sender = []byte{}
//} else {
//tx.Sender = msg.From().Bytes()
//}
//tx.Recipient = decTx.To().Bytes()
//tx.Amount = decTx.Value().Bytes()
//tx.Payload = decTx.Data()
// TODO
//if decTx.To() != nil {
//tx.Recipient = decTx.To().Bytes()
//} else {
//tx.Recipient = make([]byte, 0)
//}
//tx.Amount = decTx.Value().Bytes()
//tx.Payload = decTx.Data()
//		}
//		b.ExecutionPayload.Transactions[i] = tx
//	}
//
//	if blk.Block.Body.SyncAggregate != nil {
//		bits := blk.Block.Body.SyncAggregate.SyncCommitteeBits.Bytes()
//		b.SyncAggregate = &types.SyncAggregate{
//			SyncCommitteeBits:          bits,
//			SyncAggregateParticipation: bitlistParticipation(bits),
//			SyncCommitteeSignature:     blk.Block.Body.SyncAggregate.SyncCommitteeSignature,
//		}
//	}
//
//	for i, proposerSlashing := range blk.Block.Body.ProposerSlashings {
//		b.ProposerSlashings[i] = &types.ProposerSlashing{
//			ProposerIndex: uint64(proposerSlashing.Header_1.Header.ProposerIndex),
//			Header1: &types.Block{
//				Slot:       uint64(proposerSlashing.Header_1.Header.Slot),
//				ParentRoot: proposerSlashing.Header_1.Header.ParentRoot,
//				StateRoot:  proposerSlashing.Header_1.Header.StateRoot,
//				Signature:  proposerSlashing.Header_1.Signature,
//				BodyRoot:   proposerSlashing.Header_1.Header.BodyRoot,
//			},
//			Header2: &types.Block{
//				Slot:       uint64(proposerSlashing.Header_2.Header.Slot),
//				ParentRoot: proposerSlashing.Header_2.Header.ParentRoot,
//				StateRoot:  proposerSlashing.Header_2.Header.StateRoot,
//				Signature:  proposerSlashing.Header_2.Signature,
//				BodyRoot:   proposerSlashing.Header_2.Header.BodyRoot,
//			},
//		}
//	}
//
//	for i, attesterSlashing := range blk.Block.Body.AttesterSlashings {
//		b.AttesterSlashings[i] = &types.AttesterSlashing{
//			Attestation1: &types.IndexedAttestation{
//				Data: &types.AttestationData{
//					Slot:            uint64(attesterSlashing.Attestation_1.Data.Slot),
//					CommitteeIndex:  uint64(attesterSlashing.Attestation_1.Data.CommitteeIndex),
//					BeaconBlockRoot: attesterSlashing.Attestation_1.Data.BeaconBlockRoot,
//					Source: &types.Checkpoint{
//						Epoch: uint64(attesterSlashing.Attestation_1.Data.Source.Epoch),
//						Root:  attesterSlashing.Attestation_1.Data.Source.Root,
//					},
//					Target: &types.Checkpoint{
//						Epoch: uint64(attesterSlashing.Attestation_1.Data.Target.Epoch),
//						Root:  attesterSlashing.Attestation_1.Data.Target.Root,
//					},
//				},
//				Signature:        attesterSlashing.Attestation_1.Signature,
//				AttestingIndices: attesterSlashing.Attestation_1.AttestingIndices,
//			},
//			Attestation2: &types.IndexedAttestation{
//				Data: &types.AttestationData{
//					Slot:            uint64(attesterSlashing.Attestation_2.Data.Slot),
//					CommitteeIndex:  uint64(attesterSlashing.Attestation_2.Data.CommitteeIndex),
//					BeaconBlockRoot: attesterSlashing.Attestation_2.Data.BeaconBlockRoot,
//					Source: &types.Checkpoint{
//						Epoch: uint64(attesterSlashing.Attestation_2.Data.Source.Epoch),
//						Root:  attesterSlashing.Attestation_2.Data.Source.Root,
//					},
//					Target: &types.Checkpoint{
//						Epoch: uint64(attesterSlashing.Attestation_2.Data.Target.Epoch),
//						Root:  attesterSlashing.Attestation_2.Data.Target.Root,
//					},
//				},
//				Signature:        attesterSlashing.Attestation_2.Signature,
//				AttestingIndices: attesterSlashing.Attestation_2.AttestingIndices,
//			},
//		}
//	}
//
//	for i, attestation := range blk.Block.Body.Attestations {
//		a := &types.Attestation{
//			AggregationBits: attestation.AggregationBits,
//			Data: &types.AttestationData{
//				Slot:            uint64(attestation.Data.Slot),
//				CommitteeIndex:  uint64(attestation.Data.CommitteeIndex),
//				BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
//				Source: &types.Checkpoint{
//					Epoch: uint64(attestation.Data.Source.Epoch),
//					Root:  attestation.Data.Source.Root,
//				},
//				Target: &types.Checkpoint{
//					Epoch: uint64(attestation.Data.Target.Epoch),
//					Root:  attestation.Data.Target.Root,
//				},
//			},
//			Signature: attestation.Signature,
//		}
//
//		aggregationBits := bitfield.Bitlist(a.AggregationBits)
//		assignments, err := pc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.Config.SlotsPerEpoch)
//		if err != nil {
//			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %v", a.Data.Slot/utils.Config.Chain.Config.SlotsPerEpoch, err)
//		}
//
//		a.Attesters = make([]uint64, 0)
//		for i := uint64(0); i < aggregationBits.Len(); i++ {
//			if aggregationBits.BitAt(i) {
//				validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, a.Data.CommitteeIndex, i)]
//				if !found { // This should never happen!
//					validator = 0
//					logger.Errorf("error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v", i, b.Slot, a.Data.Slot, a.Data.CommitteeIndex, i)
//				}
//				a.Attesters = append(a.Attesters, validator)
//			}
//		}
//
//		b.Attestations[i] = a
//	}
//	for i, deposit := range blk.Block.Body.Deposits {
//		b.Deposits[i] = &types.Deposit{
//			Proof:                 deposit.Proof,
//			PublicKey:             deposit.Data.PublicKey,
//			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
//			Amount:                deposit.Data.Amount,
//			Signature:             deposit.Data.Signature,
//		}
//	}
//
//	for i, voluntaryExit := range blk.Block.Body.VoluntaryExits {
//		b.VoluntaryExits[i] = &types.VoluntaryExit{
//			Epoch:          uint64(voluntaryExit.Exit.Epoch),
//			ValidatorIndex: uint64(voluntaryExit.Exit.ValidatorIndex),
//			Signature:      voluntaryExit.Signature,
//		}
//	}
//	return b, nil
//}

// GetValidatorParticipation will get the validator participation from Prysm client
func (pc *PrysmClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	validatorParticipationRequest := &ethpb.GetValidatorParticipationRequest{QueryFilter: &ethpb.GetValidatorParticipationRequest_Epoch{Epoch: eth2types.Epoch(epoch)}}
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

func (pc *PrysmClient) GetFinalityCheckpoints(epoch uint64) (*types.FinalityCheckpoints, error) {
	// finalityResp, err := lc.get(fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", lc.endpoint, id))
	// if err != nil {
	// 	return nil, fmt.Errorf("error retrieving finality checkpoints of head: %v", err)
	// }
	return nil, fmt.Errorf("not implemented yet")
}

func (pc *PrysmClient) GetSyncCommittee(stateID string, epoch uint64) (*StandardSyncCommittee, error) {
	return nil, fmt.Errorf("not implemented")
}
