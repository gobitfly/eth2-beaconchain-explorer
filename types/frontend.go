package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"html/template"
	"math/big"
	"strings"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EventName string

const (
	ValidatorBalanceDecreasedEventName               EventName = "validator_balance_decreased"
	ValidatorMissedProposalEventName                 EventName = "validator_proposal_missed"
	ValidatorExecutedProposalEventName               EventName = "validator_proposal_submitted"
	ValidatorMissedAttestationEventName              EventName = "validator_attestation_missed"
	ValidatorGotSlashedEventName                     EventName = "validator_got_slashed"
	ValidatorDidSlashEventName                       EventName = "validator_did_slash"
	ValidatorIsOfflineEventName                      EventName = "validator_is_offline"
	ValidatorReceivedWithdrawalEventName             EventName = "validator_withdrawal"
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
	RocketpoolCollateralMinReached                   EventName = "rocketpool_colleteral_min"
	RocketpoolCollateralMaxReached                   EventName = "rocketpool_colleteral_max"
	SyncCommitteeSoon                                EventName = "validator_synccommittee_soon"
)

var MachineEvents = []EventName{
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineMemoryUsageEventName,
	MonitoringMachineSwitchedToETH2FallbackEventName,
	MonitoringMachineSwitchedToETH1FallbackEventName,
}

var UserIndexEvents = []EventName{
	EthClientUpdateEventName,
	MonitoringMachineCpuLoadEventName,
	EthClientUpdateEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineMemoryUsageEventName,
	MonitoringMachineSwitchedToETH2FallbackEventName,
	MonitoringMachineSwitchedToETH1FallbackEventName,
}

var EventLabel map[EventName]string = map[EventName]string{
	ValidatorBalanceDecreasedEventName:               "Your validator(s) balance decreased",
	ValidatorMissedProposalEventName:                 "Your validator(s) missed a proposal",
	ValidatorExecutedProposalEventName:               "Your validator(s) submitted a proposal",
	ValidatorMissedAttestationEventName:              "Your validator(s) missed an attestation",
	ValidatorGotSlashedEventName:                     "Your validator(s) got slashed",
	ValidatorDidSlashEventName:                       "Your validator(s) slashed another validator",
	ValidatorIsOfflineEventName:                      "Your validator(s) state changed",
	ValidatorReceivedDepositEventName:                "Your validator(s) received a deposit",
	ValidatorReceivedWithdrawalEventName:             "A withdrawal was initiated for your validators",
	NetworkSlashingEventName:                         "A slashing event has been registered by the network",
	NetworkValidatorActivationQueueFullEventName:     "The activation queue is full",
	NetworkValidatorActivationQueueNotFullEventName:  "The activation queue is empty",
	NetworkValidatorExitQueueFullEventName:           "The validator exit queue is full",
	NetworkValidatorExitQueueNotFullEventName:        "The validator exit queue is empty",
	NetworkLivenessIncreasedEventName:                "The network is experiencing liveness issues",
	EthClientUpdateEventName:                         "An Ethereum client has a new update available",
	MonitoringMachineOfflineEventName:                "Your machine(s) might be offline",
	MonitoringMachineDiskAlmostFullEventName:         "Your machine(s) disk space is running low",
	MonitoringMachineCpuLoadEventName:                "Your machine(s) has a high CPU load",
	MonitoringMachineMemoryUsageEventName:            "Your machine(s) has a high memory load",
	MonitoringMachineSwitchedToETH2FallbackEventName: "Your machine(s) is using its consensus client fallback",
	MonitoringMachineSwitchedToETH1FallbackEventName: "Your machine(s) is using its execution client fallback",
	TaxReportEventName:                               "You have an available tax report",
	RocketpoolCommissionThresholdEventName:           "Your configured Rocket Pool commission threshold is reached",
	RocketpoolNewClaimRoundStartedEventName:          "Your Rocket Pool claim from last round is available",
	RocketpoolCollateralMinReached:                   "You reached the Rocket Pool min RPL collateral",
	RocketpoolCollateralMaxReached:                   "You reached the Rocket Pool max RPL collateral",
	SyncCommitteeSoon:                                "Your validator(s) will soon be part of the sync committee",
}

func IsUserIndexed(event EventName) bool {
	for _, ev := range UserIndexEvents {
		if ev == event {
			return true
		}
	}
	return false
}

func IsMachineNotification(event EventName) bool {
	for _, ev := range MachineEvents {
		if ev == event {
			return true
		}
	}
	return false
}

var EventNames = []EventName{
	ValidatorBalanceDecreasedEventName,
	ValidatorExecutedProposalEventName,
	ValidatorMissedProposalEventName,
	ValidatorMissedAttestationEventName,
	ValidatorGotSlashedEventName,
	ValidatorDidSlashEventName,
	ValidatorIsOfflineEventName,
	ValidatorReceivedDepositEventName,
	ValidatorReceivedWithdrawalEventName,
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
	RocketpoolCollateralMinReached,
	RocketpoolCollateralMaxReached,
	SyncCommitteeSoon,
}

type EventNameDesc struct {
	Desc    string
	Event   EventName
	Info    template.HTML
	Warning template.HTML
}

type MachineMetricSystemUser struct {
	UserID                    uint64
	Machine                   string
	CurrentData               *MachineMetricSystem
	CurrentDataInsertTs       int64
	FiveMinuteOldData         *MachineMetricSystem
	FiveMinuteOldDataInsertTs int64
}

// this is the source of truth for the validator events that are supported by the user/notification page
var AddWatchlistEvents = []EventNameDesc{
	{
		Desc:  "Validator is Offline",
		Event: ValidatorIsOfflineEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation:<br><ul><li>Once you have been offline for 3 epochs</li><li>Every 32 Epochs (~3 hours) during your downtime</li><li>Once you are back online again</li></ul></div>" class="fas fa-question-circle"></i>`),
	},
	{
		Desc:  "Proposals missed",
		Event: ValidatorMissedProposalEventName,
	},
	{
		Desc:  "Proposals submitted",
		Event: ValidatorExecutedProposalEventName,
	},
	{
		Desc:  "Validator got slashed",
		Event: ValidatorGotSlashedEventName,
	},
	{
		Desc:  "Sync committee",
		Event: SyncCommitteeSoon,
	},
	{
		Desc:    "Attestations missed",
		Event:   ValidatorMissedAttestationEventName,
		Warning: template.HTML(`<i data-toggle="tooltip" title="Will trigger every epoch (6.4 minutes) during downtime" class="fas fa-exclamation-circle text-warning"></i>`),
	},
	{
		Desc:  "Withdrawal processed",
		Event: ValidatorReceivedWithdrawalEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation when:<br><ul><li>A partial withdrawal is processed</li><li>Your validator exits and its full balance is withdrawn</li></ul> <div>Requires that your validator has 0x01 credentials</div></div>" class="fas fa-question-circle"></i>`),
	},
}

// this is the source of truth for the network events that are supported by the user/notification page
var NetworkNotificationEvents = []EventNameDesc{
	{
		Desc:  "Network Notifications",
		Event: NetworkLivenessIncreasedEventName,
	},
	// {
	// 	Desc:  "Slashing Notifications",
	// 	Event: NetworkSlashingEventName,
	// },
}

func GetDisplayableEventName(event EventName) string {
	return cases.Title(language.English).String(strings.ReplaceAll(string(event), "_", " "))
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
	GetLatestState() string
	GetSubscriptionID() uint64
	GetEventName() EventName
	GetEpoch() uint64
	GetInfo(includeUrl bool) string
	GetTitle() string
	GetEventFilter() string
	GetEmailAttachment() *EmailAttachment
	GetUnsubscribeHash() string
	GetInfoMarkdown() string
}

// func UnMarschal

type Subscription struct {
	ID          *uint64    `db:"id,omitempty"`
	UserID      *uint64    `db:"user_id,omitempty"`
	EventName   string     `db:"event_name"`
	EventFilter string     `db:"event_filter"`
	LastSent    *time.Time `db:"last_sent_ts"`
	LastEpoch   *uint64    `db:"last_sent_epoch"`
	// Channels        pq.StringArray `db:"channels"`
	CreatedTime     time.Time      `db:"created_ts"`
	CreatedEpoch    uint64         `db:"created_epoch"`
	EventThreshold  float64        `db:"event_threshold"`
	UnsubscribeHash sql.NullString `db:"unsubscribe_hash" swaggertype:"string"`
	State           sql.NullString `db:"internal_state" swaggertype:"string"`
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
	ID               uint64    `db:"id"`
	Receipt          string    `db:"receipt"`
	Store            string    `db:"store"`
	Active           bool      `db:"active"`
	ValidateRemotely bool      `db:"validate_remotely"`
	ProductID        string    `db:"product_id"`
	UserID           uint64    `db:"user_id"`
	ExpiresAt        time.Time `db:"expires_at"`
}

type UserWithPremium struct {
	ID      uint64         `db:"id"`
	Product sql.NullString `db:"product_id"`
}

type TransitEmail struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime        `db:"delivered"`
	Channel string              `db:"channel"`
	Content TransitEmailContent `db:"content"`
}

type TransitEmailContent struct {
	Address     string            `json:"address,omitempty"`
	Subject     string            `json:"subject,omitempty"`
	Email       Email             `json:"email,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
}

func (e *TransitEmailContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitEmailContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitWebhook struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime          `db:"delivered"`
	Channel string                `db:"channel"`
	Content TransitWebhookContent `db:"content"`
}

type TransitWebhookContent struct {
	Webhook UserWebhook
	Event   WebhookEvent `json:"event"`
}

type WebhookEvent struct {
	Network     string `json:"network,omitempty"`
	Name        string `json:"event,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Epoch       uint64 `json:"epoch,omitempty"`
	Target      string `json:"target,omitempty"`
}

func (e *TransitWebhookContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitWebhookContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitDiscord struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime          `db:"delivered"`
	Channel string                `db:"channel"`
	Content TransitDiscordContent `db:"content"`
}

type TransitDiscordContent struct {
	Webhook        UserWebhook
	DiscordRequest DiscordReq `json:"discordRequest"`
}

func (e *TransitDiscordContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitDiscordContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitPush struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime       `db:"delivered"`
	Channel string             `db:"channel"`
	Content TransitPushContent `db:"content"`
}

type TransitPushContent struct {
	Messages []*messaging.Message
}

func (e *TransitPushContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitPushContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type EmailAttachment struct {
	Attachment []byte `json:"attachment"`
	Name       string `json:"name"`
}

type Email struct {
	Title                 string        `json:"title"`
	Body                  template.HTML `json:"body"`
	SubscriptionManageURL template.HTML `json:"subscriptionManageUrl"`
	UnSubURL              template.HTML `json:"unSubURL"`
}

type UserWebhook struct {
	ID          uint64         `db:"id" json:"id"`
	UserID      uint64         `db:"user_id" json:"-"`
	Url         string         `db:"url" json:"url"`
	Retries     uint64         `db:"retries" json:"retries"`
	LastSent    sql.NullTime   `db:"last_sent" json:"lastRetry"`
	Response    sql.NullString `db:"response" json:"response"`
	Request     sql.NullString `db:"request" json:"request"`
	Destination sql.NullString `db:"destination" json:"destination"`
	EventNames  pq.StringArray `db:"event_names" json:"-"`
}

type UserWebhookSubscriptions struct {
	ID             uint64 `db:"id"`
	UserID         uint64 `db:"user_id"`
	WebhookID      uint64 `db:"webhook_id"`
	SubscriptionID uint64 `db:"subscription_id"`
}

type NotificationChannel string

var NotificationChannelLabels map[NotificationChannel]template.HTML = map[NotificationChannel]template.HTML{
	EmailNotificationChannel:          "Email Notification",
	PushNotificationChannel:           "Push Notification",
	WebhookNotificationChannel:        `Webhook Notification (<a href="/user/webhooks">configure</a>)`,
	WebhookDiscordNotificationChannel: "Discord Notification",
}

const (
	EmailNotificationChannel          NotificationChannel = "email"
	PushNotificationChannel           NotificationChannel = "push"
	WebhookNotificationChannel        NotificationChannel = "webhook"
	WebhookDiscordNotificationChannel NotificationChannel = "webhook_discord"
)

var NotificationChannels = []NotificationChannel{
	EmailNotificationChannel,
	PushNotificationChannel,
	WebhookNotificationChannel,
	WebhookDiscordNotificationChannel,
}

func GetNotificationChannel(channel string) (NotificationChannel, error) {
	for _, ch := range NotificationChannels {
		if string(ch) == channel {
			return ch, nil
		}
	}
	return "", errors.Errorf("Could not convert channel from string to NotificationChannel type. %v is not a known channel type", channel)
}

type ErrorResponse struct {
	Status string // e.g. "200 OK"
	Body   string
}

func (e *ErrorResponse) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a ErrorResponse) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type EnsSearchPageData = struct {
	Error  string
	Search string
	Result *EnsDomainResponse
}

type GasNowPageData struct {
	Code int `json:"code"`
	Data struct {
		Rapid     *big.Int `json:"rapid"`
		Fast      *big.Int `json:"fast"`
		Standard  *big.Int `json:"standard"`
		Slow      *big.Int `json:"slow"`
		Timestamp int64    `json:"timestamp"`
		Price     float64  `json:"price,omitempty"`
		PriceUSD  float64  `json:"priceUSD"`
		Currency  string   `json:"currency,omitempty"`
	} `json:"data"`
}

type Eth1AddressSearchItem struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Token   string `json:"token"`
}

type RawMempoolResponse struct {
	Pending map[string]map[string]*RawMempoolTransaction `json:"pending"`
	Queued  map[string]map[string]*RawMempoolTransaction `json:"queued"`
	BaseFee map[string]map[string]*RawMempoolTransaction `json:"baseFee"`

	TxsByHash map[common.Hash]*RawMempoolTransaction
}

func (mempool RawMempoolResponse) FindTxByHash(txHashString string) *RawMempoolTransaction {
	return mempool.TxsByHash[common.HexToHash(txHashString)]
}

type RawMempoolTransaction struct {
	Hash             common.Hash     `json:"hash"`
	From             *common.Address `json:"from"`
	To               *common.Address `json:"to"`
	Value            *hexutil.Big    `json:"value"`
	Gas              *hexutil.Big    `json:"gas"`
	GasFeeCap        *hexutil.Big    `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Nonce            *hexutil.Big    `json:"nonce"`
	Input            *string         `json:"input"`
	TransactionIndex *hexutil.Big    `json:"transactionIndex"`
}

type MempoolTxPageData struct {
	RawMempoolTransaction
	TargetIsContract   bool
	IsContractCreation bool
}

type SyncCommitteesStats struct {
	ParticipatedSlots uint64 `db:"participated_sync" json:"participatedSlots"`
	MissedSlots       uint64 `db:"missed_sync" json:"missedSlots"`
	OrphanedSlots     uint64 `db:"orphaned_sync" json:"-"`
	ScheduledSlots    uint64 `json:"scheduledSlots"`
}

type SignatureType string

const (
	MethodSignature SignatureType = "method"
	EventSignature  SignatureType = "event"
)

type SignatureImportStatus struct {
	LatestTimestamp *string `json:"latestTimestamp"`
	NextPage        *string `json:"nextPage"`
	HasFinished     bool    `json:"hasFinished"`
}

type Signature struct {
	Id        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text_signature"`
	Hex       string `json:"hex_signature"`
	Bytes     string `json:"bytes_signature"`
}

type SearchValidatorsByEth1Result []struct {
	Eth1Address      string        `db:"from_address_text" json:"eth1_address"`
	ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
	Count            uint64        `db:"count" json:"-"`
}

type ValidatorStateCountRow struct {
	Name  string `db:"status"`
	Count uint64 `db:"validator_count"`
}
