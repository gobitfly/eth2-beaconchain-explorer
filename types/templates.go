package types

import (
	"html/template"
	"time"

	"github.com/lib/pq"
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
	CurrentEpoch              uint64                 `json:"current_epoch"`
	CurrentFinalizedEpoch     uint64                 `json:"current_finalized_epoch"`
	CurrentSlot               uint64                 `json:"current_slot"`
	FinalityDelay             uint64                 `json:"finality_delay"`
	ActiveValidators          uint64                 `json:"active_validators"`
	EnteringValidators        uint64                 `json:"entering_validators"`
	ExitingValidators         uint64                 `json:"exiting_validators"`
	StakedEther               string                 `json:"staked_ether"`
	AverageBalance            string                 `json:"average_balance"`
	Blocks                    []*IndexPageDataBlocks `json:"blocks"`
	StakedEtherChartData      [][]float64            `json:"staked_ether_chart_data"`
	ActiveValidatorsChartData [][]float64            `json:"active_validators_chart_data"`
}

type IndexPageDataBlocks struct {
	Epoch              uint64        `json:"epoch"`
	Slot               uint64        `json:"slot"`
	Ts                 time.Time     `json:"ts"`
	Proposer           uint64        `db:"proposer" json:"proposer"`
	ProposerFormatted  template.HTML `json:"proposer_formatted"`
	BlockRoot          []byte        `db:"blockroot" json:"block_root"`
	BlockRootFormatted string        `json:"block_root_formatted"`
	ParentRoot         []byte        `db:"parentroot" json:"parent_root"`
	Attestations       uint64        `db:"attestationscount" json:"attestations"`
	Deposits           uint64        `db:"depositscount" json:"deposits"`
	Exits              uint64        `db:"voluntaryexitscount" json:"exits"`
	Proposerslashings  uint64        `db:"proposerslashingscount" json:"proposerslashings"`
	Attesterslashings  uint64        `db:"attesterslashingscount" json:"attesterslashings"`
	Status             uint64        `db:"status" json:"status"`
	StatusFormatted    template.HTML `json:"status_formatted"`
	Votes              uint64        `db:"votes" json:"votes"`
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
	LastAttested               *uint64
	LastProposed               *uint64
}

type ValidatorPageData struct {
	Epoch                            uint64 `db:"epoch"`
	ValidatorIndex                   uint64 `db:"validatorindex"`
	PublicKey                        []byte
	WithdrawableEpoch                uint64 `db:"withdrawableepoch"`
	CurrentBalance                   uint64 `db:"balance"`
	EffectiveBalance                 uint64 `db:"effectivebalance"`
	Slashed                          bool   `db:"slashed"`
	ActivationEligibilityEpoch       uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch                  uint64 `db:"activationepoch"`
	ExitEpoch                        uint64 `db:"exitepoch"`
	Index                            uint64 `db:"index"`
	WithdrawableTs                   time.Time
	ActivationEligibilityTs          time.Time
	ActivationTs                     time.Time
	ExitTs                           time.Time
	CurrentBalanceFormatted          string
	EffectiveBalanceFormatted        string
	Status                           string
	ProposedBlocksCount              uint64
	AttestationsCount                uint64
	StatusProposedCount              uint64
	StatusMissedCount                uint64
	DailyProposalCount               []DailyProposalCount
	BalanceHistoryChartData          [][]float64
	EffectiveBalanceHistoryChartData [][]float64
}

type DailyProposalCount struct {
	Day      int64
	Proposed uint
	Missed   uint
	Orphaned uint
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

type VisPageData struct {
	ChartData  []*VisChartData
	StartEpoch uint64
	EndEpoch   uint64
}

type VisChartData struct {
	Slot       uint64 `db:"slot" json:"-"`
	BlockRoot  []byte `db:"blockroot" json:"-"`
	ParentRoot []byte `db:"parentroot" json:"-"`

	Proposer uint64 `db:"proposer" json:"proposer"`

	Number     uint64   `json:"number"`
	Timestamp  uint64   `json:"timestamp"`
	Hash       string   `json:"hash"`
	Parents    []string `json:"parents"`
	Difficulty uint64   `json:"difficulty"`
}

type VisVotesPageData struct {
	ChartData []*VotesVisChartData
}

type VotesVisChartData struct {
	Slot       uint64        `db:"slot" json:"slot"`
	BlockRoot  string        `db:"blockroot" json:"blockRoot"`
	ParentRoot string        `db:"parentroot" json:"parentRoot"`
	Validators pq.Int64Array `db:"validators" json:"validators"`
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

	Attestations []*BlockPageAttestation // Attestations included in this block
	Deposits     []*BlockPageDeposit
	Votes        []*BlockPageAttestation // Attestations that voted for that block
}

type BlockPageMinMaxSlot struct {
	MinSlot uint64
	MaxSlot uint64
}

type BlockPageAttestation struct {
	BlockSlot       uint64        `db:"block_slot"`
	BlockIndex      uint64        `db:"block_index"`
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
	Draw            uint64          `json:"draw"`
	RecordsTotal    uint64          `json:"recordsTotal"`
	RecordsFiltered uint64          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
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

type EpochPageData struct {
	Epoch                            uint64  `db:"epoch"`
	BlocksCount                      uint64  `db:"blockscount"`
	ProposerSlashingsCount           uint64  `db:"proposerslashingscount"`
	AttesterSlashingsCount           uint64  `db:"attesterslashingscount"`
	AttestationsCount                uint64  `db:"attestationscount"`
	DepositsCount                    uint64  `db:"depositscount"`
	VoluntaryExitsCount              uint64  `db:"voluntaryexitscount"`
	ValidatorsCount                  uint64  `db:"validatorscount"`
	AverageValidatorBalance          uint64  `db:"averagevalidatorbalance"`
	Finalized                        bool    `db:"finalized"`
	EligibleEther                    uint64  `db:"eligibleether"`
	GlobalParticipationRate          float64 `db:"globalparticipationrate"`
	VotedEther                       uint64  `db:"votedether"`
	VotedEtherFormatted              string
	EligibleEtherFormatted           string
	GlobalParticipationRateFormatted string

	Blocks []*IndexPageDataBlocks

	Ts            time.Time
	NextEpoch     uint64
	PreviousEpoch uint64
}

type EpochPageMinMaxSlot struct {
	MinEpoch uint64
	MaxEpoch uint64
}

type SearchAheadEpochsResult []struct {
	Epoch string `db:"epoch" json:"epoch,omitempty"`
}

type SearchAheadBlocksResult []struct {
	Slot string `db:"slot" json:"slot,omitempty"`
	Root string `db:"blockroot" json:"blockroot,omitempty"`
}

type SearchAheadValidatorsResult []struct {
	Index  string `db:"index" json:"index,omitempty"`
	Pubkey string `db:"pubkey" json:"pubkey,omitempty"`
}

type GenericChartData struct {
	Title        string                    `json:"title"`
	Subtitle     string                    `json:"subtitle"`
	XAxisTitle   string                    `json:"x_axis_title"`
	YAxisTitle   string                    `json:"y_axis_title"`
	StackingMode string                    `json:"stacking_mode"`
	Series       []*GenericChartDataSeries `json:"series"`
}

type GenericChartDataSeries struct {
	Name string      `json:"name"`
	Data [][]float64 `json:"data"`
}

type DashboardValidatorBalanceHistory struct {
	Epoch          uint64  `db:"epoch"`
	Balance        uint64  `db:"balance"`
	ValidatorCount float64 `db:"validatorcount"`
}
