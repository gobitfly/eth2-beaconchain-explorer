package services

import (
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"eth2-exporter/db"
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/mail"
	"eth2-exporter/metrics"
	"eth2-exporter/notify"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html"
	"html/template"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func notificationsSender() {
	for {
		// check if the explorer is not too far behind, if we set this value to close (10m) it could potentially never send any notifications
		// if IsSyncing() {

		if time.Now().Add(time.Minute * -20).After(utils.EpochToTime(LatestEpoch())) {
			logger.Infof("skipping notifications because the explorer is syncing, latest epoch: %v", LatestEpoch())
			time.Sleep(time.Second * 60)
			continue
		}
		start := time.Now()

		// Network DB Notifications (network related)
		notifications := collectNotifications()
		// for user, notification := range notifications {
		// 	log.Printf("Sending Notification to User: %v", user)
		// 	for event, n := range notification {
		// 		log.Printf("Notification Event: %v, Notifications: %+v", event)
		// 		for _, ev := range n {
		// 			log.Printf("event Info: %v", ev.GetInfo(true))
		// 		}
		// 	}
		// }
		sendNotifications(notifications, db.FrontendDB)

		// Network DB Notifications (user related)
		if utils.Config.Notifications.UserDBNotifications {
			userNotifications := collectUserDbNotifications()
			sendNotifications(userNotifications, db.FrontendDB)
		}

		logger.WithField("notifications", len(notifications)).WithField("duration", time.Since(start)).Info("notifications completed")
		metrics.TaskDuration.WithLabelValues("service_notifications").Observe(time.Since(start).Seconds())
		time.Sleep(time.Second * 120)
	}
}

func collectNotifications() map[uint64]map[types.EventName][]types.Notification {
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	start := time.Now()
	var err error
	// if utils.Config.Notifications.ValidatorBalanceDecreasedNotificationsEnabled {
	// 	err = collectValidatorBalanceDecreasedNotifications(notificationsByUserID)
	// 	if err != nil {
	// 		logger.Errorf("error collecting validator_balance_decreased notifications: %v", err)
	// 	}
	// }
	logger.Infof("Started collecting notifications")
	err = collectValidatorGotSlashedNotifications(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting validator_got_slashed notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_validator_got_slashed").Inc()
	}
	logger.Infof("Collecting validator got slashed notifications took: %v\n", time.Since(start))

	// executed Proposals
	err = collectBlockProposalNotifications(notificationsByUserID, 1, types.ValidatorExecutedProposalEventName)
	if err != nil {
		logger.Errorf("error collecting validator_proposal_submitted notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_executed_block_proposal").Inc()
	}
	logger.Infof("Collecting block proposal proposed notifications took: %v\n", time.Since(start))

	// Missed proposals
	err = collectBlockProposalNotifications(notificationsByUserID, 2, types.ValidatorMissedProposalEventName)
	if err != nil {
		logger.Errorf("error collecting validator_proposal_missed notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_missed_block_proposal").Inc()
	}
	logger.Infof("Collecting block proposal missed notifications took: %v\n", time.Since(start))

	// Missed attestations
	err = collectAttestationNotifications(notificationsByUserID, 0, types.ValidatorMissedAttestationEventName)
	if err != nil {
		logger.Errorf("error collecting validator_attestation_missed notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_missed_attestation").Inc()
	}
	logger.Infof("Collecting attestation notifications took: %v\n", time.Since(start))

	// Network liveness
	err = collectNetworkNotifications(notificationsByUserID, types.NetworkLivenessIncreasedEventName)
	if err != nil {
		logger.Errorf("error collecting network notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_network").Inc()
	}
	logger.Infof("Collecting collecting network notifications took: %v\n", time.Since(start))

	// Rocketpool fee comission alert
	err = collectRocketpoolComissionNotifications(notificationsByUserID, types.RocketpoolCommissionThresholdEventName)
	if err != nil {
		logger.Errorf("error collecting rocketpool commision: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_rocketpool_comission").Inc()
	}
	logger.Infof("Collecting collecting rocketpool commissions took: %v\n", time.Since(start))

	err = collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID, types.RocketpoolNewClaimRoundStartedEventName)
	if err != nil {
		logger.Errorf("error collecting new rocketpool claim round: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_rocketpool_reward_claim").Inc()
	}
	logger.Infof("Collecting collecting rocketpool claim round took: %v\n", time.Since(start))

	err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolColleteralMaxReached)
	if err != nil {
		logger.Errorf("error collecting rocketpool max colleteral: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_max_reached").Inc()
	}
	logger.Infof("Collecting collecting rocketpool max collateral took: %v\n", time.Since(start))

	err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolColleteralMinReached)
	if err != nil {
		logger.Errorf("error collecting rocketpool min colleteral: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_min_reached").Inc()
	}
	logger.Infof("Collecting collecting rocketpool min collateral took: %v\n", time.Since(start))

	err = collectSyncCommittee(notificationsByUserID, types.SyncCommitteeSoon)
	if err != nil {
		logger.Errorf("error collecting sync committee: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_sync_committee").Inc()
	}
	logger.Infof("Collecting collecting sync committee took: %v\n", time.Since(start))

	return notificationsByUserID
}

func collectUserDbNotifications() map[uint64]map[types.EventName][]types.Notification {
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	var err error

	// Monitoring (premium): machine offline
	err = collectMonitoringMachineOffline(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting Eth client offline notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_offline").Inc()
	}

	// Monitoring (premium): disk full warnings
	err = collectMonitoringMachineDiskAlmostFull(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting Eth client disk full notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_disk_almost_full").Inc()
	}

	// Monitoring (premium): cpu load
	err = collectMonitoringMachineCPULoad(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting Eth client cpu notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_cpu_load").Inc()
	}

	// Monitoring (premium): ram
	err = collectMonitoringMachineMemoryUsage(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting Eth client memory notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_memory_usage").Inc()
	}

	// New ETH clients
	err = collectEthClientNotifications(notificationsByUserID, types.EthClientUpdateEventName)
	if err != nil {
		logger.Errorf("error collecting Eth client notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_eth_client").Inc()
	}

	//Tax Report
	err = collectTaxReportNotificationNotifications(notificationsByUserID, types.TaxReportEventName)
	if err != nil {
		logger.Errorf("error collecting tax report notifications: %v", err)
		metrics.Errors.WithLabelValues("notifications_collect_tax_report").Inc()
	}

	return notificationsByUserID
}

func sendNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) {
	sendEmailNotifications(notificationsByUserID, useDB)
	sendPushNotifications(notificationsByUserID, useDB)
	// sendWebhookNotifications(notificationsByUserID)
}

func getNetwork() string {
	domainParts := strings.Split(utils.Config.Frontend.SiteDomain, ".")
	if len(domainParts) >= 3 {
		return fmt.Sprintf("%s: ", strings.Title(domainParts[0]))
	}
	return ""
}

func sendPushNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}

	tokensByUserID, err := db.GetUserPushTokenByIds(userIDs)
	if err != nil {
		logger.Errorf("error when sending push-notificaitons: could not get tokens: %v", err)
		metrics.Errors.WithLabelValues("notifications_send_push_notifications").Inc()
		return
	}

	for userID, userNotifications := range notificationsByUserID {
		userTokens, exists := tokensByUserID[userID]
		if !exists {
			continue
		}

		go func(userTokens []string, userNotifications map[types.EventName][]types.Notification) {
			var batch []*messaging.Message
			sentSubsByEpoch := map[uint64][]uint64{}

			for _, ns := range userNotifications {
				for _, n := range ns {
					for _, userToken := range userTokens {

						notification := new(messaging.Notification)
						notification.Title = fmt.Sprintf("%s%s", getNetwork(), n.GetTitle())
						notification.Body = n.GetInfo(false)
						if notification.Body == "" {
							continue
						}

						message := new(messaging.Message)
						message.Notification = notification
						message.Token = userToken

						message.APNS = new(messaging.APNSConfig)
						message.APNS.Payload = new(messaging.APNSPayload)
						message.APNS.Payload.Aps = new(messaging.Aps)
						message.APNS.Payload.Aps.Sound = "default"

						batch = append(batch, message)
					}

					e := n.GetEpoch()
					if _, exists := sentSubsByEpoch[e]; !exists {
						sentSubsByEpoch[e] = []uint64{n.GetSubscriptionID()}
					} else {
						sentSubsByEpoch[e] = append(sentSubsByEpoch[e], n.GetSubscriptionID())
					}
				}
			}

			_, err := notify.SendPushBatch(batch)
			if err != nil {
				logger.Errorf("firebase batch job failed: %v", err)
				metrics.Errors.WithLabelValues("notifications_send_push_batch").Inc()
				return
			}

			for epoch, subIDs := range sentSubsByEpoch {
				err = db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch, useDB)
				if err != nil {
					logger.Errorf("error updating sent-time of sent notifications: %v", err)
					metrics.Errors.WithLabelValues("notifications_updating_sent_time").Inc()
				}
			}
		}(userTokens, userNotifications)
	}

}

func sendEmailNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}
	emailsByUserID, err := db.GetUserEmailsByIds(userIDs)
	if err != nil {
		logger.Errorf("error when sending eamil-notificaitons: could not get emails: %v", err)
		metrics.Errors.WithLabelValues("notifications_get_user_mail_by_id").Inc()
		return
	}

	for userID, userNotifications := range notificationsByUserID {
		userEmail, exists := emailsByUserID[userID]
		if !exists {
			logger.Errorf("error when sending email-notification: could not find email for user %v", userID)
			metrics.Errors.WithLabelValues("notifications_mail_not_found").Inc()
			continue
		}
		go func(userEmail string, userNotifications map[types.EventName][]types.Notification) {
			sentSubsByEpoch := map[uint64][]uint64{}
			notification := ""
			othernotifications := ""
			i := 0
			for notificationEvent := range userNotifications {
				if i == 0 {
					notification = string(notificationEvent)
				} else if i == 1 {
					othernotifications = fmt.Sprintf(" and %s", notificationEvent)
				}
				i++
			}
			if i > 1 {
				othernotifications = fmt.Sprintf(",... and %d other notifications", i)
			}
			subject := fmt.Sprintf("%s: %s", utils.Config.Frontend.SiteDomain, notification+othernotifications)
			attachments := []types.EmailAttachment{}

			var msg types.Email

			for event, ns := range userNotifications {
				if len(msg.Body) > 0 {
					msg.Body += "<br>"
				}
				event_title := event
				if event == types.TaxReportEventName {
					event_title = "income_history"
				}
				msg.Body += template.HTML(fmt.Sprintf("%s<br>====<br><br>", types.EventLabel[event_title]))
				unsubURL := "https://" + utils.Config.Frontend.SiteDomain + "/notifications/unsubscribe"
				for i, n := range ns {
					unsubHash := n.GetUnsubscribeHash()
					if unsubHash == "" {
						id := n.GetSubscriptionID()

						tx, err := db.FrontendDB.Beginx()
						if err != nil {
							logger.WithError(err).Error("error starting transaction")
						}
						var sub types.Subscription
						err = tx.Get(&sub, `
							SELECT
								id,
								user_id,
								event_name,
								event_filter,
								last_sent_ts,
								last_sent_epoch,
								created_ts,
								created_epoch,
								event_threshold
							FROM users_subscriptions
							WHERE id = $1
						`, id)
						if err != nil {
							logger.WithError(err).Error("error getting user subscription by subscription id")
							tx.Rollback()
						}

						raw := fmt.Sprintf("%v%v%v%v", sub.ID, sub.UserID, sub.EventName, sub.CreatedTime)
						digest := sha256.Sum256([]byte(raw))

						_, err = tx.Exec("UPDATE users_subscriptions set unsubscribe_hash = $1 where id = $2", digest[:], id)
						if err != nil {
							logger.WithError(err).Error("error updating users subscriptions table with unsubscribe hash")
							tx.Rollback()
						}

						err = tx.Commit()
						if err != nil {
							logger.WithError(err).Error("error committing transaction to update users subscriptions with an unsubscribe hash")
							tx.Rollback()
						}

						unsubHash = hex.EncodeToString(digest[:])
					}
					if i == 0 {
						unsubURL += "?hash=" + html.EscapeString(unsubHash)
					} else {
						unsubURL += "&hash=" + html.EscapeString(unsubHash)
					}
					msg.UnSubURL = template.HTML(fmt.Sprintf(`<a href="%v">Unsubscribe</a>`, unsubURL))
					msg.Body += template.HTML(fmt.Sprintf("%s<br>", n.GetInfo(true)))
					e := n.GetEpoch()
					if _, exists := sentSubsByEpoch[e]; !exists {
						sentSubsByEpoch[e] = []uint64{n.GetSubscriptionID()}
					} else {
						sentSubsByEpoch[e] = append(sentSubsByEpoch[e], n.GetSubscriptionID())
					}
					if att := n.GetEmailAttachment(); att != nil {
						attachments = append(attachments, *att)
					}

				}
				if event == "validator_balance_decreased" {
					msg.Body += template.HTML("<br>You will not receive any further balance decrease mails for these validators until the balance of a validator is increasing again.<br>")
				}
			}

			// msg.Body += template.HTML(fmt.Sprintf("<br>Best regards<br>\n%s", utils.Config.Frontend.SiteDomain))

			err := mail.SendMailRateLimited(userEmail, subject, msg, attachments)
			if err != nil {
				logger.Errorf("error sending notification-email: %v", err)
				return
			}

			for epoch, subIDs := range sentSubsByEpoch {
				err = db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch, useDB)
				if err != nil {
					logger.Errorf("error updating sent-time of sent notifications: %v", err)
					metrics.Errors.WithLabelValues("notifications_updating_sent_time").Inc()
				}
			}
		}(userEmail, userNotifications)
	}
}

type validatorBalanceDecreasedNotification struct {
	ValidatorIndex     uint64
	ValidatorPublicKey string
	StartEpoch         uint64
	EndEpoch           uint64
	StartBalance       uint64
	EndBalance         uint64
	SubscriptionID     uint64
	EventFilter        string
	UnsubscribeHash    sql.NullString
}

func (n *validatorBalanceDecreasedNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorBalanceDecreasedNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorBalanceDecreasedNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorBalanceDecreasedNotification) GetEpoch() uint64 {
	return n.StartEpoch
}

func (n *validatorBalanceDecreasedNotification) GetEventName() types.EventName {
	return types.ValidatorBalanceDecreasedEventName
}

func (n *validatorBalanceDecreasedNotification) GetInfo(includeUrl bool) string {
	balance := float64(n.EndBalance) / 1e9
	diff := float64(n.StartBalance-n.EndBalance) / 1e9

	generalPart := fmt.Sprintf(`The balance of validator %[1]v decreased for 3 consecutive epochs by %.9[2]f ETH to %.9[3]f ETH from epoch %[4]v to epoch %[5]v.`, n.ValidatorIndex, diff, balance, n.StartEpoch, n.EndEpoch)
	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorBalanceDecreasedNotification) GetTitle() string {
	return "Validator Balance Decreased"
}

func (n *validatorBalanceDecreasedNotification) GetEventFilter() string {
	return n.EventFilter
}

func getUrlPart(validatorIndex uint64) string {
	return fmt.Sprintf(` For more information visit: https://%[2]s/validator/%[1]v`, validatorIndex, utils.Config.Frontend.SiteDomain)
}

// collectValidatorBalanceDecreasedNotifications finds all validators whose balance decreased for 3 consecutive epochs
// and creates notifications for all subscriptions which have not been notified about the validator since the last time its balance increased.
// It looks 10 epochs back for when the balance increased the last time, this means if the explorer is not running for 10 epochs it is possible
// that no new notification is sent even if there was a balance-increase.
func collectValidatorBalanceDecreasedNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	latestEpoch := LatestEpoch()
	if latestEpoch < 3 {
		return nil
	}
	dbResult, err := db.GetValidatorsBalanceDecrease(latestEpoch)
	if err != nil {
		return err
	}

	query := ""
	resultsLen := len(dbResult)
	for i, event := range dbResult {
		query += fmt.Sprintf(`SELECT %d as ref, id, user_id, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash from users_subscriptions where event_name = $1 AND event_filter = '%s'  AND (last_sent_epoch > $2 OR last_sent_epoch IS NULL) AND created_epoch <= $2`, i, event.Pubkey)
		if i < resultsLen-1 {
			query += " UNION "
		}
	}
	if query == "" {
		return nil
	}
	var subscribers []struct {
		Ref             uint64         `db:"ref"`
		Id              uint64         `db:"id"`
		UserId          uint64         `db:"user_id"`
		UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	}

	err = db.FrontendDB.Select(&subscribers, query, types.ValidatorBalanceDecreasedEventName, latestEpoch)
	if err != nil {
		return err
	}

	for _, sub := range subscribers {
		event := dbResult[sub.Ref]
		n := &validatorBalanceDecreasedNotification{
			SubscriptionID:  sub.Id,
			ValidatorIndex:  event.ValidatorIndex,
			StartEpoch:      latestEpoch - 3,
			EndEpoch:        latestEpoch,
			StartBalance:    event.StartBalance,
			EndBalance:      event.EndBalance,
			EventFilter:     event.Pubkey,
			UnsubscribeHash: sub.UnsubscribeHash,
		}

		if _, exists := notificationsByUserID[sub.UserId]; !exists {
			notificationsByUserID[sub.UserId] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[sub.UserId][n.GetEventName()]; !exists {
			notificationsByUserID[sub.UserId][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[sub.UserId][n.GetEventName()] = append(notificationsByUserID[sub.UserId][n.GetEventName()], n)
	}

	return nil
}

func collectBlockProposalNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, status uint64, eventName types.EventName) error {
	latestEpoch := LatestEpoch()

	type dbResult struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Epoch          uint64 `db:"epoch"`
		Status         uint64 `db:"status"`
		EventFilter    []byte `db:"pubkey"`
	}

	pubkeys, subMap, err := db.GetSubsForEventFilter(eventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missted attestations %w", err)
	}

	events := make([]dbResult, 0)
	batchSize := 5000
	dataLen := len(pubkeys)
	for i := 0; i < dataLen; i += batchSize {
		var keys [][]byte
		start := i
		end := i + batchSize

		if dataLen < end {
			end = dataLen
		}

		keys = pubkeys[start:end]

		var partial []dbResult

		err = db.DB.Select(&partial, `
				SELECT 
					v.validatorindex, 
					pa.epoch,
					pa.status,
					v.pubkey as pubkey
				FROM 
				(SELECT 
					v.validatorindex as validatorindex, 
					v.pubkey as pubkey
				FROM validators v
				WHERE pubkey = ANY($3)) v
				INNER JOIN proposal_assignments pa ON v.validatorindex = pa.validatorindex AND pa.epoch >= ($1 - 5) 
				WHERE pa.status = $2 AND pa.epoch >= ($1 - 5)`, latestEpoch, status, pq.ByteaArray(keys))
		if err != nil {
			return err
		}
		events = append(events, partial...)
	}

	for _, event := range events {
		subscribers, ok := subMap[hex.EncodeToString(event.EventFilter)]
		if !ok {
			return fmt.Errorf("error event returned that does not exist: %x", event.EventFilter)
		}
		for _, sub := range subscribers {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			if sub.LastEpoch != nil {
				lastSentEpoch := *sub.LastEpoch
				if lastSentEpoch >= event.Epoch || event.Epoch < sub.CreatedEpoch {
					continue
				}
			}
			n := &validatorProposalNotification{
				SubscriptionID: *sub.ID,
				ValidatorIndex: event.ValidatorIndex,
				Epoch:          event.Epoch,
				Status:         event.Status,
				EventName:      eventName,
				EventFilter:    hex.EncodeToString(event.EventFilter),
			}
			if _, exists := notificationsByUserID[*sub.UserID]; !exists {
				notificationsByUserID[*sub.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[*sub.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[*sub.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

type validatorProposalNotification struct {
	SubscriptionID     uint64
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	Status             uint64 // * Can be 0 = scheduled, 1 executed, 2 missed */
	EventName          types.EventName
	EventFilter        string
	UnsubscribeHash    sql.NullString
}

func (n *validatorProposalNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorProposalNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorProposalNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorProposalNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *validatorProposalNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *validatorProposalNotification) GetInfo(includeUrl bool) string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`New scheduled block proposal for Validator %[1]v.`, n.ValidatorIndex)
	case 1:
		generalPart = fmt.Sprintf(`Validator %[1]v proposed a new block.`, n.ValidatorIndex)
	case 2:
		generalPart = fmt.Sprintf(`Validator %[1]v missed a block proposal.`, n.ValidatorIndex)
	}

	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorProposalNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Block Proposal Scheduled"
	case 1:
		return "New Block Proposal"
	case 2:
		return "Block Proposal Missed"
	}
	return "-"
}

func (n *validatorProposalNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectAttestationNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, status uint64, eventName types.EventName) error {
	latestEpoch := LatestEpoch()
	latestSlot := LatestSlot()

	pubkeys, subMap, err := db.GetSubsForEventFilter(types.ValidatorMissedAttestationEventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missted attestations %w", err)
	}

	type dbResult struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Epoch          uint64 `db:"epoch"`
		Status         uint64 `db:"status"`
		Slot           uint64 `db:"attesterslot"`
		InclusionSlot  uint64 `db:"inclusionslot"`
		EventFilter    []byte `db:"pubkey"`
	}

	events := make([]dbResult, 0)
	batchSize := 5000
	dataLen := len(pubkeys)
	for i := 0; i < dataLen; i += batchSize {
		var keys [][]byte
		start := i
		end := i + batchSize

		if dataLen < end {
			end = dataLen
		}

		keys = pubkeys[start:end]

		var partial []dbResult
		err = db.DB.Select(&partial, `
		SELECT 
			v.validatorindex,
			v.pubkey,
			aa.epoch,
			aa.status,
			aa.attesterslot,
			aa.inclusionslot
		FROM
		(SELECT 
				v.validatorindex as validatorindex, 
				v.pubkey as pubkey
			FROM validators v
			WHERE pubkey = ANY($4)) v
			INNER JOIN attestation_assignments_p aa ON v.validatorindex = aa.validatorindex AND aa.week >= ($1 - 3) / 1575 AND aa.epoch >= ($1 - 3)
			WHERE status = $3
			AND aa.inclusionslot = 0 AND aa.attesterslot < ($2 - 32)
			`, latestEpoch, latestSlot, status, pq.ByteaArray(keys))
		if err != nil {
			return err
		}

		events = append(events, partial...)
	}

	for _, event := range events {
		subscribers, ok := subMap[hex.EncodeToString(event.EventFilter)]
		if !ok {
			return fmt.Errorf("error event returned that does not exist: %x", event.EventFilter)
		}
		for _, sub := range subscribers {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			if sub.LastEpoch != nil {
				lastSentEpoch := *sub.LastEpoch
				if lastSentEpoch >= event.Epoch || event.Epoch < sub.CreatedEpoch {
					continue
				}
			}
			n := &validatorAttestationNotification{
				SubscriptionID: *sub.ID,
				ValidatorIndex: event.ValidatorIndex,
				Epoch:          event.Epoch,
				Status:         event.Status,
				EventName:      eventName,
				Slot:           event.Slot,
				InclusionSlot:  event.InclusionSlot,
				EventFilter:    hex.EncodeToString(event.EventFilter),
			}
			if _, exists := notificationsByUserID[*sub.UserID]; !exists {
				notificationsByUserID[*sub.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[*sub.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[*sub.UserID][n.GetEventName()] = []types.Notification{}
			}
			isDuplicate := false
			for _, userEvent := range notificationsByUserID[*sub.UserID][n.GetEventName()] {
				if userEvent.GetSubscriptionID() == n.SubscriptionID {
					isDuplicate = true
				}
			}
			if isDuplicate {
				continue
			}
			notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

type validatorAttestationNotification struct {
	SubscriptionID     uint64
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	Status             uint64 // * Can be 0 = scheduled | missed, 1 executed
	EventName          types.EventName
	Slot               uint64
	InclusionSlot      uint64
	EventFilter        string
	UnsubscribeHash    sql.NullString
}

func (n *validatorAttestationNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorAttestationNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorAttestationNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorAttestationNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *validatorAttestationNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *validatorAttestationNotification) GetInfo(includeUrl bool) string {
	var generalPart = ""
	if includeUrl {
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator <a href="%[3]v/validator/%[1]v">%[1]v</a> missed an attestation at slot <a href="%[3]v/block/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
			//generalPart = fmt.Sprintf(`New scheduled attestation for Validator %[1]v at slot %[2]v.`, n.ValidatorIndex, n.Slot)
		case 1:
			generalPart = fmt.Sprintf(`Validator <a href="%[3]v/validator/%[1]v">%[1]v</a> submitted a successfull attestation for slot  <a href="%[3]v/block/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
		}
		// return generalPart + getUrlPart(n.ValidatorIndex)
	} else {
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator %[1]v missed an attestation at slot %[2]v.`, n.ValidatorIndex, n.Slot)
			//generalPart = fmt.Sprintf(`New scheduled attestation for Validator %[1]v at slot %[2]v.`, n.ValidatorIndex, n.Slot)
		case 1:
			generalPart = fmt.Sprintf(`Validator %[1]v submitted a successfull attestation for slot %[2]v.`, n.ValidatorIndex, n.Slot)
		}
	}
	return generalPart
}

func (n *validatorAttestationNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

func (n *validatorAttestationNotification) GetEventFilter() string {
	return n.EventFilter
}

type validatorGotSlashedNotification struct {
	SubscriptionID  uint64
	ValidatorIndex  uint64
	Epoch           uint64
	Slasher         uint64
	Reason          string
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *validatorGotSlashedNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorGotSlashedNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorGotSlashedNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorGotSlashedNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *validatorGotSlashedNotification) GetEventName() types.EventName {
	return types.ValidatorGotSlashedEventName
}

func (n *validatorGotSlashedNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`Validator %[1]v has been slashed at epoch %[2]v by validator %[3]v for %[4]s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason)
	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorGotSlashedNotification) GetTitle() string {
	return "Validator got Slashed"
}

func (n *validatorGotSlashedNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectValidatorGotSlashedNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	latestEpoch := LatestEpoch()
	if latestEpoch == 0 {
		return nil
	}

	// only consider the most recent epochs
	lookBack := int64(latestEpoch) - 50
	if lookBack < 0 {
		lookBack = 0
	}

	dbResult, err := db.GetValidatorsGotSlashed(uint64(lookBack))
	if err != nil {
		return fmt.Errorf("error getting slashed validators from database, err: %w", err)
	}
	query := ""
	resultsLen := len(dbResult)
	for i, event := range dbResult {
		query += fmt.Sprintf(`SELECT %d as ref, id, user_id, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash from users_subscriptions where event_name = $1 AND event_filter = '%x'  AND (last_sent_epoch > $2 OR last_sent_epoch IS NULL)`, i, event.SlashedValidatorPubkey)
		if i < resultsLen-1 {
			query += " UNION "
		}
	}

	if query == "" {
		return nil
	}

	var subscribers []struct {
		Ref             uint64         `db:"ref"`
		Id              uint64         `db:"id"`
		UserId          uint64         `db:"user_id"`
		UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	}

	name := string(types.ValidatorGotSlashedEventName)
	if utils.Config.Chain.Phase0.ConfigName != "" {
		name = utils.Config.Chain.Phase0.ConfigName + ":" + name
	}
	err = db.FrontendDB.Select(&subscribers, query, name, latestEpoch)
	if err != nil {
		return fmt.Errorf("error querying subscribers, err: %w", err)
	}

	for _, sub := range subscribers {
		event := dbResult[sub.Ref]
		n := &validatorGotSlashedNotification{
			SubscriptionID:  sub.Id,
			Slasher:         event.SlasherIndex,
			Epoch:           event.Epoch,
			Reason:          event.Reason,
			ValidatorIndex:  event.SlashedValidatorIndex,
			EventFilter:     hex.EncodeToString(event.SlashedValidatorPubkey),
			UnsubscribeHash: sub.UnsubscribeHash,
		}

		if _, exists := notificationsByUserID[sub.UserId]; !exists {
			notificationsByUserID[sub.UserId] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[sub.UserId][n.GetEventName()]; !exists {
			notificationsByUserID[sub.UserId][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[sub.UserId][n.GetEventName()] = append(notificationsByUserID[sub.UserId][n.GetEventName()], n)
	}

	return nil
}

type ethClientNotification struct {
	SubscriptionID  uint64
	UserID          uint64
	Epoch           uint64
	EthClient       string
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *ethClientNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *ethClientNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *ethClientNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *ethClientNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *ethClientNotification) GetEventName() types.EventName {
	return types.EthClientUpdateEventName
}

func (n *ethClientNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	if includeUrl {
		url := ""
		switch n.EthClient {
		case "Geth":
			url = "https://github.com/ethereum/go-ethereum/releases"
		case "Nethermind":
			url = "https://github.com/NethermindEth/nethermind/releases"
		case "OpenEthereum":
			url = "https://github.com/openethereum/openethereum/releases"
		case "Teku":
			url = "https://github.com/ConsenSys/teku/releases"
		case "Prysm":
			url = "https://github.com/prysmaticlabs/prysm/releases"
		case "Nimbus":
			url = "https://github.com/status-im/nimbus-eth2/releases"
		case "Lighthouse":
			url = "https://github.com/sigp/lighthouse/releases"
		default:
			url = "https://beaconcha.in/ethClients"
		}

		return generalPart + " " + url
	}
	return generalPart
}

func (n *ethClientNotification) GetTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
}

func (n *ethClientNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectEthClientNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	updatedClients := ethclients.GetUpdatedClients() //only check if there are new updates
	for _, client := range updatedClients {
		var dbResult []struct {
			SubscriptionID  uint64         `db:"id"`
			UserID          uint64         `db:"user_id"`
			Epoch           uint64         `db:"created_epoch"`
			EventFilter     string         `db:"event_filter"`
			UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
		}

		err := db.FrontendDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
			FROM users_subscriptions AS us
			WHERE 
				us.event_name=$1 
			AND 
				us.event_filter=$2 
			AND 
				((us.last_sent_ts <= NOW() - INTERVAL '2 DAY' AND TO_TIMESTAMP($3) > us.last_sent_ts) OR us.last_sent_ts IS NULL)
			`,
			eventName, strings.ToLower(client.Name), client.Date.Unix()) // was last notification sent 2 days ago for this client

		if err != nil {
			return err
		}

		for _, r := range dbResult {
			n := &ethClientNotification{
				SubscriptionID:  r.SubscriptionID,
				UserID:          r.UserID,
				Epoch:           r.Epoch,
				EventFilter:     r.EventFilter,
				EthClient:       client.Name,
				UnsubscribeHash: r.UnsubscribeHash,
			}
			if _, exists := notificationsByUserID[r.UserID]; !exists {
				notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		}
	}
	return nil
}

func collectMonitoringMachineOffline(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineOfflineEventName,
		`SELECT 
		us.user_id,
		ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash,
		max(us.id) as id,
		machine
	FROM users_subscriptions us
	JOIN (
		SELECT max(id) as id, user_id, machine, max(created_trunc) as created_trunc from stats_meta_p 
		WHERE day >= $3 
		group by user_id, machine
	) v on v.user_id = us.user_id 
	WHERE us.event_name = $1 AND us.created_epoch <= $2 
	AND us.event_filter = v.machine 
	AND (us.last_sent_epoch < ($2 - 120) OR us.last_sent_epoch IS NULL)
	AND v.created_trunc < now() - interval '4 minutes' AND v.created_trunc > now() - interval '1 hours'
	group by us.user_id, machine
	`)
}

func collectMonitoringMachineDiskAlmostFull(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineDiskAlmostFullEventName,
		`SELECT 
			us.user_id,
			ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash,
			max(us.id) as id,
			machine 
		FROM users_subscriptions us 
		INNER JOIN stats_meta_p v ON us.user_id = v.user_id
		INNER JOIN stats_system sy ON v.id = sy.meta_id
		WHERE us.event_name = $1 AND us.created_epoch <= $2 
		AND v.day >= $3 
		AND v.machine = us.event_filter 
		AND (us.last_sent_epoch < ($2 - 750) OR us.last_sent_epoch IS NULL)
		AND sy.disk_node_bytes_free::decimal / sy.disk_node_bytes_total < event_threshold
		AND v.created_trunc > NOW() - INTERVAL '1 hours' 
		group by us.user_id, machine
	`)
}

func collectMonitoringMachineCPULoad(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineCpuLoadEventName,
		`SELECT 
			max(us.id) as id,
			ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash,
			us.user_id,
			machine 
		FROM users_subscriptions us 
		INNER JOIN (
			SELECT max(id) as id, user_id, machine, max(created_trunc) as created_trunc from stats_meta_p
			where process = 'system' AND day >= $3 
			group by user_id, machine
		) v ON us.user_id = v.user_id 
		WHERE v.machine = us.event_filter 
		AND us.event_name = $1 AND us.created_epoch <= $2 
		AND (us.last_sent_epoch < ($2 - 10) OR us.last_sent_epoch IS NULL)
		AND v.created_trunc > now() - interval '45 minutes' 
		AND event_threshold < (SELECT 
			1 - (cpu_node_idle_seconds_total::decimal - lag(cpu_node_idle_seconds_total::decimal, 4, 0::decimal) OVER (PARTITION BY m.user_id, machine ORDER BY sy.id asc)) / (cpu_node_system_seconds_total::decimal - lag(cpu_node_system_seconds_total::decimal, 4, 0::decimal) OVER (PARTITION BY m.user_id, machine ORDER BY sy.id asc)) as cpu_load 
			FROM stats_system as sy 
			INNER JOIN stats_meta_p m on meta_id = m.id 
			WHERE m.id = meta_id 
			AND m.day >= $3 
			AND m.user_id = v.user_id 
			AND m.machine = us.event_filter 
			ORDER BY sy.id desc
			LIMIT 1
		) 
		group by us.user_id, machine;
	`)
}

func collectMonitoringMachineMemoryUsage(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineMemoryUsageEventName,
		`SELECT 
			max(us.id) as id,
			ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash,
			us.user_id,
			machine 
		FROM users_subscriptions us 
		INNER JOIN (
			SELECT max(id) as id, user_id, machine, max(created_trunc) as created_trunc from stats_meta_p
			where process = 'system' AND day >= $3 
			group by user_id, machine
		) v ON us.user_id = v.user_id 
		WHERE v.machine = us.event_filter 
		AND us.event_name = $1 AND us.created_epoch <= $2
		AND (us.last_sent_epoch < ($2 - 10) OR us.last_sent_epoch IS NULL)
		AND v.created_trunc > now() - interval '1 hours' 
		AND event_threshold < (SELECT avg(usage) FROM (SELECT 
		1 - ((memory_node_bytes_free + memory_node_bytes_cached + memory_node_bytes_buffers) / memory_node_bytes_total::decimal) as usage
		FROM stats_system as sy 
		INNER JOIN stats_meta_p m on meta_id = m.id 
		WHERE m.id = meta_id 
		AND m.day >= $3 
		AND m.user_id = v.user_id 
		AND m.machine = us.event_filter 
		ORDER BY sy.id desc
		LIMIT 5
		) p) 
		group by us.user_id, machine;
	`)
}

func collectMonitoringMachine(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName, query string) error {
	latestEpoch := LatestEpoch()
	if latestEpoch == 0 {
		return nil
	}

	var dbResult []struct {
		SubscriptionID  uint64         `db:"id"`
		UserID          uint64         `db:"user_id"`
		MachineName     string         `db:"machine"`
		UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	}

	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs/86400) - 1 // -1 so we have no issue on partition table change

	err := db.FrontendDB.Select(&dbResult, query, eventName, latestEpoch, day)
	if err != nil {
		return err
	}

	for _, r := range dbResult {
		n := &monitorMachineNotification{
			SubscriptionID:  r.SubscriptionID,
			MachineName:     r.MachineName,
			UserID:          r.UserID,
			EventName:       eventName,
			Epoch:           latestEpoch,
			UnsubscribeHash: r.UnsubscribeHash,
		}

		if _, exists := notificationsByUserID[r.UserID]; !exists {
			notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
			notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
	}

	return nil
}

type monitorMachineNotification struct {
	SubscriptionID  uint64
	MachineName     string
	UserID          uint64
	Epoch           uint64
	EventName       types.EventName
	UnsubscribeHash sql.NullString
}

func (n *monitorMachineNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *monitorMachineNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *monitorMachineNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *monitorMachineNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *monitorMachineNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *monitorMachineNotification) GetInfo(includeUrl bool) string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return fmt.Sprintf(`Your staking machine "%v" is running low on storage space.`, n.MachineName)
	case types.MonitoringMachineOfflineEventName:
		return fmt.Sprintf(`Your staking machine "%v" might be offline. It has not been seen for a couple minutes now.`, n.MachineName)
	case types.MonitoringMachineCpuLoadEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured CPU usage threshold.`, n.MachineName)
	case types.MonitoringMachineSwitchedToETH1FallbackEventName:
		return fmt.Sprintf(`Your staking machine "%v" has switched to your configured ETH1 fallback`, n.MachineName)
	case types.MonitoringMachineSwitchedToETH2FallbackEventName:
		return fmt.Sprintf(`Your staking machine "%v" has switched to your configured ETH2 fallback`, n.MachineName)
	case types.MonitoringMachineMemoryUsageEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured RAM threshold.`, n.MachineName)
	}
	return ""
}

func (n *monitorMachineNotification) GetTitle() string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return "Storage Warning"
	case types.MonitoringMachineOfflineEventName:
		return "Staking Machine Offline"
	case types.MonitoringMachineCpuLoadEventName:
		return "High CPU Load"
	case types.MonitoringMachineSwitchedToETH1FallbackEventName:
		return "ETH1 Fallback Active"
	case types.MonitoringMachineSwitchedToETH2FallbackEventName:
		return "ETH2 Fallback Active"
	case types.MonitoringMachineMemoryUsageEventName:
		return "Memory Warning"
	}
	return ""
}

func (n *monitorMachineNotification) GetEventFilter() string {
	return n.MachineName
}

type taxReportNotification struct {
	SubscriptionID  uint64
	UserID          uint64
	Epoch           uint64
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *taxReportNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *taxReportNotification) GetEmailAttachment() *types.EmailAttachment {
	tNow := time.Now()
	lastDay := time.Date(tNow.Year(), tNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstDay := lastDay.AddDate(0, -1, 0)

	q, err := url.ParseQuery(n.EventFilter)

	if err != nil {
		logger.Warn("Failed to parse rewards report eventfilter")
		return nil
	}

	currency := q.Get("currency")

	validators := []uint64{}
	valSlice := strings.Split(q.Get("validators"), ",")
	if len(valSlice) > 0 {
		for _, val := range valSlice {
			v, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				continue
			}
			validators = append(validators, v)
		}
	} else {
		logger.Warn("Validators Not found in rewards report eventfilter")
		return nil
	}

	pdf := GetPdfReport(validators, currency, uint64(firstDay.Unix()), uint64(lastDay.Unix()))

	return &types.EmailAttachment{Attachment: pdf, Name: fmt.Sprintf("income_history_%v_%v.pdf", firstDay.Format("20060102"), lastDay.Format("20060102"))}
}

func (n *taxReportNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *taxReportNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *taxReportNotification) GetEventName() types.EventName {
	return types.TaxReportEventName
}

func (n *taxReportNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprint(`Please find attached the income history of your selected validators.`)
	return generalPart
}

func (n *taxReportNotification) GetTitle() string {
	return fmt.Sprint("New report ready")
}

func (n *taxReportNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectTaxReportNotificationNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	tNow := time.Now()
	firstDayOfMonth := time.Date(tNow.Year(), tNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	if tNow.Year() == firstDayOfMonth.Year() && tNow.Month() == firstDayOfMonth.Month() && tNow.Day() == firstDayOfMonth.Day() { // Send the reports on the first day of the month
		var dbResult []struct {
			SubscriptionID  uint64         `db:"id"`
			UserID          uint64         `db:"user_id"`
			Epoch           uint64         `db:"created_epoch"`
			EventFilter     string         `db:"event_filter"`
			UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
		}

		name := string(eventName)
		if utils.Config.Chain.Phase0.ConfigName != "" {
			name = utils.Config.Chain.Phase0.ConfigName + ":" + name
		}

		err := db.FrontendDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
			FROM users_subscriptions AS us
			WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '2 DAY' OR us.last_sent_ts IS NULL);
			`,
			name)

		if err != nil {
			return err
		}

		for _, r := range dbResult {
			n := &taxReportNotification{
				SubscriptionID:  r.SubscriptionID,
				UserID:          r.UserID,
				Epoch:           r.Epoch,
				EventFilter:     r.EventFilter,
				UnsubscribeHash: r.UnsubscribeHash,
			}
			if _, exists := notificationsByUserID[r.UserID]; !exists {
				notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

type networkNotification struct {
	SubscriptionID  uint64
	UserID          uint64
	Epoch           uint64
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *networkNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *networkNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *networkNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *networkNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *networkNotification) GetEventName() types.EventName {
	return types.NetworkLivenessIncreasedEventName
}

func (n *networkNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func (n *networkNotification) GetTitle() string {
	return fmt.Sprint("Beaconchain Network Issues")
}

func (n *networkNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectNetworkNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	count := 0
	err := db.DB.Get(&count, `
		select count(ts) from network_liveness where (headepoch-finalizedepoch)!=2 AND ts > now() - interval '20 minutes';
	`)

	if err != nil {
		return err
	}

	if count > 0 {

		var dbResult []struct {
			SubscriptionID  uint64         `db:"id"`
			UserID          uint64         `db:"user_id"`
			Epoch           uint64         `db:"created_epoch"`
			EventFilter     string         `db:"event_filter"`
			UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
		}

		err := db.FrontendDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex')
			FROM users_subscriptions AS us
			WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '1 hour' OR us.last_sent_ts IS NULL);
			`,
			utils.GetNetwork()+":"+string(eventName))

		if err != nil {
			return err
		}

		for _, r := range dbResult {
			n := &networkNotification{
				SubscriptionID:  r.SubscriptionID,
				UserID:          r.UserID,
				Epoch:           r.Epoch,
				EventFilter:     r.EventFilter,
				UnsubscribeHash: r.UnsubscribeHash,
			}
			if _, exists := notificationsByUserID[r.UserID]; !exists {
				notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

type rocketpoolNotification struct {
	SubscriptionID  uint64
	UserID          uint64
	Epoch           uint64
	EventFilter     string
	EventName       types.EventName
	ExtraData       string
	UnsubscribeHash sql.NullString
}

func (n *rocketpoolNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *rocketpoolNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *rocketpoolNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *rocketpoolNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *rocketpoolNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *rocketpoolNotification) GetInfo(includeUrl bool) string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return fmt.Sprintf(`A new reward round has started. You can now claim your rewards from the previous round.`)
	case types.RocketpoolColleteralMaxReached:
		return `Your RPL collateral has reached your configured threshold at 150%.`
	case types.RocketpoolColleteralMinReached:
		return `Your RPL collateral has reached your configured threshold at 10%.`
	case types.SyncCommitteeSoon:
		extras := strings.Split(n.ExtraData, "|")
		if len(extras) != 3 {
			logger.Errorf("Invalid number of arguments passed to sync committee extra data. Notification will not be sent until code is corrected.")
			return ""
		}
		var inTime time.Duration
		syncStartEpoch, err := strconv.ParseUint(extras[1], 10, 64)
		if err != nil {
			inTime = time.Duration(24 * time.Hour)
		} else {
			inTime = time.Until(utils.EpochToTime(syncStartEpoch))
		}

		return fmt.Sprintf(`Your validator %v has been elected to be part of the next sync committee. The additional duties start at epoch %v, which is in %s and will last for a day until epoch %v.`, extras[0], extras[1], inTime.Round(time.Second), extras[2])
	}

	return ""
}

func (n *rocketpoolNotification) GetTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`Rocketpool Commission`)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return fmt.Sprintf(`Rocketpool Claim Available`)
	case types.RocketpoolColleteralMaxReached:
		return `Rocketpool Max Collateral`
	case types.RocketpoolColleteralMinReached:
		return `Rocketpool Min Collateral`
	case types.SyncCommitteeSoon:
		return `Sync Committee Duty`
	}
	return ""
}

func (n *rocketpoolNotification) GetEventFilter() string {
	return n.EventFilter
}

func collectRocketpoolComissionNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	fee := 0.0
	err := db.DB.Get(&fee, `
		select current_node_fee from rocketpool_network_stats order by id desc LIMIT 1;
	`)

	if err != nil {
		return err
	}

	if fee > 0 {

		var dbResult []struct {
			SubscriptionID  uint64         `db:"id"`
			UserID          uint64         `db:"user_id"`
			Epoch           uint64         `db:"created_epoch"`
			EventFilter     string         `db:"event_filter"`
			UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
		}

		err := db.FrontendDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex')
			FROM users_subscriptions AS us
			WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '8 hours' OR us.last_sent_ts IS NULL) AND (us.event_threshold <= $2 OR (us.event_threshold < 0 AND us.event_threshold * -1 >= $2));
			`,
			utils.GetNetwork()+":"+string(eventName), fee)

		if err != nil {
			return err
		}

		for _, r := range dbResult {
			n := &rocketpoolNotification{
				SubscriptionID:  r.SubscriptionID,
				UserID:          r.UserID,
				Epoch:           r.Epoch,
				EventFilter:     r.EventFilter,
				EventName:       eventName,
				ExtraData:       strconv.FormatInt(int64(fee*100), 10) + "%",
				UnsubscribeHash: r.UnsubscribeHash,
			}
			if _, exists := notificationsByUserID[r.UserID]; !exists {
				notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

func collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	var ts int64
	err := db.DB.Get(&ts, `
		select date_part('epoch', claim_interval_time_start)::int from rocketpool_network_stats order by id desc LIMIT 1;
	`)

	if err != nil {
		return err
	}

	if ts+3*60*60 > time.Now().Unix() {

		var dbResult []struct {
			SubscriptionID  uint64         `db:"id"`
			UserID          uint64         `db:"user_id"`
			Epoch           uint64         `db:"created_epoch"`
			EventFilter     string         `db:"event_filter"`
			UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
		}

		err := db.FrontendDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex')
			FROM users_subscriptions AS us
			WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '5 hours' OR us.last_sent_ts IS NULL);
			`,
			utils.GetNetwork()+":"+string(eventName))

		if err != nil {
			return err
		}

		for _, r := range dbResult {
			n := &rocketpoolNotification{
				SubscriptionID:  r.SubscriptionID,
				UserID:          r.UserID,
				Epoch:           r.Epoch,
				EventFilter:     r.EventFilter,
				EventName:       eventName,
				UnsubscribeHash: r.UnsubscribeHash,
			}
			if _, exists := notificationsByUserID[r.UserID]; !exists {
				notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
			}
			if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
				notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
			}
			notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		}
	}

	return nil
}

func collectRocketpoolRPLCollateralNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {

	pubkeys, subMap, err := db.GetSubsForEventFilter(eventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missted attestations %w", err)
	}

	type dbResult struct {
		Address     []byte
		RPLStake    BigFloat `db:"rpl_stake"`
		RPLStakeMin BigFloat `db:"min_rpl_stake"`
		RPLStakeMax BigFloat `db:"max_rpl_stake"`
	}

	events := make([]dbResult, 0)
	batchSize := 5000
	dataLen := len(pubkeys)
	for i := 0; i < dataLen; i += batchSize {
		var keys [][]byte
		start := i
		end := i + batchSize

		if dataLen < end {
			end = dataLen
		}

		keys = pubkeys[start:end]

		var partial []dbResult

		err = db.DB.Select(&partial, `
		SELECT address, rpl_stake, min_rpl_stake, max_rpl_stake                    
		FROM rocketpool_nodes WHERE address = ANY($1)`, pq.ByteaArray(keys))
		if err != nil {
			return err
		}
		events = append(events, partial...)
	}

	for _, r := range events {
		subs, ok := subMap[hex.EncodeToString(r.Address)]
		if !ok {
			continue
		}
		sub := subs[0]
		var alertConditionMet bool = false

		if sub.EventThreshold >= 0 {
			var threshold float64 = sub.EventThreshold
			if threshold == 0 {
				threshold = 1.0
			}
			if eventName == types.RocketpoolColleteralMaxReached {
				alertConditionMet = r.RPLStake.bigFloat().Cmp(r.RPLStakeMax.bigFloat().Mul(r.RPLStakeMax.bigFloat(), bigFloat(threshold))) >= 1
			} else {
				alertConditionMet = r.RPLStake.bigFloat().Cmp(r.RPLStakeMin.bigFloat().Mul(r.RPLStakeMin.bigFloat(), bigFloat(threshold))) <= -1
			}
		} else {
			if eventName == types.RocketpoolColleteralMaxReached {
				alertConditionMet = r.RPLStake.bigFloat().Cmp(r.RPLStakeMax.bigFloat().Mul(r.RPLStakeMax.bigFloat(), bigFloat(sub.EventThreshold*-1))) <= -1
			} else {
				alertConditionMet = r.RPLStake.bigFloat().Cmp(r.RPLStakeMax.bigFloat().Mul(r.RPLStakeMin.bigFloat(), bigFloat(sub.EventThreshold*-1))) >= -1
			}
		}

		if !alertConditionMet {
			continue
		}

		currentEpoch := LatestEpoch()
		if sub.LastEpoch != nil {
			lastSentEpoch := *sub.LastEpoch
			if lastSentEpoch >= currentEpoch-80 || currentEpoch < sub.CreatedEpoch {
				continue
			}
		}

		n := &rocketpoolNotification{
			SubscriptionID:  *sub.ID,
			UserID:          *sub.UserID,
			Epoch:           currentEpoch,
			EventFilter:     sub.EventFilter,
			EventName:       eventName,
			UnsubscribeHash: sub.UnsubscribeHash,
		}
		if _, exists := notificationsByUserID[*sub.UserID]; !exists {
			notificationsByUserID[*sub.UserID] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[*sub.UserID][n.GetEventName()]; !exists {
			notificationsByUserID[*sub.UserID][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], n)
	}

	return nil
}

type BigFloat big.Float

func (b *BigFloat) Value() (driver.Value, error) {
	if b != nil {
		return (*big.Float)(b).String(), nil
	}
	return nil, nil
}

func (b *BigFloat) Scan(value interface{}) error {
	if value == nil {
		return errors.New("Can not cast nil to BigFloat")
	}

	switch t := value.(type) {
	case float64:
		(*big.Float)(b).SetFloat64(value.(float64))
	case []uint8:
		_, ok := (*big.Float)(b).SetString(string(value.([]uint8)))
		if !ok {
			return fmt.Errorf("failed to load value to []uint8: %v", value)
		}
	case string:
		_, ok := (*big.Float)(b).SetString(value.(string))
		if !ok {
			return fmt.Errorf("failed to load value to []uint8: %v", value)
		}
	default:
		return fmt.Errorf("Could not scan type %T into BigFloat", t)
	}

	return nil
}

func (b *BigFloat) bigFloat() *big.Float {
	return (*big.Float)(b)
}
func bigFloat(x float64) *big.Float {
	return new(big.Float).SetFloat64(x)
}

func collectSyncCommittee(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {

	slotsPerSyncCommittee := utils.Config.Chain.EpochsPerSyncCommitteePeriod * utils.Config.Chain.SlotsPerEpoch
	currentPeriod := LatestSlot() / slotsPerSyncCommittee
	nextPeriod := currentPeriod + 1

	var validators []struct {
		PubKey string `db:"pubkey"`
		Index  uint64 `db:"validatorindex"`
	}
	err := db.DB.Select(&validators, `SELECT encode(pubkey, 'hex') as pubkey, validators.validatorindex FROM sync_committees LEFT JOIN validators ON validators.validatorindex = sync_committees.validatorindex WHERE period = $1`, nextPeriod)

	if err != nil {
		return err
	}

	if len(validators) <= 0 {
		return nil
	}

	var pubKeys []string
	var mapping map[string]uint64 = make(map[string]uint64)
	for _, val := range validators {
		mapping[val.PubKey] = val.Index
		pubKeys = append(pubKeys, val.PubKey)
	}

	var dbResult []struct {
		SubscriptionID  uint64         `db:"id"`
		UserID          uint64         `db:"user_id"`
		Epoch           uint64         `db:"created_epoch"`
		EventFilter     string         `db:"event_filter"`
		UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	}

	err = db.FrontendDB.Select(&dbResult, `
				SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex')
				FROM users_subscriptions AS us 
				WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '26 hours' OR us.last_sent_ts IS NULL) AND event_filter = ANY($2);
				`,
		utils.GetNetwork()+":"+string(eventName), pq.StringArray(pubKeys),
	)

	if err != nil {
		return err
	}

	for _, r := range dbResult {
		n := &rocketpoolNotification{
			SubscriptionID:  r.SubscriptionID,
			UserID:          r.UserID,
			Epoch:           r.Epoch,
			EventFilter:     r.EventFilter,
			EventName:       eventName,
			ExtraData:       fmt.Sprintf("%v|%v|%v", mapping[r.EventFilter], nextPeriod*utils.Config.Chain.EpochsPerSyncCommitteePeriod, (nextPeriod+1)*utils.Config.Chain.EpochsPerSyncCommitteePeriod),
			UnsubscribeHash: r.UnsubscribeHash,
		}
		if _, exists := notificationsByUserID[r.UserID]; !exists {
			notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
			notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
	}

	return nil
}
