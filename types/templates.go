package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/lib/pq"
)

// PageData is a struct to hold web page data
type PageData struct {
	Active                string
	HeaderAd              bool
	Meta                  *Meta
	ShowSyncingMessage    bool
	User                  *User
	Data                  interface{}
	Version               string
	ChainSlotsPerEpoch    uint64
	ChainSecondsPerSlot   uint64
	ChainGenesisTimestamp uint64
	CurrentEpoch          uint64
	CurrentSlot           uint64
	FinalizationDelay     uint64
	Mainnet               bool
	DepositContract       string
	EthPrice              float64
	Currency              string
	ExchangeRate          float64
	InfoBanner            *template.HTML
	ClientsUpdated        bool
	IsUserClientUpdated   func(uint64) bool
}

// Meta is a struct to hold metadata about the page
type Meta struct {
	Title       string
	Description string
	Path        string
	Tlabel1     string
	Tdata1      string
	Tlabel2     string
	Tdata2      string
	GATag       string
}

//LatestState is a struct to hold data for the banner
type LatestState struct {
	LastProposedSlot      uint64  `json:"lastProposedSlot"`
	CurrentSlot           uint64  `json:"currentSlot"`
	CurrentEpoch          uint64  `json:"currentEpoch"`
	CurrentFinalizedEpoch uint64  `json:"currentFinalizedEpoch"`
	FinalityDelay         uint64  `json:"finalityDelay"`
	IsSyncing             bool    `json:"syncing"`
	EthPrice              float64 `json:"ethPrice"`
}

type Stats struct {
	TopDepositors        *[]StatsTopDepositors
	InvalidDepositCount  *uint64 `db:"count"`
	UniqueValidatorCount *uint64 `db:"count"`
}

type StatsTopDepositors struct {
	Address      string `db:"from_address"`
	DepositCount uint64 `db:"count"`
}

// IndexPageData is a struct to hold info for the main web page
type IndexPageData struct {
	NetworkName               string `json:"-"`
	DepositContract           string `json:"-"`
	ShowSyncingMessage        bool
	CurrentEpoch              uint64                 `json:"current_epoch"`
	CurrentFinalizedEpoch     uint64                 `json:"current_finalized_epoch"`
	CurrentSlot               uint64                 `json:"current_slot"`
	ScheduledCount            uint8                  `json:"scheduled_count"`
	FinalityDelay             uint64                 `json:"finality_delay"`
	ActiveValidators          uint64                 `json:"active_validators"`
	EnteringValidators        uint64                 `json:"entering_validators"`
	ExitingValidators         uint64                 `json:"exiting_validators"`
	StakedEther               string                 `json:"staked_ether"`
	AverageBalance            string                 `json:"average_balance"`
	DepositedTotal            float64                `json:"deposit_total"`
	DepositThreshold          float64                `json:"deposit_threshold"`
	ValidatorsRemaining       float64                `json:"validators_remaining"`
	NetworkStartTs            int64                  `json:"network_start_ts"`
	MinGenesisTime            int64                  `json:"-"`
	Blocks                    []*IndexPageDataBlocks `json:"blocks"`
	Epochs                    []*IndexPageDataEpochs `json:"epochs"`
	StakedEtherChartData      [][]float64            `json:"staked_ether_chart_data"`
	ActiveValidatorsChartData [][]float64            `json:"active_validators_chart_data"`
	Subtitle                  template.HTML          `json:"-"`
	Genesis                   bool                   `json:"genesis"`
	GenesisPeriod             bool                   `json:"genesis_period"`
	Mainnet                   bool                   `json:"-"`
	DepositChart              *ChartsPageDataChart
	DepositDistribution       *ChartsPageDataChart
	Lang                      string
}

type IndexPageDataEpochs struct {
	Epoch                            uint64        `json:"epoch"`
	Ts                               time.Time     `json:"ts"`
	Finalized                        bool          `json:"finalized"`
	FinalizedFormatted               template.HTML `json:"finalized_formatted"`
	EligibleEther                    uint64        `json:"eligibleether"`
	EligibleEtherFormatted           template.HTML `json:"eligibleether_formatted"`
	GlobalParticipationRate          float64       `json:"globalparticipationrate"`
	GlobalParticipationRateFormatted template.HTML `json:"globalparticipationrate_formatted"`
	VotedEther                       uint64        `json:"votedether"`
	VotedEtherFormatted              template.HTML `json:"votedether_formatted"`
}

// IndexPageDataBlocks is a struct to hold detail data for the main web page
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
	Graffiti           []byte        `db:"graffiti"`
	ProposerName       string        `db:"name"`
}

// IndexPageEpochHistory is a struct to hold the epoch history for the main web page
type IndexPageEpochHistory struct {
	Epoch           uint64 `db:"epoch"`
	ValidatorsCount uint64 `db:"validatorscount"`
	EligibleEther   uint64 `db:"eligibleether"`
	Finalized       bool   `db:"finalized"`
}

// ValidatorsPageData is a struct to hold data about the validators page
type ValidatorsPageData struct {
	TotalCount           uint64
	DepositedCount       uint64
	PendingCount         uint64
	ActiveCount          uint64
	ActiveOnlineCount    uint64
	ActiveOfflineCount   uint64
	SlashingCount        uint64
	SlashingOnlineCount  uint64
	SlashingOfflineCount uint64
	ExitingCount         uint64
	ExitingOnlineCount   uint64
	ExitingOfflineCount  uint64
	ExitedCount          uint64
	UnknownCount         uint64
	Validators           []*ValidatorsPageDataValidators
}

// ValidatorsPageDataValidators is a struct to hold data about validators for the validators page
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
	LastAttestationSlot        *int64 `db:"lastattestationslot"`
	Name                       string `db:"name"`
	State                      string `db:"state"`
	MissedProposals            uint64 `db:"missedproposals"`
	ExecutedProposals          uint64 `db:"executedproposals"`
	MissedAttestations         uint64 `db:"missedattestations"`
	ExecutedAttestations       uint64 `db:"executedattestations"`
	Performance7d              int64  `db:"performance7d"`
}

// ValidatorPageData is a struct to hold data for the validators page
type ValidatorPageData struct {
	Epoch                               uint64 `db:"epoch"`
	ValidatorIndex                      uint64 `db:"validatorindex"`
	PublicKey                           []byte `db:"pubkey"`
	WithdrawableEpoch                   uint64 `db:"withdrawableepoch"`
	CurrentBalance                      uint64 `db:"balance"`
	BalanceActivation                   uint64 `db:"balanceactivation"`
	Balance7d                           uint64 `db:"balance7d"`
	Balance31d                          uint64 `db:"balance31d"`
	EffectiveBalance                    uint64 `db:"effectivebalance"`
	Slashed                             bool   `db:"slashed"`
	SlashedBy                           uint64
	SlashedAt                           uint64
	SlashedFor                          string
	ActivationEligibilityEpoch          uint64  `db:"activationeligibilityepoch"`
	ActivationEpoch                     uint64  `db:"activationepoch"`
	ExitEpoch                           uint64  `db:"exitepoch"`
	Index                               uint64  `db:"index"`
	LastAttestationSlot                 *uint64 `db:"lastattestationslot"`
	Name                                string  `db:"name"`
	WithdrawableTs                      time.Time
	ActivationEligibilityTs             time.Time
	ActivationTs                        time.Time
	ExitTs                              time.Time
	Status                              string `db:"status"`
	ProposedBlocksCount                 uint64
	AttestationsCount                   uint64
	StatusProposedCount                 uint64
	StatusMissedCount                   uint64
	DepositsCount                       uint64
	SlashingsCount                      uint64
	PendingCount                        uint64
	Income1d                            int64
	Income7d                            int64
	Income31d                           int64
	Rank7d                              int64 `db:"rank7d"`
	RankPercentage                      float64
	Apr                                 float64
	Proposals                           [][]uint64
	IncomeHistoryChartData              []*ChartDataPoint
	Deposits                            *ValidatorDeposits
	Eth1DepositAddress                  []byte
	FlashMessage                        string
	Watchlist                           []*TaggedValidators
	SubscriptionFlash                   []interface{}
	User                                *User
	AverageAttestationInclusionDistance float64
	AttestationInclusionEffectiveness   float64
	CsrfField                           template.HTML
	NetworkStats                        *IndexPageData
	EstimatedActivationTs               int64
	InclusionDelay                      int64
	CurrentAttestationStreak            uint64
	LongestAttestationStreak            uint64
}

type ValidatorStatsTablePageData struct {
	ValidatorIndex uint64
	Rows           []*ValidatorStatsTableRow
	Currency       string
}

type ValidatorStatsTableRow struct {
	ValidatorIndex         uint64
	Day                    uint64
	StartBalance           sql.NullInt64 `db:"start_balance"`
	EndBalance             sql.NullInt64 `db:"end_balance"`
	Income                 int64         `db:"-"`
	IncomeExchangeRate     float64       `db:"-"`
	IncomeExchangeCurrency string        `db:"-"`
	IncomeExchanged        float64       `db:"-"`
	MinBalance             sql.NullInt64 `db:"min_balance"`
	MaxBalance             sql.NullInt64 `db:"max_balance"`
	StartEffectiveBalance  sql.NullInt64 `db:"start_effective_balance"`
	EndEffectiveBalance    sql.NullInt64 `db:"end_effective_balance"`
	MinEffectiveBalance    sql.NullInt64 `db:"min_effective_balance"`
	MaxEffectiveBalance    sql.NullInt64 `db:"max_effective_balance"`
	MissedAttestations     sql.NullInt64 `db:"missed_attestations"`
	OrphanedAttestations   sql.NullInt64 `db:"orphaned_attestations"`
	ProposedBlocks         sql.NullInt64 `db:"proposed_blocks"`
	MissedBlocks           sql.NullInt64 `db:"missed_blocks"`
	OrphanedBlocks         sql.NullInt64 `db:"orphaned_blocks"`
	AttesterSlashings      sql.NullInt64 `db:"attester_slashings"`
	ProposerSlashings      sql.NullInt64 `db:"proposer_slashings"`
	Deposits               sql.NullInt64 `db:"deposits"`
	DepositsAmount         sql.NullInt64 `db:"deposits_amount"`
}

type ChartDataPoint struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
}

//ValidatorRank is a struct for validator rank data
type ValidatorRank struct {
	Rank int64 `db:"rank" json:"rank"`
}

// DailyProposalCount is a struct for the daily proposal count data
type DailyProposalCount struct {
	Day      int64
	Proposed uint
	Missed   uint
	Orphaned uint
}

// ValidatorBalanceHistory is a struct for the validator balance history data
type ValidatorBalanceHistory struct {
	Day              uint64 `db:"day"`
	Balance          uint64 `db:"balance"`
	EffectiveBalance uint64 `db:"effectivebalance"`
}

// ValidatorBalanceHistory is a struct for the validator income history data
type ValidatorIncomeHistory struct {
	Day          uint64 `db:"day"`
	Income       int64
	StartBalance int64 `db:"start_balance" json:"-"`
	EndBalance   int64 `db:"end_balance" json:"-"`
	Deposits     int64 `db:"deposits_amount" json:"-"`
}

type ValidatorBalanceHistoryChartData struct {
	Epoch   uint64
	Balance uint64
}

// ValidatorBalance is a struct for the validator balance data
type ValidatorBalance struct {
	Epoch            uint64 `db:"epoch"`
	Balance          uint64 `db:"balance"`
	EffectiveBalance uint64 `db:"effectivebalance"`
	Index            uint64 `db:"validatorindex"`
	PublicKey        []byte `db:"pubkey"`
}

// ValidatorPerformance is a struct for the validator performance data
type ValidatorPerformance struct {
	Rank            uint64 `db:"rank"`
	Index           uint64 `db:"validatorindex"`
	PublicKey       []byte `db:"pubkey"`
	Name            string `db:"name"`
	Balance         uint64 `db:"balance"`
	Performance1d   int64  `db:"performance1d"`
	Performance7d   int64  `db:"performance7d"`
	Performance31d  int64  `db:"performance31d"`
	Performance365d int64  `db:"performance365d"`
	Rank7d          int64  `db:"rank7d"`
	TotalCount      uint64 `db:"total_count"`
}

// ValidatorAttestation is a struct for the validators attestations data
type ValidatorAttestation struct {
	Epoch          uint64 `db:"epoch"`
	AttesterSlot   uint64 `db:"attesterslot"`
	CommitteeIndex uint64 `db:"committeeindex"`
	Status         uint64 `db:"status"`
	InclusionSlot  uint64 `db:"inclusionslot"`
	Delay          int64  `db:"delay"`
	// EarliestInclusionSlot uint64 `db:"earliestinclusionslot"`
}

// type AvgInclusionDistance struct {
// 	InclusionSlot         uint64 `db:"inclusionslot"`
// 	EarliestInclusionSlot uint64 `db:"earliestinclusionslot"`
// }

// VisPageData is a struct to hold the visualizations page data
type VisPageData struct {
	ChartData  []*VisChartData
	StartEpoch uint64
	EndEpoch   uint64
}

// VisChartData is a struct to hold the visualizations chart data
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

type GraffitiwallData struct {
	X         uint64 `db:"x" json:"x"`
	Y         uint64 `db:"y" json:"y"`
	Color     string `db:"color" json:"color"`
	Slot      uint64 `db:"slot" json:"slot"`
	Validator uint64 `db:"validator" json:"validator"`
}

// VisVotesPageData is a struct for the visualization votes page data
type VisVotesPageData struct {
	ChartData []*VotesVisChartData
}

// VotesVisChartData is a struct for the visualization chart data
type VotesVisChartData struct {
	Slot       uint64        `db:"slot" json:"slot"`
	BlockRoot  string        `db:"blockroot" json:"blockRoot"`
	ParentRoot string        `db:"parentroot" json:"parentRoot"`
	Validators pq.Int64Array `db:"validators" json:"validators"`
}

// BlockPageData is a struct block data used in the block page
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
	ProposerName           string `db:"name"`
	Eth1dataDepositroot    []byte `db:"eth1data_depositroot"`
	Eth1dataDepositcount   uint64 `db:"eth1data_depositcount"`
	Eth1dataBlockhash      []byte `db:"eth1data_blockhash"`
	ProposerSlashingsCount uint64 `db:"proposerslashingscount"`
	AttesterSlashingsCount uint64 `db:"attesterslashingscount"`
	AttestationsCount      uint64 `db:"attestationscount"`
	DepositsCount          uint64 `db:"depositscount"`
	VoluntaryExitscount    uint64 `db:"voluntaryexitscount"`
	SlashingsCount         uint64
	VotesCount             uint64
	Mainnet                bool

	Attestations      []*BlockPageAttestation // Attestations included in this block
	VoluntaryExits    []*BlockPageVoluntaryExits
	Votes             []*BlockVote // Attestations that voted for that block
	AttesterSlashings []*BlockPageAttesterSlashing
	ProposerSlashings []*BlockPageProposerSlashing
}

func (u *BlockPageData) MarshalJSON() ([]byte, error) {
	type Alias BlockPageData
	return json.Marshal(&struct {
		BlockRoot string
		Ts        int64
		*Alias
	}{
		BlockRoot: fmt.Sprintf("%x", u.BlockRoot),
		Ts:        u.Ts.Unix(),
		Alias:     (*Alias)(u),
	})
}

// BlockVote stores a vote for a given block
type BlockVote struct {
	Validator      uint64
	IncludedIn     uint64
	CommitteeIndex uint64
}

// BlockPageMinMaxSlot is a struct to hold min/max slot data
type BlockPageMinMaxSlot struct {
	MinSlot uint64
	MaxSlot uint64
}

// BlockPageAttestation is a struct to hold attestations on the block page
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

// BlockPageDeposit is a struct to hold data for deposits on the block page
type BlockPageDeposit struct {
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

// BlockPageVoluntaryExits is a struct to hold data for voluntary exits on the block page
type BlockPageVoluntaryExits struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	Signature      []byte `db:"signature"`
}

// BlockPageAttesterSlashing is a struct to hold data for attester slashings on the block page
type BlockPageAttesterSlashing struct {
	BlockSlot                   uint64        `db:"block_slot"`
	BlockIndex                  uint64        `db:"block_index"`
	Attestation1Indices         pq.Int64Array `db:"attestation1_indices"`
	Attestation1Signature       []byte        `db:"attestation1_signature"`
	Attestation1Slot            uint64        `db:"attestation1_slot"`
	Attestation1Index           uint64        `db:"attestation1_index"`
	Attestation1BeaconBlockRoot []byte        `db:"attestation1_beaconblockroot"`
	Attestation1SourceEpoch     uint64        `db:"attestation1_source_epoch"`
	Attestation1SourceRoot      []byte        `db:"attestation1_source_root"`
	Attestation1TargetEpoch     uint64        `db:"attestation1_target_epoch"`
	Attestation1TargetRoot      []byte        `db:"attestation1_target_root"`
	Attestation2Indices         pq.Int64Array `db:"attestation2_indices"`
	Attestation2Signature       []byte        `db:"attestation2_signature"`
	Attestation2Slot            uint64        `db:"attestation2_slot"`
	Attestation2Index           uint64        `db:"attestation2_index"`
	Attestation2BeaconBlockRoot []byte        `db:"attestation2_beaconblockroot"`
	Attestation2SourceEpoch     uint64        `db:"attestation2_source_epoch"`
	Attestation2SourceRoot      []byte        `db:"attestation2_source_root"`
	Attestation2TargetEpoch     uint64        `db:"attestation2_target_epoch"`
	Attestation2TargetRoot      []byte        `db:"attestation2_target_root"`
	SlashedValidators           []int64
}

// BlockPageProposerSlashing is a struct to hold data for proposer slashings on the block page
type BlockPageProposerSlashing struct {
	BlockSlot         uint64 `db:"block_slot"`
	BlockIndex        uint64 `db:"block_index"`
	BlockRoot         []byte `db:"block_root" json:"block_root"`
	ProposerIndex     uint64 `db:"proposerindex"`
	Header1Slot       uint64 `db:"header1_slot"`
	Header1ParentRoot []byte `db:"header1_parentroot"`
	Header1StateRoot  []byte `db:"header1_stateroot"`
	Header1BodyRoot   []byte `db:"header1_bodyroot"`
	Header1Signature  []byte `db:"header1_signature"`
	Header2Slot       uint64 `db:"header2_slot"`
	Header2ParentRoot []byte `db:"header2_parentroot"`
	Header2StateRoot  []byte `db:"header2_stateroot"`
	Header2BodyRoot   []byte `db:"header2_bodyroot"`
	Header2Signature  []byte `db:"header2_signature"`
}

// DataTableResponse is a struct to hold data for data table responses
type DataTableResponse struct {
	Draw            uint64          `json:"draw"`
	RecordsTotal    uint64          `json:"recordsTotal"`
	RecordsFiltered uint64          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

// EpochsPageData is a struct to hold epoch data for the epochs page
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

// EpochPageData is a struct to hold detailed epoch data for the epoch page
type EpochPageData struct {
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

	Blocks []*IndexPageDataBlocks

	Ts             time.Time
	NextEpoch      uint64
	PreviousEpoch  uint64
	ProposedCount  uint64
	MissedCount    uint64
	ScheduledCount uint64
	OrphanedCount  uint64
}

// EpochPageMinMaxSlot is a struct for the min/max epoch data
type EpochPageMinMaxSlot struct {
	MinEpoch uint64
	MaxEpoch uint64
}

// SearchAheadEpochsResult is a struct to hold the search ahead epochs results
type SearchAheadEpochsResult []struct {
	Epoch string `db:"epoch" json:"epoch,omitempty"`
}

// SearchAheadBlocksResult is a struct to hold the search ahead blocks results
type SearchAheadBlocksResult []struct {
	Slot string `db:"slot" json:"slot,omitempty"`
	Root string `db:"blockroot" json:"blockroot,omitempty"`
}

// SearchAheadGraffitiResult is a struct to hold the search ahead blocks results with a given graffiti
type SearchAheadGraffitiResult []struct {
	Graffiti string `db:"graffiti" json:"graffiti,omitempty"`
	Count    string `db:"count" json:"count,omitempty"`
}

// SearchAheadEth1Result is a struct to hold the search ahead eth1 results
type SearchAheadEth1Result []struct {
	Publickey   string `db:"publickey" json:"publickey,omitempty"`
	Eth1Address string `db:"from_address" json:"address,omitempty"`
}

// SearchAheadValidatorsResult is a struct to hold the search ahead validators results
type SearchAheadValidatorsResult []struct {
	Index  string `db:"index" json:"index,omitempty"`
	Pubkey string `db:"pubkey" json:"pubkey,omitempty"`
}

// GenericChartData is a struct to hold chart data
type GenericChartData struct {
	IsNormalChart                   bool
	ShowGapHider                    bool
	XAxisLabelsFormatter            template.JS
	TooltipFormatter                template.JS
	TooltipShared                   bool
	TooltipUseHTML                  bool
	TooltipSplit                    bool
	TooltipFollowPointer            bool
	PlotOptionsSeriesEventsClick    template.JS
	PlotOptionsPie                  template.JS
	DataLabelsEnabled               bool
	DataLabelsFormatter             template.JS
	PlotOptionsSeriesCursor         string
	Title                           string                    `json:"title"`
	Subtitle                        string                    `json:"subtitle"`
	XAxisTitle                      string                    `json:"x_axis_title"`
	YAxisTitle                      string                    `json:"y_axis_title"`
	Type                            string                    `json:"type"`
	StackingMode                    string                    `json:"stacking_mode"`
	ColumnDataGroupingApproximation string                    // "average", "averages", "open", "high", "low", "close" and "sum"
	Series                          []*GenericChartDataSeries `json:"series"`
}

// GenericChartDataSeries is a struct to hold chart series data
type GenericChartDataSeries struct {
	Name  string      `json:"name"`
	Data  interface{} `json:"data"`
	Stack string      `json:"stack,omitempty"`
	Type  string      `json:"type,omitempty"`
	Color string      `json:"color,omitempty"`
}

// ChartsPageData is an array to hold charts for the charts-page
type ChartsPageData []*ChartsPageDataChart

// ChartsPageDataChart is a struct to hold a chart for the charts-page
type ChartsPageDataChart struct {
	Order int
	Path  string
	Data  *GenericChartData
}

// DashboardData is a struct to hold data for the dashboard-page
type DashboardData struct {
	// BalanceHistory DashboardValidatorBalanceHistory `json:"balance_history"`
	// Earnings       ValidatorEarnings                `json:"earnings"`
	// Validators     [][]interface{}                  `json:"validators"`
	Csrf string `json:"csrf"`
}

// DashboardValidatorBalanceHistory is a struct to hold data for the balance-history on the dashboard-page
type DashboardValidatorBalanceHistory struct {
	Epoch            uint64  `db:"epoch"`
	Balance          uint64  `db:"balance"`
	EffectiveBalance uint64  `db:"effectivebalance"`
	ValidatorCount   float64 `db:"validatorcount"`
}

// ValidatorEarnings is a struct to hold the earnings of one or multiple validators
type ValidatorEarnings struct {
	Total         int64   `json:"total"`
	LastDay       int64   `json:"lastDay"`
	LastWeek      int64   `json:"lastWeek"`
	LastMonth     int64   `json:"lastMonth"`
	APR           float64 `json:"apr"`
	TotalDeposits int64   `json:"totalDeposits"`
}

// ValidatorAttestationSlashing is a struct to hold data of an attestation-slashing
type ValidatorAttestationSlashing struct {
	Epoch                  uint64        `db:"epoch" json:"epoch,omitempty"`
	Slot                   uint64        `db:"slot" json:"slot,omitempty"`
	Proposer               uint64        `db:"proposer" json:"proposer,omitempty"`
	Attestestation1Indices pq.Int64Array `db:"attestation1_indices" json:"attestation1_indices,omitempty"`
	Attestestation2Indices pq.Int64Array `db:"attestation2_indices" json:"attestation2_indices,omitempty"`
}

type ValidatorProposerSlashing struct {
	Epoch         uint64 `db:"epoch" json:"epoch,omitempty"`
	Slot          uint64 `db:"slot" json:"slot,omitempty"`
	Proposer      uint64 `db:"proposer" json:"proposer,omitempty"`
	ProposerIndex uint64 `db:"proposerindex" json:"proposer_index,omitempty"`
}

type ValidatorHistory struct {
	Epoch             uint64        `db:"epoch" json:"epoch,omitempty"`
	BalanceChange     sql.NullInt64 `db:"balancechange" json:"balance_change,omitempty"`
	AttesterSlot      sql.NullInt64 `db:"attestatation_attesterslot" json:"attester_slot,omitempty"`
	InclusionSlot     sql.NullInt64 `db:"attestation_inclusionslot" json:"inclusion_slot,omitempty"`
	AttestationStatus uint64        `db:"attestation_status" json:"attestation_status,omitempty"`
	ProposalStatus    sql.NullInt64 `db:"proposal_status" json:"proposal_status,omitempty"`
	ProposalSlot      sql.NullInt64 `db:"proposal_slot" json:"proposal_slot,omitempty"`
}

type ValidatorSlashing struct {
	Epoch                  uint64        `db:"epoch" json:"epoch,omitempty"`
	Slot                   uint64        `db:"slot" json:"slot,omitempty"`
	Proposer               uint64        `db:"proposer" json:"proposer,omitempty"`
	SlashedValidator       *uint64       `db:"slashedvalidator" json:"slashed_validator,omitempty"`
	Attestestation1Indices pq.Int64Array `db:"attestation1_indices" json:"attestation1_indices,omitempty"`
	Attestestation2Indices pq.Int64Array `db:"attestation2_indices" json:"attestation2_indices,omitempty"`
	Type                   string        `db:"type" json:"type"`
}

type StakingCalculatorPageData struct {
	BestValidatorBalanceHistory *[]ValidatorBalanceHistory
	WatchlistBalanceHistory     [][]interface{}
	TotalStaked                 uint64
}

type EthOneDepositsPageData struct {
	*Stats
	DepositContract string
	DepositChart    *ChartsPageDataChart
}

type EthOneDepositLeaderBoardPageData struct {
	DepositContract string
}

// EpochsPageData is a struct to hold epoch data for the epochs page
type EthOneDepositsData struct {
	TxHash                []byte    `db:"tx_hash"`
	TxInput               []byte    `db:"tx_input"`
	TxIndex               uint64    `db:"tx_index"`
	BlockNumber           uint64    `db:"block_number"`
	BlockTs               time.Time `db:"block_ts"`
	FromAddress           []byte    `db:"from_address"`
	PublicKey             []byte    `db:"publickey"`
	WithdrawalCredentials []byte    `db:"withdrawal_credentials"`
	Amount                uint64    `db:"amount"`
	Signature             []byte    `db:"signature"`
	MerkletreeIndex       []byte    `db:"merkletree_index"`
	State                 string    `db:"state"`
	ValidSignature        bool      `db:"valid_signature"`
}

type EthOneDepositLeaderboardData struct {
	FromAddress        []byte `db:"from_address"`
	Amount             uint64 `db:"amount"`
	ValidCount         uint64 `db:"validcount"`
	InvalidCount       uint64 `db:"invalidcount"`
	TotalCount         uint64 `db:"totalcount"`
	PendingCount       uint64 `db:"pendingcount"`
	SlashedCount       uint64 `db:"slashedcount"`
	ActiveCount        uint64 `db:"activecount"`
	VoluntaryExitCount uint64 `db:"voluntary_exit_count"`
}

type EthTwoDepositData struct {
	BlockSlot             uint64 `db:"block_slot"`
	BlockIndex            uint64 `db:"block_index"`
	Proof                 []byte `db:"proof"`
	Publickey             []byte `db:"publickey"`
	ValidatorIndex        uint64 `db:"validatorindex"`
	Withdrawalcredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

type ValidatorDeposits struct {
	Eth1Deposits      []Eth1Deposit
	LastEth1DepositTs int64
	Eth2Deposits      []Eth2Deposit
}

type MyCryptoSignature struct {
	Address string `json:"address"`
	Msg     string `json:"msg"`
	Sig     string `json:"sig"`
	Version string `json:"version"`
}

type User struct {
	UserID        uint64 `json:"user_id"`
	Authenticated bool   `json:"authenticated"`
}

type UserSubscription struct {
	UserID         uint64  `db:"id"`
	Email          string  `db:"email"`
	Active         *bool   `db:"active"`
	CustomerID     *string `db:"stripe_customer_id"`
	SubscriptionID *string `db:"subscription_id"`
	PriceID        *string `db:"price_id"`
	ApiKey         *string `db:"api_key"`
}

type StripeSubscription struct {
	CustomerID     *string `db:"customer_id"`
	SubscriptionID *string `db:"subscription_id"`
	PriceID        *string `db:"price_id"`
	Active         bool    `db:"active"`
}

type FilterSubscription struct {
	User     uint64
	PriceIds []string
}

type AuthData struct {
	Flashes   []interface{}
	Email     string
	State     string
	CsrfField template.HTML
}

type CsrfData struct {
	CsrfField template.HTML
}

type UserSettingsPageData struct {
	CsrfField template.HTML
	AuthData
	Subscription  UserSubscription
	PairedDevices []PairedDevice
	Sapphire      *string
	Emerald       *string
	Diamond       *string
}

type PairedDevice struct {
	ID            uint      `json:"id"`
	DeviceName    string    `json:"device_name"`
	NotifyEnabled bool      `json:"notify_enabled"`
	Active        bool      `json:"active"`
	AppName       string    `json:"app_name"`
	CreatedAt     time.Time `json:"created_ts"`
}

type UserAuthorizeConfirmPageData struct {
	AppData *OAuthAppData
	AuthData
}

type UserNotificationsPageData struct {
	Email              string   `json:"email"`
	CountWatchlist     int      `json:"countwatchlist"`
	CountSubscriptions int      `json:"countsubscriptions"`
	WatchlistIndices   []uint64 `json:"watchlistIndices"`
	DashboardLink      string   `json:"dashboardLink"`
	AuthData
	// Subscriptions []*Subscription
}

type AdvertiseWithUsPageData struct {
	FlashMessage string
	CsrfField    template.HTML
	RecaptchaKey string
}

type ApiPricing struct {
	FlashMessage string
	User         *User
	CsrfField    template.HTML
	RecaptchaKey string
	Subscription UserSubscription
	StripePK     string
	Sapphire     string
	Emerald      string
	Diamond      string
}

type StakeWithUsPageData struct {
	FlashMessage string
	RecaptchaKey string
}

type EthClients struct {
	ClientReleaseVersion string
	ClientReleaseDate    string
	NetworkShare         string
}

type EthClientServicesPageData struct {
	LastUpdate   time.Time
	Geth         EthClients
	Nethermind   EthClients
	OpenEthereum EthClients
	Besu         EthClients
	Teku         EthClients
	Prysm        EthClients
	Nimbus       EthClients
	Lighthouse   EthClients
	Banner       string
	CsrfField    template.HTML
}

type RateLimitError struct {
	TimeLeft time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit has been exceeded, %v left", e.TimeLeft)
}

type Empty struct {
}

// GoogleRecaptchaResponse ...
type GoogleRecaptchaResponse struct {
	Success            bool     `json:"success"`
	ChallengeTimestamp string   `json:"challenge_ts"`
	Hostname           string   `json:"hostname"`
	ErrorCodes         []string `json:"error-codes"`
	Score              float32  `json:"score,omitempty"`
	Action             string   `json:"action,omitempty"`
}
