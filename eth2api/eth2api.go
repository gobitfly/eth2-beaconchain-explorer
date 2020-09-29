/*

package eth2api implements the eth2 api v1. it takes what lighthouse gives
us right now so it wont work with other clients right now most likely.

- https://github.com/ethereum/eth2.0-apis
- LH implementing the v1 api https://github.com/sigp/lighthouse/pull/1569
- validator-statuses by proto https://hackmd.io/ofFJ5gOmQpu1jjHilHbdQQ
- validator-statuses by LH https://hackmd.io/bQxMDRt1RbS1TLno8K4NPg

*/

package eth2api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}
	client := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
	return client, nil
}

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"string,omitempty"`
}

func (c *Client) get(endpoint string, result interface{}, queryParams ...interface{}) error {
	url := c.baseURL + endpoint
	if len(queryParams) > 0 {
		qvs := []string{}
		for _, qp := range queryParams {
			qv, err := query.Values(qp)
			if err != nil {
				return err
			}
			qvs = append(qvs, qv.Encode())
		}
		url += "?" + strings.Join(qvs, "&")
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	r := &Response{
		Data: result,
	}
	if err := json.Unmarshal(body, r); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetGenesis() (*Genesis, error) {
	res := &Genesis{}
	err := c.get("/eth/v1/beacon/genesis", res)
	return res, err
}

type Root struct {
	Root string `json:"root"`
}

func (c *Client) GetRoot(stateID string) (string, error) {
	res := Root{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/root", stateID), &res)
	return res.Root, err
}

func (c *Client) GetFork(stateID string) (*Fork, error) {
	res := Fork{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/fork", stateID), &res)
	return &res, err
}

func (c *Client) GetFinalityCheckpoints(stateID string) (*FinalityCheckpoints, error) {
	res := FinalityCheckpoints{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/finality_checkpoints", stateID), &res)
	return &res, err
}

type GetValidatorsParams struct {
	Indices []string `url:"id"`
	States  []string `url:"status"`
}

func (c *Client) GetValidators(stateID string, params ...interface{}) (*[]Validator, error) {
	res := []Validator{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/validators", stateID), &res, params...)
	return &res, err
}

func (c *Client) GetValidator(stateID string, validatorID string) (*Validator, error) {
	res := Validator{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/validator/%v", stateID, validatorID), &res)
	return &res, err
}

type GetCommitteesParams struct {
	Index string `url:"index"`
	Slot  string `url:"slot"`
}

func (c *Client) GetCommittees(stateID string, epoch uint64, params ...interface{}) (*[]Committee, error) {
	res := []Committee{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/states/%v/committees/%v", stateID, epoch), &res, params...)
	return &res, err
}

type GetHeadersParams struct {
	Slot       string `url:"slot"`
	ParentRoot string `url:"parent_root"`
}

func (c *Client) GetHeaders(params ...interface{}) (*[]Header, error) {
	res := []Header{}
	err := c.get("/eth/v1/beacon/headers", &res, params...)
	return &res, err
}

func (c *Client) GetHeader(blockID string) (*Header, error) {
	res := Header{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/headers/%v", blockID), &res)
	return &res, err
}

func (c *Client) GetBlock(blockID string) (*SignedBlock, error) {
	res := SignedBlock{}
	err := c.get(fmt.Sprintf("/eth/v1/beacon/blocks/%v", blockID), &res)
	return &res, err
}

func (c *Client) GetBlockRoot(blockID string) (string, error) {
	res := ""
	err := c.get(fmt.Sprintf("/eth/v1/beacon/blocks/%v/root", blockID), &res)
	return res, err
}

func (c *Client) GetPoolAttesterSlashings() (*[]AttesterSlashing, error) {
	res := []AttesterSlashing{}
	err := c.get("/eth/v1/beacon/pool/attester_slashings", &res)
	return &res, err
}

func (c *Client) GetPoolProposerSlashings() (*[]ProposerSlashing, error) {
	res := []ProposerSlashing{}
	err := c.get("/eth/v1/beacon/pool/proposer_slashings", &res)
	return &res, err
}

func (c *Client) GetPoolVoluntaryExits() (*[]VoluntaryExit, error) {
	res := []VoluntaryExit{}
	err := c.get("/eth/v1/beacon/pool/voluntary_exits", &res)
	return &res, err
}

type GetAttesterDutiesParams struct {
	Indices []string `url:"index"`
}

func (c *Client) GetAttesterDuties(epoch uint64, params ...interface{}) (*[]AttesterDuty, error) {
	res := []AttesterDuty{}
	err := c.get(fmt.Sprintf("/eth/v1/validator/duties/attester/%v", epoch), &res)
	return &res, err
}

func (c *Client) GetProposerDuties(epoch uint64) (*[]ProposerDuty, error) {
	res := []ProposerDuty{}
	err := c.get(fmt.Sprintf("/eth/v1/validator/duties/proposer/%v", epoch), &res)
	return &res, err
}

type Genesis struct {
	GenesisTime           string `json:"genesis_time"`
	GenesisValidatorsRoot string `json:"genesis_validators_root"`
	GenesisForkVersion    string `json:"genesis_fork_version"`
}

type Fork struct {
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
	Epoch           uint64 `json:"epoch"`
}

type FinalityCheckpoints struct {
	PreviousJustified Checkpoint `json:"previous_justified"`
	CurrentJustified  Checkpoint `json:"current_justified"`
	Finalized         Checkpoint `json:"finalized"`
}

type Checkpoint struct {
	Epoch uint64 `json:"epoch"`
	Root  string `json:"root"`
}

// see:
// - validator-statuses by proto https://hackmd.io/ofFJ5gOmQpu1jjHilHbdQQ
// - validator-statuses by LH https://hackmd.io/bQxMDRt1RbS1TLno8K4NPg
var lighthouseValidatorStatusMap = map[string]string{
	// "Unknown":                   "unknown",
	"WaitingForEligibility":       "pending_initialized",
	"WaitingForFinality":          "pending_initialized",
	"WaitingInQueue":              "pending_queued",
	"StandbyForActive":            "active_ongoing",
	"Active":                      "active_ongoing",
	"ActiveAwaitingVoluntaryExit": "active_exiting",
	"ActiveAwaitingSlashedExit":   "active_slashed",
	"ExitedVoluntarily":           "exited_unslashed",
	"ExitedSlashed":               "exited_slashed",
	"Withdrawable":                "withdrawal_possible",
	"Withdrawn":                   "withdrawal_done",
}

type LighthouseValidatorStatus map[string]int

type ValidatorStatus string

func (vs *ValidatorStatus) UnmarshalJSON(data []byte) error {
	var err error
	var a string
	if err = json.Unmarshal(data, &a); err == nil {
		s, exists := lighthouseValidatorStatusMap[a]
		if !exists {
			return fmt.Errorf("unknown state: %v", a)
		}
		*vs = ValidatorStatus(s)
		return nil
	}
	var b LighthouseValidatorStatus
	if err = json.Unmarshal(data, &b); err == nil {
		if len(b) == 0 {
			return fmt.Errorf("invalid LighthouseValidatorStatus")
		}
		for k := range b {
			s, exists := lighthouseValidatorStatusMap[k]
			if !exists {
				return fmt.Errorf("unknown state: %v", a)
			}
			*vs = ValidatorStatus(s)
			return nil
		}
	}
	return err
}

type Validator struct {
	Index     string          `json:"index"`
	Balance   uint64          `json:"balance,string"`
	Status    ValidatorStatus `json:"status"`
	Validator struct {
		Pubkey                     string `json:"pubkey"`
		WithdrawalCredentials      string `json:"withdrawal_credentials"`
		EffectiveBalance           uint64 `json:"effective_balance"`
		Slashed                    bool   `json:"slashed"`
		ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
		ActivationEpoch            uint64 `json:"activation_epoch"`
		ExitEpoch                  uint64 `json:"exit_epoch"`
		WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
	} `json:"validator"`
}

type ValidatorIndices []uint64
func (vs *ValidatorIndices) UnmarshalJSON(data []byte) error {
	var err error
	var vsString []string
	if err = json.Unmarshal(data)
	strconv.ParseUint(data, 10, 64)
}

type Committee struct {
	Index      uint64   `json:"index,string"`
	Slot       uint64   `json:"slot"`
	Validators []string `json:"validators"`
}

func (c *Committee) UnmarshalJSON(data []byte) error {
	return nil
}

type Header struct {
	Root      string            `json:"root"`
	Canonical bool              `json:"canonical"`
	Header    SignedBlockHeader `json:"header"`
}

type SignedBlockHeader struct {
	Message struct {
		Slot          string `json:"slot"`
		ProposerIndex string `json:"proposer_index"`
		ParentRoot    string `json:"parent_root"`
		StateRoot     string `json:"state_root"`
		BodyRoot      string `json:"body_root"`
	} `json:"message"`
	Signature string `json:"signature"`
}

type BlockHeader struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

type SignedBlock struct {
	Message struct {
		Slot          uint64 `json:"slot"`
		ProposerIndex uint64 `json:"proposer_index,string"`
		ParentRoot    string `json:"parent_root"`
		StateRoot     string `json:"state_root"`
		Body          struct {
			RandaoReveal string `json:"randao_reveal"`
			Eth1Data     struct {
				DepositRoot  string `json:"deposit_root"`
				DepositCount string `json:"deposit_count"`
				BlockHash    string `json:"block_hash"`
			} `json:"eth1_data"`
			Graffiti          string                `json:"graffiti"`
			ProposerSlashings []ProposerSlashing    `json:"proposer_slashings"`
			AttesterSlashings []AttesterSlashing    `json:"attester_slashings"`
			Attestations      []Attestation         `json:"attestations"`
			Deposits          []Deposit             `json:"deposits"`
			VoluntaryExits    []SignedVoluntaryExit `json:"voluntary_exits"`
		} `json:"body"`
	} `json:"message"`
	Signature string `json:"signature"`
}

type ProposerSlashing struct {
	SignedHeader1 struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			BodyRoot      string `json:"body_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"signed_header_1"`
	SignedHeader2 struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			BodyRoot      string `json:"body_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"signed_header_2"`
}

type AttesterSlashing struct {
	Attestation1 IndexedAttestation `json:"attestation_1"`
	Attestation2 IndexedAttestation `json:"attestation_2"`
}

type Attestation struct {
	AggregationBits string          `json:"aggregation_bits"`
	Signature       string          `json:"signature"`
	Data            AttestationData `json:"data"`
}

type IndexedAttestation struct {
	AttestingIndices []string        `json:"attesting_indices"`
	Signature        string          `json:"signature"`
	Data             AttestationData `json:"data"`
}

type AttestationData struct {
	Slot            uint64 `json:"slot"`
	Index           uint64 `json:"index,string"`
	BeaconBlockRoot string `json:"beacon_block_root"`
	Source          struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"source"`
	Target struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"target"`
}

type Deposit struct {
	Proof []string `json:"proof"`
	Data  struct {
		Pubkey                string `json:"pubkey"`
		WithdrawalCredentials string `json:"withdrawal_credentials"`
		Amount                string `json:"amount"`
		Signature             string `json:"signature"`
	} `json:"data"`
}

type SignedVoluntaryExit struct {
	Message   VoluntaryExit `json:"message"`
	Signature string        `json:"signature"`
}

type VoluntaryExit struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index,string"`
}

type AttesterDuty struct {
	Pubkey                  string `json:"pubkey"`
	ValidatorIndex          uint64 `json:"validator_index,string"`
	CommitteeIndex          uint64 `json:"committee_index,string"`
	CommitteeLength         uint64 `json:"committee_length,string"`
	CommitteesAtSlot        uint64 `json:"committees_at_slot,string"`
	ValidatorCommitteeIndex uint64 `json:"validator_committee_index,string"`
	Slot                    uint64 `json:"slot"`
}

type ProposerDuty struct {
	Pubkey         string `json:"pubkey"`
	ValidatorIndex uint64 `json:"validator_index,string"`
	Slot           string `json:"slot"`
}
