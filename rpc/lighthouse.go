package rpc

import (
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

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
	client.assignmentsCache, _ = lru.New(128)

	return client, nil
}

// GetChainHead gets the chain head from Lighthouse
func (lc *LighthouseClient) GetChainHead() (*types.ChainHead, error) {
	resp, err := lc.get(fmt.Sprintf("%v%v", lc.endpoint, "/beacon/head"))

	if err != nil {
		return nil, fmt.Errorf("error retrieving chain head: %v", err)
	}

	var parsedResponse lighthouseBeaconHeadResponse
	err = json.Unmarshal(resp, &parsedResponse)

	if err != nil {
		return nil, fmt.Errorf("error parsing chain head: %v", err)
	}

	return &types.ChainHead{
		HeadSlot:                   parsedResponse.Slot,
		HeadEpoch:                  parsedResponse.Slot / utils.Config.Chain.SlotsPerEpoch,
		HeadBlockRoot:              utils.MustParseHex(parsedResponse.BlockRoot),
		FinalizedSlot:              parsedResponse.FinalizedSlot,
		FinalizedEpoch:             parsedResponse.FinalizedSlot / utils.Config.Chain.SlotsPerEpoch,
		FinalizedBlockRoot:         utils.MustParseHex(parsedResponse.FinalizedBlockRoot),
		JustifiedSlot:              parsedResponse.JustifiedSlot,
		JustifiedEpoch:             parsedResponse.JustifiedSlot / utils.Config.Chain.SlotsPerEpoch,
		JustifiedBlockRoot:         utils.MustParseHex(parsedResponse.JustifiedBlockRoot),
		PreviousJustifiedSlot:      parsedResponse.PreviousJustifiedSlot,
		PreviousJustifiedEpoch:     parsedResponse.PreviousJustifiedSlot / utils.Config.Chain.SlotsPerEpoch,
		PreviousJustifiedBlockRoot: utils.MustParseHex(parsedResponse.PreviousJustifiedBlockRoot),
	}, nil
}

// GetValidatorQueue returns an empty validator queue as the Lighthouse RPC api does not support receiving the validator queue.
func (lc *LighthouseClient) GetValidatorQueue() (*types.ValidatorQueue, map[string]uint64, error) {
	return &types.ValidatorQueue{
		ChurnLimit:           0,
		ActivationPublicKeys: [][]byte{},
		ExitPublicKeys:       [][]byte{},
	}, make(map[string]uint64), nil
}

// GetAttestationPool returns an empty Attestation as the Lighthouse RPC api does not support receiving the attestation pool.
func (lc *LighthouseClient) GetAttestationPool() ([]*types.Attestation, error) {
	return []*types.Attestation{}, nil
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

	resp, err := lc.get(fmt.Sprintf("%v%v?epoch=%v", lc.endpoint, "/validator/duties/active", epoch))

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator duties: %v", err)
	}

	var parsedResponse []*lighthouseValidatorDutiesResponse

	err = json.Unmarshal(resp, &parsedResponse)

	if err != nil {
		return nil, fmt.Errorf("error parsing validator duties: %v", err)
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	for _, assignment := range parsedResponse {
		for _, proposerSlot := range assignment.BlockProposalSlots {
			assignments.ProposerAssignments[proposerSlot] = assignment.ValidatorIndex
		}

		assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(assignment.AttestationSlot, assignment.AttestationCommitteeIndex, assignment.AttestationCommitteePosition)] = assignment.ValidatorIndex
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

	stateRoot, err := lc.get(fmt.Sprintf("%v%v?slot=%v", lc.endpoint, "/beacon/state_root", epoch*utils.Config.Chain.SlotsPerEpoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving state root for epoch %v: %v", epoch, err)
	}

	stateRootString := strings.Replace(string(stateRoot), "\"", "", -1)
	// Retrieve the validator balances for the epoch (NOTE: Currently the API call is broken and allows only to retrieve the balances for the current epoch
	data.ValidatorIndices = make(map[string]uint64)
	data.Validators = make([]*types.Validator, 0)

	var parsedResponse []*lighthouseValidatorResponse

	resp, err := lc.get(fmt.Sprintf("%v%v?state_root=%v", lc.endpoint, "/beacon/validators/all", stateRootString))
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch validators: %v", err)
	}

	err = json.Unmarshal(resp, &parsedResponse)

	if err != nil {
		return nil, fmt.Errorf("error parsing epoch validators: %v", err)
	}

	for _, validator := range parsedResponse {

		pubKey := utils.MustParseHex(validator.Pubkey)
		data.ValidatorIndices[utils.FormatPublicKey(pubKey)] = validator.ValidatorIndex

		data.Validators = append(data.Validators, &types.Validator{
			Index:                      validator.ValidatorIndex,
			PublicKey:                  pubKey,
			WithdrawalCredentials:      utils.MustParseHex(validator.Validator.WithdrawalCredentials),
			Balance:                    validator.Balance,
			EffectiveBalance:           validator.Validator.EffectiveBalance,
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.Validator.ActivationEpoch,
			ExitEpoch:                  validator.Validator.ExitEpoch,
			WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
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

			block.Proposer = data.ValidatorAssignmentes.ProposerAssignments[block.Slot]
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
		} else {
			for _, block := range data.Blocks[slot] {
				block.Proposer = proposer
			}
		}
	}

	// Unused for now
	data.BeaconCommittees = make(map[uint64][]*types.BeaconCommitteItem)

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

// GetBlocksBySlot will get the blocks by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {
	resp, err := lc.get(fmt.Sprintf("%v%v?slot=%v", lc.endpoint, "/beacon/block", slot))

	if err != nil {
		return nil, fmt.Errorf("error retrieving block data at slot %v: %v", slot, err)
	}

	var parsedResponse *lighthouseBlockResponse

	err = json.Unmarshal(resp, &parsedResponse)

	if err != nil {
		logger.Errorf("error parsing block data at slot %v: %v", slot, err)
		return []*types.Block{}, nil
	}

	block := &types.Block{
		Status:       1,
		BlockRoot:    utils.MustParseHex(parsedResponse.Root),
		Slot:         slot,
		ParentRoot:   utils.MustParseHex(parsedResponse.BeaconBlock.Message.ParentRoot),
		StateRoot:    utils.MustParseHex(parsedResponse.BeaconBlock.Message.StateRoot),
		Signature:    utils.MustParseHex(parsedResponse.BeaconBlock.Signature),
		RandaoReveal: utils.MustParseHex(parsedResponse.BeaconBlock.Message.Body.RandaoReveal),
		Graffiti:     utils.MustParseHex(parsedResponse.BeaconBlock.Message.Body.Graffiti),
		Eth1Data: &types.Eth1Data{
			DepositRoot:  utils.MustParseHex(parsedResponse.BeaconBlock.Message.Body.Eth1Data.DepositRoot),
			DepositCount: parsedResponse.BeaconBlock.Message.Body.Eth1Data.DepositCount,
			BlockHash:    utils.MustParseHex(parsedResponse.BeaconBlock.Message.Body.Eth1Data.BlockHash),
		},
		ProposerSlashings: make([]*types.ProposerSlashing, len(parsedResponse.BeaconBlock.Message.Body.ProposerSlashings)),
		AttesterSlashings: make([]*types.AttesterSlashing, len(parsedResponse.BeaconBlock.Message.Body.AttesterSlashings)),
		Attestations:      make([]*types.Attestation, len(parsedResponse.BeaconBlock.Message.Body.Attestations)),
		Deposits:          make([]*types.Deposit, len(parsedResponse.BeaconBlock.Message.Body.Deposits)),
		VoluntaryExits:    make([]*types.VoluntaryExit, len(parsedResponse.BeaconBlock.Message.Body.VoluntaryExits)),
	}

	if block.Eth1Data.DepositCount > 2147483647 { // Sometimes the lighthouse node does return bogus data for the DepositCount value
		block.Eth1Data.DepositCount = 0
	}

	for i, attestation := range parsedResponse.BeaconBlock.Message.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: utils.MustParseHex(attestation.AggregationBits),
			Attesters:       []uint64{},
			Data: &types.AttestationData{
				Slot:            attestation.Data.Slot,
				CommitteeIndex:  attestation.Data.Index,
				BeaconBlockRoot: utils.MustParseHex(attestation.Data.BeaconBlockRoot),
				Source: &types.Checkpoint{
					Epoch: attestation.Data.Source.Epoch,
					Root:  utils.MustParseHex(attestation.Data.Source.Root),
				},
				Target: &types.Checkpoint{
					Epoch: attestation.Data.Target.Epoch,
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

	for i, deposit := range parsedResponse.BeaconBlock.Message.Body.Deposits {
		d := &types.Deposit{
			Proof:                 nil,
			PublicKey:             utils.MustParseHex(deposit.Data.Pubkey),
			WithdrawalCredentials: utils.MustParseHex(deposit.Data.WithdrawalCredentials),
			Amount:                uint64(deposit.Data.Amount),
			Signature:             utils.MustParseHex(deposit.Data.Signature),
		}

		block.Deposits[i] = d
	}

	return []*types.Block{block}, nil
}

// GetValidatorParticipation will get the validator participation from the Lighthouse RPC api
func (lc *LighthouseClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	resp, err := lc.get(fmt.Sprintf("%v%v?epoch=%v", lc.endpoint, "/consensus/global_votes", epoch))

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator participation data for epoch %v: %v", epoch, err)
	}

	var parsedResponse *lighthouseValidatorParticipationResponse

	err = json.Unmarshal(resp, &parsedResponse)

	if err != nil {
		return nil, fmt.Errorf("error parsing validator participation data for epoch %v: %v", epoch, err)
	}

	return &types.ValidatorParticipation{
		Epoch:                   epoch,
		Finalized:               float32(parsedResponse.CurrentEpochTargetAttestingGwei)/float32(parsedResponse.CurrentEpochActiveGwei) > (float32(2) / float32(3)),
		GlobalParticipationRate: float32(parsedResponse.CurrentEpochAttestingGwei) / float32(parsedResponse.CurrentEpochActiveGwei),
		VotedEther:              parsedResponse.CurrentEpochActiveGwei,
		EligibleEther:           parsedResponse.CurrentEpochAttestingGwei,
	}, nil
}

func (lc *LighthouseClient) get(url string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * 60}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error-response: %s", data)
	}

	return data, err

}

type lighthouseBeaconHeadResponse struct {
	BlockRoot                  string `json:"block_root"`
	FinalizedBlockRoot         string `json:"finalized_block_root"`
	FinalizedSlot              uint64 `json:"finalized_slot"`
	JustifiedBlockRoot         string `json:"justified_block_root"`
	JustifiedSlot              uint64 `json:"justified_slot"`
	PreviousJustifiedBlockRoot string `json:"previous_justified_block_root"`
	PreviousJustifiedSlot      uint64 `json:"previous_justified_slot"`
	Slot                       uint64 `json:"slot"`
	StateRoot                  string `json:"state_root"`
}

type lighthouseValidatorDutiesResponse struct {
	AttestationCommitteeIndex    uint64   `json:"attestation_committee_index"`
	AttestationCommitteePosition uint64   `json:"attestation_committee_position"`
	AttestationSlot              uint64   `json:"attestation_slot"`
	BlockProposalSlots           []uint64 `json:"block_proposal_slots"`
	ValidatorIndex               uint64   `json:"validator_index"`
	ValidatorPubkey              string   `json:"validator_pubkey"`
}

type lighthouseValidatorParticipationResponse struct {
	CurrentEpochActiveGwei           uint64 `json:"current_epoch_active_gwei"`
	CurrentEpochAttestingGwei        uint64 `json:"current_epoch_attesting_gwei"`
	CurrentEpochTargetAttestingGwei  uint64 `json:"current_epoch_target_attesting_gwei"`
	PreviousEpochActiveGwei          uint64 `json:"previous_epoch_active_gwei"`
	PreviousEpochAttestingGwei       uint64 `json:"previous_epoch_attesting_gwei"`
	PreviousEpochHeadAttestingGwei   uint64 `json:"previous_epoch_head_attesting_gwei"`
	PreviousEpochTargetAttestingGwei uint64 `json:"previous_epoch_target_attesting_gwei"`
}

type lighthouseBlockResponse struct {
	BeaconBlock struct {
		Message struct {
			Body struct {
				Attestations []struct {
					AggregationBits string `json:"aggregation_bits"`
					Data            struct {
						BeaconBlockRoot string `json:"beacon_block_root"`
						Index           uint64 `json:"index"`
						Slot            uint64 `json:"slot"`
						Source          struct {
							Epoch uint64 `json:"epoch"`
							Root  string `json:"root"`
						} `json:"source"`
						Target struct {
							Epoch uint64 `json:"epoch"`
							Root  string `json:"root"`
						} `json:"target"`
					} `json:"data"`
					Signature string `json:"signature"`
				} `json:"attestations"`
				AttesterSlashings []interface{} `json:"attester_slashings"`
				Deposits          []struct {
					Data struct {
						Amount                int    `json:"amount"`
						Pubkey                string `json:"pubkey"`
						Signature             string `json:"signature"`
						WithdrawalCredentials string `json:"withdrawal_credentials"`
					} `json:"data"`
					Proof []string `json:"proof"`
				} `json:"deposits"`
				Eth1Data struct {
					BlockHash    string `json:"block_hash"`
					DepositCount uint64 `json:"deposit_count"`
					DepositRoot  string `json:"deposit_root"`
				} `json:"eth1_data"`
				Graffiti          string        `json:"graffiti"`
				ProposerSlashings []interface{} `json:"proposer_slashings"`
				RandaoReveal      string        `json:"randao_reveal"`
				VoluntaryExits    []interface{} `json:"voluntary_exits"`
			} `json:"body"`
			ParentRoot string `json:"parent_root"`
			Slot       uint64 `json:"slot"`
			StateRoot  string `json:"state_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"beacon_block"`
	Root string `json:"root"`
}

type lighthouseValidatorResponse struct {
	Balance   uint64 `json:"balance"`
	Pubkey    string `json:"pubkey"`
	Validator struct {
		ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
		ActivationEpoch            uint64 `json:"activation_epoch"`
		EffectiveBalance           uint64 `json:"effective_balance"`
		ExitEpoch                  uint64 `json:"exit_epoch"`
		Pubkey                     string `json:"pubkey"`
		Slashed                    bool   `json:"slashed"`
		WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
		WithdrawalCredentials      string `json:"withdrawal_credentials"`
	} `json:"validator"`
	ValidatorIndex uint64 `json:"validator_index"`
}
