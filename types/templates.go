package types

import (
	"github.com/lib/pq"
	"time"
)

type PageData struct {
	Active             string
	Meta               *Meta
	ShowSyncingMessage bool
	Data               interface{}
}

type Meta struct {
	Title       string
	Description string
	Path        string
	Tlabel1     string
	Tdata1      string
	Tlabel2     string
	Tdata2      string
}

type IndexPageData struct {
	CurrentEpoch              uint64
	CurrentFinalizedEpoch     uint64
	CurrentSlot               uint64
	FinalityDelay             uint64
	ActiveValidators          uint64
	EnteringValidators        uint64
	ExitingValidators         uint64
	StakedEther               string
	AverageBalance            string
	Blocks                    []*IndexPageDataBlocks
	StakedEtherChartData      [][]float64
	ActiveValidatorsChartData [][]float64
}

type IndexPageDataBlocks struct {
	Epoch             uint64
	Slot              uint64
	Ts                time.Time
	Proposer          uint64 `db:"proposer"`
	BlockRoot         []byte `db:"blockroot"`
	ParentRoot        []byte `db:"parentroot"`
	Attestations      uint64 `db:"attestationscount"`
	Deposits          uint64 `db:"depositscount"`
	Exits             uint64 `db:"voluntaryexitscount"`
	Proposerslashings uint64 `db:"proposerslashingscount"`
	Attesterslashings uint64 `db:"attesterslashingscount"`
	Status            uint64 `db:"status"`
}

type IndexPageEpochHistory struct {
	Epoch           uint64 `db:"epoch"`
	ValidatorsCount uint64 `db:"validatorscount"`
	EligibleEther   uint64 `db:"eligibleether"`
	Finalized       bool   `db:"finalized"`
}

type ValidatorsPageData struct {
	ActiveCount  uint64
	PendingCount uint64
	EjectedCount uint64
}

type ValidatorsPageDataValidators struct {
	Epoch                      uint64 `db:"epoch"`
	PublicKey                  []byte `db:"pubkey"`
	ValidatorIndex             uint64 `db:"validatorindex"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	CurrentBalance             uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	Status                     string
}

type ValidatorPageData struct {
	Epoch                      uint64 `db:"epoch"`
	ValidatorIndex             uint64 `db:"validatorindex"`
	PublicKey                  []byte
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	CurrentBalance             uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	Index                      uint64 `db:"index"`
	WithdrawableTs             time.Time
	ActivationEligibilityTs    time.Time
	ActivationTs               time.Time
	ExitTs                     time.Time
	CurrentBalanceFormatted    string
	EffectiveBalanceFormatted  string
	Status                     string
	ProposedBlocksCount        uint64
	AttestationsCount          uint64

	BalanceHistoryChartData          [][]float64
	EffectiveBalanceHistoryChartData [][]float64
}

type ValidatorBalanceHistory struct {
	Epoch   uint64 `db:"epoch"`
	Balance uint64 `db:"balance"`
}

type ValidatorAttestation struct {
	Epoch          uint64 `db:"epoch"`
	AttesterSlot   uint64 `db:"attesterslot"`
	CommitteeIndex uint64 `db:"committeeindex"`
	Status         uint64 `db:"status"`
}

type BlockPageData struct {
	Epoch                  uint64 `db:"epoch"`
	Slot                   uint64 `db:"slot"`
	Ts                     time.Time
	NextSlot               uint64
	PreviousSlot           uint64
	Proposer               uint64 `db:"proposer"`
	Status                 uint64 `db:"status"`
	BlockRoot              []byte `db:"blockroot"`
	ParentRoot             []byte `db:"parentroot"`
	StateRoot              []byte `db:"stateroot"`
	Signature              []byte `db:"signature"`
	RandaoReveal           []byte `db:"randaoreveal"`
	Graffiti               []byte `db:"graffiti"`
	Eth1Data_DepositRoot   []byte `db:"eth1data_depositroot"`
	Eth1Data_DepositCount  uint64 `db:"eth1data_depositcount"`
	Eth1Data_BlockHash     []byte `db:"eth1data_blockhash"`
	ProposerSlashingsCount uint64 `db:"proposerslashingscount"`
	AttesterSlashingsCount uint64 `db:"attesterslashingscount"`
	AttestationsCount      uint64 `db:"attestationscount"`
	DepositsCount          uint64 `db:"depositscount"`
	VoluntaryExitscount    uint64 `db:"voluntaryexitscount"`
	SlashingsCount         uint64

	Attestations []*BlockPageAttestation
	Deposits     []*BlockPageDeposit
}

type BlockPageMinMaxSlot struct {
	MinSlot uint64
	MaxSlot uint64
}

type BlockPageAttestation struct {
	AggregationBits []byte        `db:"aggregationbits"`
	Validators      pq.Int64Array `db:"validators"`
	Signature       []byte        `db:"signature"`
	Slot            uint64        `db:"slot"`
	CommitteeIndex  uint64        `db:"committeeindex"`
	BeaconBlockRoot []byte        `db:"beaconblockroot"`
	SourceEpoch     uint64        `db:"source_epoch"`
	SourceRoot      []byte        `db:"source_root"`
	TargetEpoch     uint64        `db:"target_epoch"`
	TargetRoot      []byte        `db:"target_root"`
}

type BlockPageDeposit struct {
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	AmountFormatted       string
	Signature             []byte `db:"signature"`
}

type DataTableResponse struct {
	Draw            uint64     `json:"draw"`
	RecordsTotal    uint64     `json:"recordsTotal"`
	RecordsFiltered uint64     `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

type EpochsPageData struct {
	Epoch                   uint64  `db:"epoch"`
	BlocksCount             uint64  `db:"blockscount"`
	ProposerSlashingsCount  uint64  `db:"proposerslashingscount"`
	AttesterSlashingsCount  uint64  `db:"attesterslashingscount"`
	AttestationsCount       uint64  `db:"attestationscount"`
	DepositsCount           uint64  `db:"depositscount"`
	VoluntaryExitsCount     uint64  `db:"voluntaryexitscount"`
	ValidatorsCount         uint64  `db:"validatorscount"`
	AverageValidatorBalance uint64  `db:"averagevalidatorbalance"`
	Finalized               bool    `db:"finalized"`
	EligibleEther           uint64  `db:"eligibleether"`
	GlobalParticipationRate float64 `db:"globalparticipationrate"`
	VotedEther              uint64  `db:"votedether"`
}
