package types

import (
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type EventName string

const (
	ValidatorBalanceDecreasedEventName               EventName = "validator_balance_decreased"
	ValidatorMissedProposalEventName                 EventName = "validator_proposal_missed"
	ValidatorExecutedProposalEventName               EventName = "validator_proposal_submitted"
	ValidatorMissedAttestationEventName              EventName = "validator_attestation_missed"
	ValidatorGotSlashedEventName                     EventName = "validator_got_slashed"
	ValidatorDidSlashEventName                       EventName = "validator_did_slash"
	ValidatorStateChangedEventName                   EventName = "validator_state_changed"
	ValidatorReceivedDepositEventName                EventName = "validator_received_deposit"
	NetworkSlashingEventName                         EventName = "network_slashing"
	NetworkValidatorActivationQueueFullEventName     EventName = "network_validator_activation_queue_full"
	NetworkValidatorActivationQueueNotFullEventName  EventName = "network_validator_activation_queue_not_full"
	NetworkValidatorExitQueueFullEventName           EventName = "network_validator_exit_queue_full"
	NetworkValidatorExitQueueNotFullEventName        EventName = "network_validator_exit_queue_not_full"
	NetworkLivenessIncreasedEventName                EventName = "network_liveness_increased"
	EthClientUpdateEventName                         EventName = "eth_client_update"
	MonitoringMachineOfflineEventName                EventName = "monitoring_machine_offline"
	MonitoringMachineDiskAlmostFullEventName         EventName = "monitoring_hdd_almostfull"
	MonitoringMachineCpuLoadEventName                EventName = "monitoring_cpu_load"
	MonitoringMachineMemoryUsageEventName            EventName = "monitoring_memory_usage"
	MonitoringMachineSwitchedToETH2FallbackEventName EventName = "monitoring_fallback_eth2inuse"
	MonitoringMachineSwitchedToETH1FallbackEventName EventName = "monitoring_fallback_eth1inuse"
	TaxReportEventName                               EventName = "user_tax_report"
	RocketpoolCommissionThresholdEventName           EventName = "rocketpool_commision_threshold"
	RocketpoolNewClaimRoundStartedEventName          EventName = "rocketpool_new_claimround"
	RocketpoolColleteralMinReached                   EventName = "rocketpool_colleteral_min"
	RocketpoolColleteralMaxReached                   EventName = "rocketpool_colleteral_max"
	SyncCommitteeSoon                                EventName = "validator_synccommittee_soon"
)

var EventNames = []EventName{
	ValidatorBalanceDecreasedEventName,
	ValidatorExecutedProposalEventName,
	ValidatorMissedProposalEventName,
	ValidatorMissedAttestationEventName,
	ValidatorGotSlashedEventName,
	ValidatorDidSlashEventName,
	ValidatorStateChangedEventName,
	ValidatorReceivedDepositEventName,
	NetworkSlashingEventName,
	NetworkValidatorActivationQueueFullEventName,
	NetworkValidatorActivationQueueNotFullEventName,
	NetworkValidatorExitQueueFullEventName,
	NetworkValidatorExitQueueNotFullEventName,
	NetworkLivenessIncreasedEventName,
	EthClientUpdateEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineSwitchedToETH2FallbackEventName,
	MonitoringMachineSwitchedToETH1FallbackEventName,
	MonitoringMachineMemoryUsageEventName,
	TaxReportEventName,
	RocketpoolCommissionThresholdEventName,
	RocketpoolNewClaimRoundStartedEventName,
	RocketpoolColleteralMinReached,
	RocketpoolColleteralMaxReached,
	SyncCommitteeSoon,
}

func GetDisplayableEventName(event EventName) string {
	return strings.Title(strings.ReplaceAll(string(event), "_", " "))
}

func EventNameFromString(event string) (EventName, error) {
	for _, en := range EventNames {
		if string(en) == event {
			return en, nil
		}
	}
	return "", errors.Errorf("Could not convert event to string. %v is not a known event type", event)
}

type Tag string

const (
	ValidatorTagsWatchlist Tag = "watchlist"
)

type Notification interface {
	GetSubscriptionID() uint64
	GetEventName() EventName
	GetEpoch() uint64
	GetInfo(includeUrl bool) string
	GetTitle() string
	GetEventFilter() string
	GetEmailAttachment() *EmailAttachment
}

type Subscription struct {
	ID             *uint64    `db:"id,omitempty"`
	UserID         *uint64    `db:"user_id,omitempty"`
	EventName      string     `db:"event_name"`
	EventFilter    string     `db:"event_filter"`
	LastSent       *time.Time `db:"last_sent_ts"`
	LastEpoch      *uint64    `db:"last_sent_epoch"`
	CreatedTime    time.Time  `db:"created_ts"`
	CreatedEpoch   uint64     `db:"created_epoch"`
	EventThreshold float64    `db:"event_threshold"`
}

type TaggedValidators struct {
	UserID             uint64 `db:"user_id"`
	Tag                string `db:"tag"`
	ValidatorPublickey []byte `db:"validator_publickey"`
	Validator          *Validator
	Events             []EventName `db:"events"`
}

type MinimalTaggedValidators struct {
	PubKey string
	Index  uint64
}

type OAuthAppData struct {
	ID          uint64 `db:"id"`
	Owner       uint64 `db:"owner_id"`
	AppName     string `db:"app_name"`
	RedirectURI string `db:"redirect_uri"`
	Active      bool   `db:"active"`
}

type OAuthCodeData struct {
	AppID  uint64 `db:"app_id"`
	UserID uint64 `db:"user_id"`
}

type MobileSettingsData struct {
	NotifyToken string `json:"notify_token"`
}

type MobileSubscription struct {
	ProductID   string                               `json:"id"`
	PriceMicros uint64                               `json:"priceMicros"`
	Currency    string                               `json:"currency"`
	Transaction MobileSubscriptionTransactionGeneric `json:"transaction"`
	Valid       bool                                 `json:"valid"`
}

type MobileSubscriptionTransactionGeneric struct {
	Type    string `json:"type"`
	Receipt string `json:"receipt"`
	ID      string `json:"id"`
}

type PremiumData struct {
	ID        uint64    `db:"id"`
	Receipt   string    `db:"receipt"`
	Store     string    `db:"store"`
	Active    bool      `db:"active"`
	ProductID string    `db:"product_id"`
	ExpiresAt time.Time `db:"expires_at"`
}

type UserWithPremium struct {
	ID      uint64         `db:"id"`
	Product sql.NullString `db:"product_id"`
}

type EmailAttachment struct {
	Attachment []byte
	Name       string
}
