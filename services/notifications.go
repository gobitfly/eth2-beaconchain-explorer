package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
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
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"firebase.google.com/go/messaging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// the notificationCollector is responsible for collecting & queuing notifications
// it is epoch based and will only collect notification for a given epoch once
// notifications are collected in ascending epoch order
// the epochs_notified sql table is used to keep track of already notified epochs
// before collecting notifications several db consistency checks are done
func notificationCollector() {
	for {
		latestFinalizedEpoch := LatestFinalizedEpoch()

		if latestFinalizedEpoch < 4 {
			logger.Errorf("pausing notifications until at least 4 epochs have been exported into the db")
			time.Sleep(time.Minute)
			continue
		}

		var lastNotifiedEpoch uint64
		err := db.WriterDb.Get(&lastNotifiedEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs_notified")

		if err != nil {
			logger.Errorf("error retrieving last notified epoch from the db: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		logger.Infof("latest finalized epoch is %v, latest notified epoch is %v", latestFinalizedEpoch, lastNotifiedEpoch)

		if latestFinalizedEpoch < lastNotifiedEpoch {
			logger.Errorf("notification consistency error, lastest finalized epoch is lower than the last notified epoch!")
			time.Sleep(time.Minute)
			continue
		}

		if latestFinalizedEpoch-lastNotifiedEpoch > 5 {
			logger.Infof("last notified epoch is more than 5 epochs behind the last finalized epoch, limiting lookback to last 5 epochs")
			lastNotifiedEpoch = latestFinalizedEpoch - 5
		}

		for epoch := lastNotifiedEpoch + 1; epoch <= latestFinalizedEpoch; epoch++ {
			var exported uint64
			err := db.WriterDb.Get(&exported, "SELECT COUNT(*) FROM epochs WHERE epoch <= $1 AND epoch >= $2", epoch, epoch-3)
			if err != nil {
				logger.Errorf("error retrieving export status of epoch %v: %v", epoch, err)
				ReportStatus("notification-collector", "Error", nil)
				break
			}

			if exported != 4 {
				logger.Errorf("epoch notification consistency error, epochs %v - %v are not all yet exported into the db (wanted %v, got %v)", epoch, epoch-3, 4, exported)
			}

			start := time.Now()
			logger.Infof("collecting notifications for epoch %v", epoch)

			// Network DB Notifications (network related)
			notifications, err := collectNotifications(epoch)

			if err != nil {
				logger.Errorf("error collection notifications: %v", err)
				ReportStatus("notification-collector", "Error", nil)
				break
			}

			_, err = db.WriterDb.Exec("INSERT INTO epochs_notified VALUES ($1, NOW())", epoch)
			if err != nil {
				logger.Errorf("error marking notification status for epoch %v in db: %v", epoch, err)
				ReportStatus("notification-collector", "Error", nil)
				break
			}

			queueNotifications(notifications, db.FrontendWriterDB) // this caused the collected notifications to be queued and sent

			// Network DB Notifications (user related, must only run on one instance ever!!!!)
			if utils.Config.Notifications.UserDBNotifications {
				userNotifications, err := collectUserDbNotifications(epoch)
				if err != nil {
					logger.Errorf("error collection user db notifications: %v", err)
					ReportStatus("notification-collector", "Error", nil)
					time.Sleep(time.Second * 120)
					continue
				}

				queueNotifications(userNotifications, db.FrontendWriterDB)
			}

			logger.
				WithField("notifications", len(notifications)).
				WithField("duration", time.Since(start)).
				WithField("epoch", epoch).
				Info("notifications completed")

			metrics.TaskDuration.WithLabelValues("service_notifications").Observe(time.Since(start).Seconds())
		}

		ReportStatus("notification-collector", "Running", nil)
		time.Sleep(time.Second * 10)
	}
}

func notificationSender() {
	for {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)

		conn, err := db.FrontendReaderDB.Conn(ctx)
		if err != nil {
			logger.WithError(err).Error("error creating connection")
			cancel()
			continue
		}

		_, err = conn.ExecContext(ctx, `SELECT pg_advisory_lock(500)`)
		if err != nil {
			logger.WithError(err).Errorf("error getting advisory lock from db")

			conn.Close()
			if err != nil {
				logger.WithError(err).Warn("error returning connection to connection pool (advisory lock)")
			}
			cancel()
			continue
		}

		logger.Info("lock obtained")
		err = dispatchNotifications(db.FrontendWriterDB)
		if err != nil {
			logger.WithError(err).Error("error dispatching notifications")
		}

		err = garbageCollectNotificationQueue(db.FrontendWriterDB)
		if err != nil {
			logger.WithError(err).Errorf("error garbage collecting the notification queue")
		}
		logger.WithField("duration", time.Since(start)).Info("notifications dispatched and garbage collected")
		metrics.TaskDuration.WithLabelValues("service_notifications_sender").Observe(time.Since(start).Seconds())

		unlocked := false
		rows, err := conn.QueryContext(ctx, `SELECT pg_advisory_unlock(500)`)
		if err != nil {
			logger.WithError(err).Errorf("error executing advisory unlock")

			err = conn.Close()
			if err != nil {
				logger.WithError(err).Warn("error returning connection to connection pool (advisory unlock)")
			}
			cancel()
			continue
		}

		for rows.Next() {
			rows.Scan(&unlocked)
		}

		if !unlocked {
			utils.LogError(nil, fmt.Errorf("error releasing advisory lock unlocked: %v", unlocked), 0)
		}

		conn.Close()
		if err != nil {
			logger.Warn("error returning connection to connection pool")
		}
		cancel()

		ReportStatus("notification-sender", "Running", nil)
		time.Sleep(time.Second * 30)
	}
}

func collectNotifications(epoch uint64) (map[uint64]map[types.EventName][]types.Notification, error) {
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	start := time.Now()
	var err error
	var dbIsCoherent bool

	err = db.WriterDb.Get(&dbIsCoherent, `
		select 
			not (array[false] && array_agg(is_coherent)) as is_coherent
		from (
			select 
				epoch - 1 = lead(epoch) over (order by epoch desc) as is_coherent
			from epochs
			order by epoch desc 
			limit 2^14
		) coherency`)

	if err != nil {
		logger.Errorf("failed to do epochs table coherence check, aborting: %v", err)
		return nil, err
	}
	if !dbIsCoherent {
		logger.Errorf("epochs coherence check failed, aborting.")
		return nil, fmt.Errorf("epochs coherence check failed, aborting")
	}

	logger.Infof("Started collecting notifications")

	err = collectAttestationAndOfflineValidatorNotifications(notificationsByUserID, 0, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_missed_attestation").Inc()
		return nil, fmt.Errorf("error collecting validator_attestation_missed notifications: %v", err)
	}
	logger.Infof("collecting attestation & offline notifications took: %v\n", time.Since(start))

	err = collectBlockProposalNotifications(notificationsByUserID, 1, types.ValidatorExecutedProposalEventName, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_executed_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting validator_proposal_submitted notifications: %v", err)
	}
	logger.Infof("collecting block proposal proposed notifications took: %v\n", time.Since(start))

	err = collectBlockProposalNotifications(notificationsByUserID, 2, types.ValidatorMissedProposalEventName, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_missed_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting validator_proposal_missed notifications: %v", err)
	}
	logger.Infof("collecting block proposal missed notifications took: %v\n", time.Since(start))

	err = collectValidatorGotSlashedNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_validator_got_slashed").Inc()
		return nil, fmt.Errorf("error collecting validator_got_slashed notifications: %v", err)
	}
	logger.Infof("collecting validator got slashed notifications took: %v\n", time.Since(start))

	err = collectWithdrawalNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_validator_withdrawal").Inc()
		return nil, fmt.Errorf("error collecting withdrawal notifications: %v", err)
	}
	logger.Infof("collecting withdrawal notifications took: %v\n", time.Since(start))

	err = collectNetworkNotifications(notificationsByUserID, types.NetworkLivenessIncreasedEventName)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_network").Inc()
		return nil, fmt.Errorf("error collecting network notifications: %v", err)
	}
	logger.Infof("collecting network notifications took: %v\n", time.Since(start))

	// Rocketpool
	{
		var ts int64
		err = db.ReaderDb.Get(&ts, `SELECT id FROM rocketpool_network_stats LIMIT 1;`)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.Infof("skipped the collecting of rocketpool notifications, because rocketpool_network_stats is empty")
			} else {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_notifications").Inc()
				return nil, fmt.Errorf("error collecting rocketpool notifications: %v", err)
			}
		} else {
			err = collectRocketpoolComissionNotifications(notificationsByUserID, types.RocketpoolCommissionThresholdEventName)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_comission").Inc()
				return nil, fmt.Errorf("error collecting rocketpool commission: %v", err)
			}
			logger.Infof("collecting rocketpool commissions took: %v\n", time.Since(start))

			err = collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID, types.RocketpoolNewClaimRoundStartedEventName)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_reward_claim").Inc()
				return nil, fmt.Errorf("error collecting new rocketpool claim round: %v", err)
			}
			logger.Infof("collecting rocketpool claim round took: %v\n", time.Since(start))

			err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolColleteralMaxReached, epoch)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_max_reached").Inc()
				return nil, fmt.Errorf("error collecting rocketpool max collateral: %v", err)
			}
			logger.Infof("collecting rocketpool max collateral took: %v\n", time.Since(start))

			err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolColleteralMinReached, epoch)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_min_reached").Inc()
				return nil, fmt.Errorf("error collecting rocketpool min collateral: %v", err)
			}
			logger.Infof("collecting rocketpool min collateral took: %v\n", time.Since(start))
		}
	}

	err = collectSyncCommittee(notificationsByUserID, types.SyncCommitteeSoon, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_sync_committee").Inc()
		return nil, fmt.Errorf("error collecting sync committee: %v", err)
	}
	logger.Infof("collecting sync committee took: %v\n", time.Since(start))

	return notificationsByUserID, nil
}

func collectUserDbNotifications(epoch uint64) (map[uint64]map[types.EventName][]types.Notification, error) {
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	var err error

	// Monitoring (premium): machine offline
	err = collectMonitoringMachineOffline(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_offline").Inc()
		return nil, fmt.Errorf("error collecting Eth client offline notifications: %v", err)
	}

	// Monitoring (premium): disk full warnings
	err = collectMonitoringMachineDiskAlmostFull(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_disk_almost_full").Inc()
		return nil, fmt.Errorf("error collecting Eth client disk full notifications: %v", err)
	}

	// Monitoring (premium): cpu load
	err = collectMonitoringMachineCPULoad(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_cpu_load").Inc()
		return nil, fmt.Errorf("error collecting Eth client cpu notifications: %v", err)
	}

	// Monitoring (premium): ram
	err = collectMonitoringMachineMemoryUsage(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_memory_usage").Inc()
		return nil, fmt.Errorf("error collecting Eth client memory notifications: %v", err)
	}

	// New ETH clients
	err = collectEthClientNotifications(notificationsByUserID, types.EthClientUpdateEventName)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_eth_client").Inc()
		return nil, fmt.Errorf("error collecting Eth client notifications: %v", err)
	}

	//Tax Report
	err = collectTaxReportNotificationNotifications(notificationsByUserID, types.TaxReportEventName)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_tax_report").Inc()
		return nil, fmt.Errorf("error collecting tax report notifications: %v", err)
	}

	return notificationsByUserID, nil
}

func queueNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) {
	subByEpoch := map[uint64][]uint64{}

	// prevent multiple events being sent with the same subscription id
	for user, notifications := range notificationsByUserID {
		for eventType, events := range notifications {
			filteredEvents := make([]types.Notification, 0)

			for _, ev := range events {
				isDuplicate := false
				for _, fe := range filteredEvents {
					if fe.GetSubscriptionID() == ev.GetSubscriptionID() {
						isDuplicate = true
					}
				}
				if !isDuplicate {
					filteredEvents = append(filteredEvents, ev)
				}
			}
			notificationsByUserID[user][eventType] = filteredEvents
		}
	}

	err := queueEmailNotifications(notificationsByUserID, useDB)
	if err != nil {
		logger.WithError(err).Error("error queuing email notifications")
	}

	err = queuePushNotification(notificationsByUserID, useDB)
	if err != nil {
		logger.WithError(err).Error("error queuing push notifications")
	}

	err = queueWebhookNotifications(notificationsByUserID, useDB)
	if err != nil {
		logger.WithError(err).Error("error queuing webhook notifications")
	}

	for _, events := range notificationsByUserID {
		for _, notifications := range events {
			for _, n := range notifications {
				e := n.GetEpoch()
				if _, exists := subByEpoch[e]; !exists {
					subByEpoch[e] = []uint64{n.GetSubscriptionID()}
				} else {
					subByEpoch[e] = append(subByEpoch[e], n.GetSubscriptionID())
				}
			}
		}
	}
	for epoch, subIDs := range subByEpoch {
		// update that we've queued the subscription (last sent rather means last queued)
		err := db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch, useDB)
		if err != nil {
			logger.Errorf("error updating sent-time of sent notifications: %v", err)
			metrics.Errors.WithLabelValues("notifications_updating_sent_time").Inc()
		}
	}
	// update internal state of subscriptions
	stateToSub := make(map[string]map[uint64]bool, 0)

	for _, notificationMap := range notificationsByUserID { // _ => user
		for _, notifications := range notificationMap { // _ => eventname
			for _, notification := range notifications { // _ => index
				state := notification.GetLatestState()
				if state == "" {
					continue
				}
				if _, exists := stateToSub[state]; !exists {
					stateToSub[state] = make(map[uint64]bool, 0)
				}
				if _, exists := stateToSub[state][notification.GetSubscriptionID()]; !exists {
					stateToSub[state][notification.GetSubscriptionID()] = true
				}
			}
		}
	}

	for state, subs := range stateToSub {
		subArray := make([]int64, 0)
		for subID := range subs {
			subArray = append(subArray, int64(subID))
		}
		_, err := db.FrontendWriterDB.Exec(`UPDATE users_subscriptions SET internal_state = $1 WHERE id = ANY($2)`, state, pq.Int64Array(subArray))
		if err != nil {
			logger.Errorf("failed to update internal state of notifcations: %v", err)
		}
	}
}

func dispatchNotifications(useDB *sqlx.DB) error {

	err := sendEmailNotifications(useDB)
	if err != nil {
		return fmt.Errorf("error sending email notifications, err: %w", err)
	}

	err = sendPushNotifications(useDB)
	if err != nil {
		return fmt.Errorf("error sending push notifications, err: %w", err)
	}

	err = sendWebhookNotifications(useDB)
	if err != nil {
		return fmt.Errorf("error sending webhook notifications, err: %w", err)
	}

	err = sendDiscordNotifications(useDB)
	if err != nil {
		return fmt.Errorf("error sending webhook discord notifications, err: %w", err)
	}

	return nil
}

// garbageCollectNotificationQueue deletes entries from the notification queue that have been processed
func garbageCollectNotificationQueue(useDB *sqlx.DB) error {

	rows, err := useDB.Exec(`DELETE FROM notification_queue where (sent < now() - INTERVAL '30 minutes') OR (created < now() - INTERVAL '1 hour')`)
	if err != nil {
		return fmt.Errorf("error deleting from notification_queue %w", err)
	}

	rowsAffected, _ := rows.RowsAffected()

	logger.Infof("Deleted %v rows from the notification_queue", rowsAffected)

	return nil
}

func getNetwork() string {
	domainParts := strings.Split(utils.Config.Frontend.SiteDomain, ".")
	if len(domainParts) >= 3 {
		return fmt.Sprintf("%s: ", cases.Title(language.English).String(domainParts[0]))
	}
	return ""
}

func queuePushNotification(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) error {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}

	tokensByUserID, err := db.GetUserPushTokenByIds(userIDs)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_send_push_notifications").Inc()
		return fmt.Errorf("error when sending push-notifications: could not get tokens: %w", err)
	}

	for userID, userNotifications := range notificationsByUserID {
		userTokens, exists := tokensByUserID[userID]
		if !exists {
			continue
		}

		go func(userTokens []string, userNotifications map[types.EventName][]types.Notification) {
			var batch []*messaging.Message
			for event, ns := range userNotifications {
				for _, n := range ns {
					added := false
					for _, userToken := range userTokens {
						notification := new(messaging.Notification)
						notification.Title = fmt.Sprintf("%s%s", getNetwork(), n.GetTitle())
						notification.Body = n.GetInfo(false)
						if notification.Body == "" {
							continue
						}
						added = true

						message := new(messaging.Message)
						message.Notification = notification
						message.Token = userToken

						message.APNS = new(messaging.APNSConfig)
						message.APNS.Payload = new(messaging.APNSPayload)
						message.APNS.Payload.Aps = new(messaging.Aps)
						message.APNS.Payload.Aps.Sound = "default"

						batch = append(batch, message)
					}
					if added {
						metrics.NotificationsQueued.WithLabelValues("push", string(event)).Inc()
					}
				}
			}

			transitPushContent := types.TransitPushContent{
				Messages: batch,
			}

			_, err = useDB.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES ($1, 'push', $2)`, time.Now(), transitPushContent)
			if err != nil {
				logger.WithError(err).Errorf("error writing transit push notification to db")
				return
			}
		}(userTokens, userNotifications)
	}
	return nil
}

func sendPushNotifications(useDB *sqlx.DB) error {
	var notificationQueueItem []types.TransitPush

	err := useDB.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue where sent is null and channel = 'push' order by created asc`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}

	logger.Infof("processing %v push notifications", len(notificationQueueItem))

	for _, n := range notificationQueueItem {
		err = notify.SendPushBatch(n.Content.Messages)
		if err != nil {
			metrics.Errors.WithLabelValues("notifications_send_push_batch").Inc()
			logger.WithError(err).Error("error sending firebase batch job")
		} else {
			metrics.NotificationsSent.WithLabelValues("push", "200").Add(float64(len(n.Content.Messages)))
		}

		_, err = useDB.Exec(`UPDATE notification_queue set sent = now() where id = $1`, n.Id)
		if err != nil {
			return fmt.Errorf("error updating sent status for push notification with id: %v, err: %w", n.Id, err)
		}
	}
	return nil
}

func queueEmailNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) error {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}
	emailsByUserID, err := db.GetUserEmailsByIds(userIDs)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_get_user_mail_by_id").Inc()
		return fmt.Errorf("error when sending email-notifications: could not get emails: %w", err)
	}

	for userID, userNotifications := range notificationsByUserID {
		userEmail, exists := emailsByUserID[userID]
		if !exists {
			logger.Warnf("email notification skipping user %v", userID)
			// we don't need this metrics as users can now deactivate email notifications and it would increment the counter
			// metrics.Errors.WithLabelValues("notifications_mail_not_found").Inc()
			continue
		}
		go func(userEmail string, userNotifications map[types.EventName][]types.Notification) {
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

			if utils.Config.Chain.Name != "mainnet" {
				msg.Body += template.HTML(fmt.Sprintf("<b>Notice: This email contains notifications for the %s network!</b><br>", utils.Config.Chain.Name))
			}

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

						tx, err := db.FrontendWriterDB.Beginx()
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
					msg.UnSubURL = template.HTML(fmt.Sprintf(`<a style="color: white" onMouseOver="this.style.color='#F5B498'" onMouseOut="this.style.color='#FFFFFF'" href="%v">Unsubscribe</a>`, unsubURL))
					msg.Body += template.HTML(fmt.Sprintf("%s<br>", n.GetInfo(true)))
					if att := n.GetEmailAttachment(); att != nil {
						attachments = append(attachments, *att)
					}

					metrics.NotificationsQueued.WithLabelValues("email", string(event)).Inc()
				}
				if event == "validator_balance_decreased" {
					msg.Body += template.HTML("<br>You will not receive any further balance decrease mails for these validators until the balance of a validator is increasing again.<br>")
				}
			}

			// msg.Body += template.HTML(fmt.Sprintf("<br>Best regards<br>\n%s", utils.Config.Frontend.SiteDomain))
			msg.SubscriptionManageURL = template.HTML(fmt.Sprintf(`<a href="%v" style="color: white" onMouseOver="this.style.color='#F5B498'" onMouseOut="this.style.color='#FFFFFF'">Manage</a>`, "https://"+utils.Config.Frontend.SiteDomain+"/user/notifications"))

			transitEmailContent := types.TransitEmailContent{
				Address:     userEmail,
				Subject:     subject,
				Email:       msg,
				Attachments: attachments,
			}

			_, err = useDB.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES ($1, 'email', $2)`, time.Now(), transitEmailContent)
			if err != nil {
				logger.WithError(err).Errorf("error writing transit email to db")
			}
		}(userEmail, userNotifications)
	}
	return nil
}

func sendEmailNotifications(useDb *sqlx.DB) error {
	var notificationQueueItem []types.TransitEmail

	err := useDb.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue where sent is null and channel = 'email' order by created asc`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}

	logger.Infof("processing %v email notifications", len(notificationQueueItem))

	for _, n := range notificationQueueItem {
		err = mail.SendMailRateLimited(n.Content.Address, n.Content.Subject, n.Content.Email, n.Content.Attachments)
		if err != nil {
			if !strings.Contains(err.Error(), "rate limit has been exceeded") {
				metrics.Errors.WithLabelValues("notifications_send_email").Inc()
				logger.WithError(err).Error("error sending email notification")
				// 	_, err := tx.Exec(`DELETE FROM notification_queue where id = $1`, n.Id)
				// 	if err != nil {
				// 		return fmt.Errorf("error deleting from notification queue: %w", err)
				// 	}
				// 	err = tx.Commit()
				// 	if err != nil {
				// 		tx.Rollback()
				// 		return fmt.Errorf("error committing transaction")
				// 	}
				// 	continue
			} else {
				metrics.NotificationsSent.WithLabelValues("email", "200").Inc()
			}
			//else {
			// 	tx.Rollback()
			// 	return fmt.Errorf("error sending notification-email: %w", err)
			// }
		}
		_, err = useDb.Exec(`UPDATE notification_queue set sent = now() where id = $1`, n.Id)
		if err != nil {
			return fmt.Errorf("error updating sent status for email notification with id: %v, err: %w", n.Id, err)
		}
	}
	return nil
}

func queueWebhookNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, useDB *sqlx.DB) error {
	for userID, userNotifications := range notificationsByUserID {
		var webhooks []types.UserWebhook
		err := useDB.Select(&webhooks, `
			SELECT
				id,
				user_id,
				url,
				retries,
				event_names,
				destination
			FROM 
				users_webhooks
			WHERE 
				user_id = $1 AND user_id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)
		`, userID, types.WebhookNotificationChannel)
		// continue if the user does not have a webhook
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("error quering users_webhooks, err: %w", err)
		}
		// webhook => [] notifications
		discordNotifMap := make(map[uint64][]types.TransitDiscordContent)
		notifs := make([]types.TransitWebhook, 0)
		// send the notifications to each registered webhook
		for _, w := range webhooks {
			for event, notifications := range userNotifications {
				eventSubscribed := false
				// check if the webhook is subscribed to the type of event
				for _, w := range w.EventNames {
					if w == string(event) {
						eventSubscribed = true
						break
					}
				}
				if eventSubscribed {
					if len(notifications) > 0 {
						// reset Retries
						if w.Retries > 5 && w.LastSent.Valid && w.LastSent.Time.Add(time.Hour).Before(time.Now()) {
							_, err = useDB.Exec(`UPDATE users_webhooks SET retries = 0 WHERE id = $1;`, w.ID)
							if err != nil {
								logger.WithError(err).Errorf("error updating users_webhooks table; setting retries to zero")
								continue
							}
						} else if w.Retries > 5 && !w.LastSent.Valid {
							logger.Warn("error webhook has more than 5 retries and does not have a valid last_sent timestamp")
							continue
						}

						if w.Retries >= 5 {
							// early return
							continue
						}
					}

					for _, n := range notifications {
						if w.Destination.Valid && w.Destination.String == "webhook_discord" {
							if _, exists := discordNotifMap[w.ID]; !exists {
								discordNotifMap[w.ID] = make([]types.TransitDiscordContent, 0)
							}
							l_notifs := len(discordNotifMap[w.ID])
							if l_notifs == 0 || len(discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds) >= 10 {
								discordNotifMap[w.ID] = append(discordNotifMap[w.ID], types.TransitDiscordContent{
									Webhook: w,
									DiscordRequest: types.DiscordReq{
										Username: utils.Config.Frontend.SiteDomain,
									},
								})
								l_notifs++
							}

							fields := []types.DiscordEmbedField{
								{
									Name:   "Epoch",
									Value:  fmt.Sprintf("[%[1]v](https://%[2]s/%[1]v)", n.GetEpoch(), utils.Config.Frontend.SiteDomain+"/epoch"),
									Inline: false,
								},
							}

							if strings.HasPrefix(string(n.GetEventName()), "monitoring") || n.GetEventName() == types.EthClientUpdateEventName || n.GetEventName() == types.RocketpoolColleteralMaxReached || n.GetEventName() == types.RocketpoolColleteralMinReached {
								fields = append(fields,
									types.DiscordEmbedField{
										Name:   "Target",
										Value:  fmt.Sprintf("%v", n.GetEventFilter()),
										Inline: false,
									})
							}
							discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds = append(discordNotifMap[w.ID][l_notifs-1].DiscordRequest.Embeds, types.DiscordEmbed{
								Type:        "rich",
								Color:       "16745472",
								Description: n.GetInfoMarkdown(),
								Title:       n.GetTitle(),
								Fields:      fields,
							})
						} else {
							notifs = append(notifs, types.TransitWebhook{
								Channel: w.Destination.String,
								Content: types.TransitWebhookContent{
									Webhook: w,
									Event: types.WebhookEvent{
										Network:     utils.GetNetwork(),
										Name:        string(n.GetEventName()),
										Title:       n.GetTitle(),
										Description: n.GetInfo(false),
										Epoch:       n.GetEpoch(),
										Target:      n.GetEventFilter(),
									},
								},
							})
						}
					}
				}
			}
		}
		// process notifs
		for _, n := range notifs {
			_, err = useDB.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES (now(), $1, $2);`, n.Channel, n.Content)
			if err != nil {
				logger.WithError(err).Errorf("error inserting into webhooks_queue")
			} else {
				metrics.NotificationsQueued.WithLabelValues(n.Channel, n.Content.Event.Name).Inc()
			}
		}
		// process discord notifs
		for _, dNotifs := range discordNotifMap {
			for _, n := range dNotifs {
				_, err = useDB.Exec(`INSERT INTO notification_queue (created, channel, content) VALUES (now(), 'webhook_discord', $1);`, n)
				if err != nil {
					logger.WithError(err).Errorf("error inserting into webhooks_queue (discord)")
					continue
				} else {
					metrics.NotificationsQueued.WithLabelValues("webhook_discord", "multi").Inc()
				}
			}
		}
	}
	return nil
}

func sendWebhookNotifications(useDB *sqlx.DB) error {
	var notificationQueueItem []types.TransitWebhook

	err := useDB.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue where sent is null and channel = 'webhook' order by created asc`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}
	client := &http.Client{Timeout: time.Second * 30}

	logger.Infof("processing %v webhook notifications", len(notificationQueueItem))

	// now := time.Now()
	for _, n := range notificationQueueItem {
		// do not retry after 5 attempts
		if n.Content.Webhook.Retries > 5 {
			// if n.Content.Webhook.LastSent.Valid && n.Content.Webhook.LastSent.Time.Add(time.Hour*1).Before(now) {
			// 	_, err = useDB.Exec(`UPDATE users_webhooks SET retries = 0 WHERE id = $1;`, n.Content.Webhook.ID)
			// 	if err != nil {
			// 		logger.WithError(err).Errorf("error updating users_webhooks table; setting retries to zero")
			// 		continue
			// 	}
			// } else {
			_, err := db.FrontendWriterDB.Exec(`DELETE FROM notification_queue where id = $1`, n.Id)
			if err != nil {
				return fmt.Errorf("error deleting from notification queue: %w", err)
			}
			continue
			// }
		}

		reqBody := new(bytes.Buffer)

		err := json.NewEncoder(reqBody).Encode(n.Content)
		if err != nil {
			logger.WithError(err).Errorf("error marschalling webhook event")
		}

		_, err = url.Parse(n.Content.Webhook.Url)
		if err != nil {
			_, err := db.FrontendWriterDB.Exec(`DELETE FROM notification_queue where id = $1`, n.Id)
			if err != nil {
				return fmt.Errorf("error deleting from notification queue: %w", err)
			}
			continue
		}

		go func(n types.TransitWebhook) {
			if n.Content.Webhook.Retries > 0 {
				time.Sleep(time.Duration(n.Content.Webhook.Retries) * time.Second)
			}
			resp, err := client.Post(n.Content.Webhook.Url, "application/json", reqBody)
			if err != nil {
				logger.WithError(err).Errorf("error sending request")
			} else {
				metrics.NotificationsSent.WithLabelValues("webhook", resp.Status).Inc()
			}

			if resp != nil && resp.StatusCode < 400 {
				_, err := useDB.Exec(`UPDATE notification_queue SET sent = now();`)
				if err != nil {
					logger.WithError(err).Errorf("error updating notification_queue table")
					return
				}

				_, err = useDB.Exec(`UPDATE users_webhooks SET retries = 0, last_sent = now() WHERE id = $1;`, n.Content.Webhook.ID)
				if err != nil {
					logger.WithError(err).Errorf("error updating users_webhooks table; setting retries to zero")
					return
				}
			} else {
				var errResp types.ErrorResponse

				if resp != nil {
					b, err := io.ReadAll(resp.Body)
					if err != nil {
						logger.WithError(err).Error("error reading body")
					}

					errResp.Status = resp.Status
					errResp.Body = string(b)
				}

				_, err = useDB.Exec(`UPDATE users_webhooks SET retries = retries + 1, last_sent = now(), request = $2, response = $3 WHERE id = $1;`, n.Content.Webhook.ID, n.Content, errResp)
				if err != nil {
					logger.WithError(err).Errorf("error updating users_webhooks table; increasing retries")
					return
				}
			}
		}(n)

	}
	return nil
}

func sendDiscordNotifications(useDB *sqlx.DB) error {
	var notificationQueueItem []types.TransitDiscord

	err := useDB.Select(&notificationQueueItem, `SELECT
		id,
		created,
		sent,
		channel,
		content
	FROM notification_queue where sent is null and channel = 'webhook_discord' order by created asc`)
	if err != nil {
		return fmt.Errorf("error querying notification queue, err: %w", err)
	}
	client := &http.Client{Timeout: time.Second * 30}

	logger.Infof("processing %v discord webhook notifications", len(notificationQueueItem))
	webhookMap := make(map[uint64]types.UserWebhook)

	notifMap := make(map[uint64][]types.TransitDiscord)
	// generate webhook id => discord req
	// while mapping. aggregate embeds while doing so, up to 10 per req can be sent
	for _, n := range notificationQueueItem {
		// purge the event from existence if the retry counter is over 5
		if n.Content.Webhook.Retries > 5 {
			db.FrontendWriterDB.Exec(`DELETE FROM notification_queue where id = $1`, n.Id)
			continue
		}
		if _, exists := webhookMap[n.Content.Webhook.ID]; !exists {
			webhookMap[n.Content.Webhook.ID] = n.Content.Webhook
		}
		if _, exists := notifMap[n.Content.Webhook.ID]; !exists {
			notifMap[n.Content.Webhook.ID] = make([]types.TransitDiscord, 0)
		}
		notifMap[n.Content.Webhook.ID] = append(notifMap[n.Content.Webhook.ID], n)
	}
	for _, webhook := range webhookMap {
		go func(webhook types.UserWebhook, reqs []types.TransitDiscord) {
			defer func() {
				// update retries counters in db based on end result
				_, err = useDB.Exec(`UPDATE users_webhooks SET retries = $1, last_sent = now() WHERE id = $2;`, webhook.Retries, webhook.ID)
				if err != nil {
					logger.Warnf("failed to update retries counter to %v for webhook %v: %v", webhook.Retries, webhook.ID, err)
				}

				// mark notifcations as sent in db
				ids := make([]uint64, 0)
				for _, req := range reqs {
					ids = append(ids, req.Id)
				}
				_, err = db.FrontendWriterDB.Exec(`UPDATE notification_queue SET sent = now() where id = ANY($1)`, pq.Array(ids))
				if err != nil {
					logger.Warnf("failed to update sent for notifcations in queue: %v", err)
				}
			}()

			_, err = url.Parse(webhook.Url)
			if err != nil {
				logger.Errorf("invalid url for webhook id %v: %v", webhook.ID, err)
				return
			}

			for i := 0; i < len(reqs); i++ {
				if webhook.Retries > 5 {
					break // stop
				}
				// sleep between retries
				time.Sleep(time.Duration(webhook.Retries) * time.Second)

				reqBody := new(bytes.Buffer)
				err := json.NewEncoder(reqBody).Encode(reqs[i].Content.DiscordRequest)
				if err != nil {
					logger.Errorf("error marschalling discord webhook event: %v", err)
					continue // skip
				}

				logger.Infof("discord request webhook body: %s", reqBody.String())
				resp, err := client.Post(webhook.Url, "application/json", reqBody)
				if err != nil {
					logger.Errorf("error sending discord webhook request: %v", err)
				} else {
					metrics.NotificationsSent.WithLabelValues("webhook_discord", resp.Status).Inc()
				}
				if resp != nil && resp.StatusCode < 400 {
					webhook.Retries = 0
				} else {
					webhook.Retries++
					var errResp types.ErrorResponse

					if resp != nil {
						b, err := io.ReadAll(resp.Body)
						if err != nil {
							logger.Errorf("error reading body for discord webhook response: %v", err)
						} else {
							errResp.Body = string(b)
						}
						errResp.Status = resp.Status
					}
					logger.Errorf("error pushing discord webhook: %v", errResp.Body)

					_, err = useDB.Exec(`UPDATE users_webhooks SET request = $2, response = $3 WHERE id = $1;`, webhook.ID, reqs[i].Content.DiscordRequest, errResp)
					if err != nil {
						logger.Errorf("error storing failure data in users_webhooks table: %v", err)
					}

					i-- // retry, IMPORTANT to be at the END of the ELSE, otherwise the wrong index will be used in the commands above!
				}
			}
		}(webhook, notifMap[webhook.ID])
	}

	return nil
}

func getUrlPart(validatorIndex uint64) string {
	return fmt.Sprintf(` For more information visit: https://%s/validator/%v.`, utils.Config.Frontend.SiteDomain, validatorIndex)
}

func collectBlockProposalNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, status uint64, eventName types.EventName, epoch uint64) error {
	type dbResult struct {
		Proposer      uint64 `db:"proposer"`
		Status        uint64 `db:"status"`
		Slot          uint64 `db:"slot"`
		ExecBlock     uint64 `db:"exec_block_number"`
		ExecRewardETH float64
	}

	_, subMap, err := db.GetSubsForEventFilter(eventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missted attestations %w", err)
	}

	events := make([]dbResult, 0)
	err = db.WriterDb.Select(&events, "SELECT slot, proposer, status, COALESCE(exec_block_number, 0) AS exec_block_number FROM blocks WHERE epoch = $1 AND status = $2", epoch, fmt.Sprintf("%d", status))
	if err != nil {
		return fmt.Errorf("error retrieving slots for epoch %v: %w", epoch, err)
	}

	logger.Infof("retrieved %v events", len(events))

	// Get Execution reward for proposed blocks
	if status == 1 { // if proposed
		var blockList = []uint64{}
		for _, data := range events {
			if data.ExecBlock != 0 {
				blockList = append(blockList, data.ExecBlock)
			}
		}

		if len(blockList) > 0 {
			blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, 10000)
			if err != nil {
				logger.WithError(err).Errorf("can not load blocks from bigtable for notification")
				return err
			}
			var execBlockNrToExecBlockMap = map[uint64]*types.Eth1BlockIndexed{}
			for _, block := range blocks {
				execBlockNrToExecBlockMap[block.GetNumber()] = block
			}
			relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
			if err != nil {
				return err
			}

			for j := 0; j < len(events); j++ {
				execData, found := execBlockNrToExecBlockMap[events[j].ExecBlock]
				if found {
					reward := utils.Eth1TotalReward(execData)
					relayData, found := relaysData[common.BytesToHash(execData.Hash)]
					if found {
						reward = relayData.MevBribe.BigInt()
					}
					events[j].ExecRewardETH = float64(int64(eth.WeiToEth(reward)*100000)) / 100000
				}
			}
		}
	}

	for _, event := range events {
		pubkey, err := GetGetPubkeyForIndex(event.Proposer)
		if err == nil {
			subscribers, ok := subMap[hex.EncodeToString(pubkey)]
			if ok {
				for _, sub := range subscribers {
					if sub.UserID == nil || sub.ID == nil {
						return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
					}
					if sub.LastEpoch != nil {
						lastSentEpoch := *sub.LastEpoch
						if lastSentEpoch >= epoch || epoch < sub.CreatedEpoch {
							continue
						}
					}
					logger.Infof("creating %v notification for validator %v in epoch %v", eventName, event.Proposer, epoch)
					n := &validatorProposalNotification{
						SubscriptionID: *sub.ID,
						ValidatorIndex: event.Proposer,
						Epoch:          epoch,
						Status:         event.Status,
						EventName:      eventName,
						Reward:         event.ExecRewardETH,
						EventFilter:    hex.EncodeToString(pubkey),
					}
					if _, exists := notificationsByUserID[*sub.UserID]; !exists {
						notificationsByUserID[*sub.UserID] = map[types.EventName][]types.Notification{}
					}
					if _, exists := notificationsByUserID[*sub.UserID][n.GetEventName()]; !exists {
						notificationsByUserID[*sub.UserID][n.GetEventName()] = []types.Notification{}
					}
					notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], n)
					metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
				}
			}
		} else {
			logger.Errorf("error retrieving pubkey for validator %v: %v", event.Proposer, err)
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
	Reward             float64
	UnsubscribeHash    sql.NullString
}

func (n *validatorProposalNotification) GetLatestState() string {
	return ""
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
		generalPart = fmt.Sprintf(`New scheduled block proposal for Validator %v.`, n.ValidatorIndex)
	case 1:
		generalPart = fmt.Sprintf(`Validator %v proposed a new block with %v ETH execution reward.`, n.ValidatorIndex, n.Reward)
	case 2:
		generalPart = fmt.Sprintf(`Validator %v missed a block proposal.`, n.ValidatorIndex)
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

func (n *validatorProposalNotification) GetInfoMarkdown() string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`New scheduled block proposal for Validator [%[1]v](https://%[2]v/%[1]v).`, n.ValidatorIndex, utils.Config.Frontend.SiteDomain+"/validator")
	case 1:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[2]v/%[1]v) proposed a new block with %[3]v ETH execution reward.`, n.ValidatorIndex, utils.Config.Frontend.SiteDomain+"/validator", n.Reward)
	case 2:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[2]v/%[1]v) missed a block proposal.`, n.ValidatorIndex, utils.Config.Frontend.SiteDomain+"/validator")
	}

	return generalPart
}

func collectAttestationAndOfflineValidatorNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, status uint64, epoch uint64) error {
	_, subMap, err := db.GetSubsForEventFilter(types.ValidatorMissedAttestationEventName)
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

	// get attestations for all validators for the last n epochs
	attestations, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{}, epoch-4, epoch) // retrieve attestation data of the last 3 epochs
	if err != nil {
		return fmt.Errorf("error getting validator attestations from bigtable %w", err)
	}

	logger.Infof("retrieved validator attestation history data")

	events := make([]dbResult, 0)

	epochAttested := make(map[uint64]uint64)
	epochTotal := make(map[uint64]uint64)
	participationPerEpoch := make(map[uint64]map[uint64]int) // map[validatorindex]map[epoch]attested
	for validator, history := range attestations {
		for _, attestation := range history {
			if participationPerEpoch[validator] == nil {
				participationPerEpoch[validator] = make(map[uint64]int, 4)
			}
			epochTotal[attestation.Epoch] = epochTotal[attestation.Epoch] + 1 // count the total attestations for each epoch

			if attestation.Status == 0 {

				participationPerEpoch[validator][attestation.Epoch] = 1 // missed

				pubkey, err := GetGetPubkeyForIndex(validator)
				if err == nil {
					if attestation.Epoch != epoch || subMap[hex.EncodeToString(pubkey)] == nil {
						continue
					}

					events = append(events, dbResult{
						ValidatorIndex: validator,
						Epoch:          attestation.Epoch,
						Status:         attestation.Status,
						Slot:           attestation.AttesterSlot,
						InclusionSlot:  attestation.InclusionSlot,
						EventFilter:    pubkey,
					})
				} else {
					logger.Errorf("error retrieving pubkey for validator %v: %v", validator, err)
				}
			} else {
				participationPerEpoch[validator][attestation.Epoch] = 2 // attested

				epochAttested[attestation.Epoch] = epochAttested[attestation.Epoch] + 1 // count the total attested attestation for each epoch (exlude missing)
			}
		}
	}

	// process missed attestation events
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
					// logger.Infof("skipping creating %v for validator %v (lastSentEpoch: %v, createdEpoch: %v)", types.ValidatorMissedAttestationEventName, event.ValidatorIndex, lastSentEpoch, sub.CreatedEpoch)
					continue
				}
			}

			logger.Infof("creating %v notification for validator %v in epoch %v", types.ValidatorMissedAttestationEventName, event.ValidatorIndex, event.Epoch)
			n := &validatorAttestationNotification{
				SubscriptionID: *sub.ID,
				ValidatorIndex: event.ValidatorIndex,
				Epoch:          event.Epoch,
				Status:         event.Status,
				EventName:      types.ValidatorMissedAttestationEventName,
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	// detect online & offline validators
	type indexPubkeyPair struct {
		Index  uint64
		Pubkey []byte
	}
	var offlineValidators []*indexPubkeyPair
	var onlineValidators []*indexPubkeyPair

	epochNMinus1 := epoch - 1
	epochNMinus2 := epoch - 2
	epochNMinus3 := epoch - 3

	if epochTotal[epoch] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epoch)
	}
	if epochTotal[epochNMinus1] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus1)
	}
	if epochTotal[epochNMinus2] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus2)
	}
	if epochTotal[epochNMinus3] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus3)
	}

	if epochAttested[epoch]*100/epochTotal[epoch] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epoch, epochTotal[epoch], epochAttested[epoch])
	}
	if epochAttested[epochNMinus1]*100/epochTotal[epochNMinus1] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus1, epochTotal[epochNMinus1], epochAttested[epochNMinus1])
	}
	if epochAttested[epochNMinus2]*100/epochTotal[epochNMinus2] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus2, epochTotal[epochNMinus2], epochAttested[epochNMinus2])
	}
	if epochAttested[epochNMinus3]*100/epochTotal[epochNMinus3] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus3, epochTotal[epochNMinus3], epochAttested[epochNMinus3])
	}

	for validator, participation := range participationPerEpoch {
		if participation[epochNMinus3] == 2 && participation[epochNMinus2] == 1 && participation[epochNMinus1] == 1 && participation[epoch] == 1 {
			logger.Infof("validator %v detected as offline in epoch %v (did not attest since epoch %v)", validator, epoch, epochNMinus2)
			pubkey, err := GetGetPubkeyForIndex(validator)
			if err != nil {
				return err
			}
			offlineValidators = append(offlineValidators, &indexPubkeyPair{Index: validator, Pubkey: pubkey})
		}

		if participation[epochNMinus3] == 1 && participation[epochNMinus2] == 1 && participation[epochNMinus1] == 1 && participation[epoch] == 2 {
			logger.Infof("validator %v detected as online in epoch %v (attested again in epoch %v)", validator, epoch, epoch)
			pubkey, err := GetGetPubkeyForIndex(validator)
			if err != nil {
				return err
			}
			onlineValidators = append(onlineValidators, &indexPubkeyPair{Index: validator, Pubkey: pubkey})
		}
	}

	if len(offlineValidators) > 5000 {
		return fmt.Errorf("retrieved more than 5000 offline validators notifications: %v, exiting", len(offlineValidators))
	}

	if len(onlineValidators) > 5000 {
		return fmt.Errorf("retrieved more than 5000 online validators notifications: %v, exiting", len(onlineValidators))
	}

	_, subMap, err = db.GetSubsForEventFilter(types.ValidatorIsOfflineEventName)
	if err != nil {
		return fmt.Errorf("failed to get subs for %v: %v", types.ValidatorIsOfflineEventName, err)
	}

	for _, validator := range offlineValidators {
		t := hex.EncodeToString(validator.Pubkey)
		subs := subMap[t]
		for _, sub := range subs {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			logger.Infof("new event: validator %v detected as offline since epoch %v", validator.Index, epoch)

			n := validatorIsOfflineNotification{
				SubscriptionID: *sub.ID,
				ValidatorIndex: validator.Index,
				IsOffline:      true,
				EventEpoch:     epoch,
				EventName:      types.ValidatorIsOfflineEventName,
				InternalState:  fmt.Sprint(epoch - 2), // first epoch the validator stopped attesting
				EventFilter:    hex.EncodeToString(validator.Pubkey),
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
					break
				}
			}
			if isDuplicate {
				logger.Infof("duplicate offline notification detected")
				continue
			}
			notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], &n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	for _, validator := range onlineValidators {
		t := hex.EncodeToString(validator.Pubkey)
		subs := subMap[t]
		for _, sub := range subs {
			if sub.State.String == "" || sub.State.String == "-" { // discard online notifications that do not have a corresponding offline notification
				continue
			}

			originalLastSeenEpoch, err := strconv.ParseUint(sub.State.String, 10, 64)
			if err != nil {
				// i have no idea what just happened.
				return fmt.Errorf("this should never happen. couldn't parse state as uint64: %v", err)
			}

			epochsSinceOffline := epoch - originalLastSeenEpoch

			if epochsSinceOffline > epoch { // fix overflow
				epochsSinceOffline = 4
			}

			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}

			logger.Infof("new event: validator %v detected as online again at epoch %v", validator.Index, epoch)

			n := validatorIsOfflineNotification{
				SubscriptionID: *sub.ID,
				ValidatorIndex: validator.Index,
				IsOffline:      false,
				EventEpoch:     epoch,
				EventName:      types.ValidatorIsOfflineEventName,
				InternalState:  "-",
				EventFilter:    hex.EncodeToString(validator.Pubkey),
				EpochsOffline:  epochsSinceOffline,
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
					break
				}
			}
			if isDuplicate {
				logger.Infof("duplicate online notification detected")
				continue
			}
			notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], &n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

type validatorIsOfflineNotification struct {
	SubscriptionID  uint64
	ValidatorIndex  uint64
	EventEpoch      uint64
	EpochsOffline   uint64
	IsOffline       bool
	EventName       types.EventName
	EventFilter     string
	UnsubscribeHash sql.NullString
	InternalState   string
}

func (n *validatorIsOfflineNotification) GetLatestState() string {
	return n.InternalState
}

func (n *validatorIsOfflineNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorIsOfflineNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *validatorIsOfflineNotification) GetEpoch() uint64 {
	return n.EventEpoch
}

func (n *validatorIsOfflineNotification) GetInfo(includeUrl bool) string {
	if n.IsOffline {
		if includeUrl {
			return fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> is offline since epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>).`, n.ValidatorIndex, n.EventEpoch, utils.Config.Frontend.SiteDomain)
		} else {
			return fmt.Sprintf(`Validator %v is offline since epoch %v.`, n.ValidatorIndex, n.EventEpoch)
		}
	} else {
		if includeUrl {
			return fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> is back online since epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a> (was offline for %[4]v epoch(s)).`, n.ValidatorIndex, n.EventEpoch, utils.Config.Frontend.SiteDomain, n.EpochsOffline)
		} else {
			return fmt.Sprintf(`Validator %v is back online since epoch %v (was offline for %v epoch(s)).`, n.ValidatorIndex, n.EventEpoch, n.EpochsOffline)
		}
	}
}

func (n *validatorIsOfflineNotification) GetTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

func (n *validatorIsOfflineNotification) GetEventFilter() string {
	return n.EventFilter
}

func (n *validatorIsOfflineNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorIsOfflineNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorIsOfflineNotification) GetInfoMarkdown() string {
	if n.IsOffline {
		return fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) is offline since epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.EventEpoch, utils.Config.Frontend.SiteDomain)
	} else {
		return fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) is back online since epoch [%[2]v](https://%[3]v/epoch/%[2]v) (was offline for %[4]v epoch(s)).`, n.ValidatorIndex, n.EventEpoch, utils.Config.Frontend.SiteDomain, n.EpochsOffline)
	}
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

func (n *validatorAttestationNotification) GetLatestState() string {
	return ""
}

func (n *validatorAttestationNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorAttestationNotification) GetEventName() types.EventName {
	return n.EventName
}

func (n *validatorAttestationNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *validatorAttestationNotification) GetInfo(includeUrl bool) string {
	var generalPart = ""
	if includeUrl {
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> missed an attestation at slot <a href="https://%[3]v/slot/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
			//generalPart = fmt.Sprintf(`New scheduled attestation for Validator %v at slot %v.`, n.ValidatorIndex, n.Slot)
		case 1:
			generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> submitted a successful attestation for slot  <a href="https://%[3]v/slot/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
		}
		// return generalPart + getUrlPart(n.ValidatorIndex)
	} else {
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator %v missed an attestation at slot %v.`, n.ValidatorIndex, n.Slot)
			//generalPart = fmt.Sprintf(`New scheduled attestation for Validator %v at slot %v.`, n.ValidatorIndex, n.Slot)
		case 1:
			generalPart = fmt.Sprintf(`Validator %v submitted a successful attestation for slot %v.`, n.ValidatorIndex, n.Slot)
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

func (n *validatorAttestationNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorAttestationNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorAttestationNotification) GetInfoMarkdown() string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) missed an attestation at slot [%[2]v](https://%[3]v/slot/%[2]v).`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
	case 1:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) submitted a successful attestation for slot [%[2]v](https://%[3]v/slot/%[2]v).`, n.ValidatorIndex, n.Slot, utils.Config.Frontend.SiteDomain)
	}
	return generalPart
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

func (n *validatorGotSlashedNotification) GetLatestState() string {
	return ""
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
	generalPart := fmt.Sprintf(`Validator %v has been slashed at epoch %v by validator %v for %s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason)
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

func (n *validatorGotSlashedNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`Validator [%[1]v](https://%[5]v/validator/%[1]v) has been slashed at epoch [%[2]v](https://%[5]v/epoch/%[2]v) by validator [%[3]v](https://%[5]v/validator/%[3]v) for %[4]s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func collectValidatorGotSlashedNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {
	dbResult, err := db.GetValidatorsGotSlashed(epoch)
	if err != nil {
		return fmt.Errorf("error getting slashed validators from database, err: %w", err)
	}
	query := ""
	resultsLen := len(dbResult)
	for i, event := range dbResult {
		query += fmt.Sprintf(`SELECT %d as ref, id, user_id, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash from users_subscriptions where event_name = $1 AND event_filter = '%x'`, i, event.SlashedValidatorPubkey)
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
	if utils.Config.Chain.Config.ConfigName != "" {
		name = utils.Config.Chain.Config.ConfigName + ":" + name
	}
	err = db.FrontendWriterDB.Select(&subscribers, query, name)
	if err != nil {
		return fmt.Errorf("error querying subscribers, err: %w", err)
	}

	for _, sub := range subscribers {
		event := dbResult[sub.Ref]

		logger.Infof("creating %v notification for validator %v in epoch %v", event.SlashedValidatorPubkey, event.Reason, epoch)

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
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
	}

	return nil
}

type validatorWithdrawalNotification struct {
	SubscriptionID  uint64
	ValidatorIndex  uint64
	Epoch           uint64
	Slot            uint64
	Amount          uint64
	Address         []byte
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *validatorWithdrawalNotification) GetLatestState() string {
	return ""
}

func (n *validatorWithdrawalNotification) GetUnsubscribeHash() string {
	if n.UnsubscribeHash.Valid {
		return n.UnsubscribeHash.String
	}
	return ""
}

func (n *validatorWithdrawalNotification) GetEmailAttachment() *types.EmailAttachment {
	return nil
}

func (n *validatorWithdrawalNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorWithdrawalNotification) GetEpoch() uint64 {
	return n.Epoch
}

func (n *validatorWithdrawalNotification) GetEventName() types.EventName {
	return types.ValidatorReceivedWithdrawalEventName
}

func (n *validatorWithdrawalNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`A withdrawal of %v has been processed for validator %v.`, utils.FormatCurrentBalance(n.Amount, "ETH"), n.ValidatorIndex)
	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorWithdrawalNotification) GetTitle() string {
	return "Withdrawal Processed"
}

func (n *validatorWithdrawalNotification) GetEventFilter() string {
	return n.EventFilter
}

func (n *validatorWithdrawalNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`A withdrawal of %[2]v has been processed for validator [%[1]v](https://%[6]v/validator/%[1]v) during slot [%[3]v](https://%[6]v/slot/%[3]v). The funds have been sent to: [%[4]v](https://%[6]v/address/%[4]v).`, n.ValidatorIndex, utils.FormatCurrentBalance(n.Amount, "ETH"), n.Slot, utils.FormatHash(n.Address), n.Address, utils.Config.Frontend.SiteDomain)
	return generalPart
}

// collectWithdrawalNotifications collects all notifications validator withdrawals
func collectWithdrawalNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {

	// get all users that are subscribed to this event (scale: a few thousand rows depending on how many users we have)
	_, subMap, err := db.GetSubsForEventFilter(types.ValidatorReceivedWithdrawalEventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missed attestations %w", err)
	}

	// get all the withdrawal events for a specific epoch. Will be at most X per slot (currently 16 on mainnet, which is 32 * 16 per epoch; 512 rows).
	events, err := db.GetEpochWithdrawals(epoch)
	if err != nil {
		return fmt.Errorf("error getting withdrawals from database, err: %w", err)
	}

	// logger.Infof("retrieved %v events", len(events))
	for _, event := range events {
		subscribers, ok := subMap[hex.EncodeToString(event.Pubkey)]
		if ok {
			for _, sub := range subscribers {
				if sub.UserID == nil || sub.ID == nil {
					return fmt.Errorf("error expected userId or subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
				}
				if sub.LastEpoch != nil {
					lastSentEpoch := *sub.LastEpoch
					if lastSentEpoch >= epoch || epoch < sub.CreatedEpoch {
						continue
					}
				}
				// logger.Infof("creating %v notification for validator %v in epoch %v", types.ValidatorReceivedWithdrawalEventName, event.Index, epoch)
				n := &validatorWithdrawalNotification{
					SubscriptionID:  *sub.ID,
					ValidatorIndex:  event.Index,
					Epoch:           epoch,
					Slot:            event.Slot,
					Amount:          event.Amount,
					Address:         event.Address,
					EventFilter:     hex.EncodeToString(event.Pubkey),
					UnsubscribeHash: sub.UnsubscribeHash,
				}
				if _, exists := notificationsByUserID[*sub.UserID]; !exists {
					notificationsByUserID[*sub.UserID] = map[types.EventName][]types.Notification{}
				}
				if _, exists := notificationsByUserID[*sub.UserID][n.GetEventName()]; !exists {
					notificationsByUserID[*sub.UserID][n.GetEventName()] = []types.Notification{}
				}
				notificationsByUserID[*sub.UserID][n.GetEventName()] = append(notificationsByUserID[*sub.UserID][n.GetEventName()], n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
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

func (n *ethClientNotification) GetLatestState() string {
	return ""
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
		case "Teku":
			url = "https://github.com/ConsenSys/teku/releases"
		case "Prysm":
			url = "https://github.com/prysmaticlabs/prysm/releases"
		case "Nimbus":
			url = "https://github.com/status-im/nimbus-eth2/releases"
		case "Lighthouse":
			url = "https://github.com/sigp/lighthouse/releases"
		case "Erigon":
			url = "https://github.com/ledgerwatch/erigon/releases"
		case "Rocketpool":
			url = "https://github.com/rocket-pool/smartnode-install/releases"
		case "MEV-Boost":
			url = "https://github.com/flashbots/mev-boost/releases"
		case "Lodestar":
			url = "https://github.com/chainsafe/lodestar/releases"
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

func (n *ethClientNotification) GetInfoMarkdown() string {
	url := ""
	switch n.EthClient {
	case "Geth":
		url = "https://github.com/ethereum/go-ethereum/releases"
	case "Nethermind":
		url = "https://github.com/NethermindEth/nethermind/releases"
	case "Teku":
		url = "https://github.com/ConsenSys/teku/releases"
	case "Prysm":
		url = "https://github.com/prysmaticlabs/prysm/releases"
	case "Nimbus":
		url = "https://github.com/status-im/nimbus-eth2/releases"
	case "Lighthouse":
		url = "https://github.com/sigp/lighthouse/releases"
	case "Erigon":
		url = "https://github.com/ledgerwatch/erigon/releases"
	case "Rocketpool":
		url = "https://github.com/rocket-pool/smartnode-install/releases"
	case "MEV-Boost":
		url = "https://github.com/flashbots/mev-boost/releases"
	case "Lodestar":
		url = "https://github.com/chainsafe/lodestar/releases"
	default:
		url = "https://beaconcha.in/ethClients"
	}

	generalPart := fmt.Sprintf(`A new version for [%s](%s) is available.`, n.EthClient, url)

	return generalPart
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

		err := db.FrontendWriterDB.Select(&dbResult, `
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}
	return nil
}

type MachineEvents struct {
	SubscriptionID  uint64         `db:"id"`
	UserID          uint64         `db:"user_id"`
	MachineName     string         `db:"machine"`
	UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	EventThreshold  float64        `db:"event_threshold"`
}

func collectMonitoringMachineOffline(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {
	nowTs := time.Now().Unix()
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineOfflineEventName, 120,
		// notify condition
		func(_ *MachineEvents, machineData *types.MachineMetricSystemUser) bool {
			if machineData.CurrentDataInsertTs < nowTs-10*60 && machineData.CurrentDataInsertTs > nowTs-90*60 {
				return true
			}
			return false
		},
		epoch,
	)
}

func isMachineDataRecent(machineData *types.MachineMetricSystemUser) bool {
	nowTs := time.Now().Unix()
	return machineData.CurrentDataInsertTs >= nowTs-60*60
}

func collectMonitoringMachineDiskAlmostFull(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineDiskAlmostFullEventName, 750,
		// notify condition
		func(subscribeData *MachineEvents, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			percentFree := float64(machineData.CurrentData.DiskNodeBytesFree) / float64(machineData.CurrentData.DiskNodeBytesTotal+1)
			return percentFree < subscribeData.EventThreshold
		},
		epoch,
	)
}

func collectMonitoringMachineCPULoad(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineCpuLoadEventName, 10,
		// notify condition
		func(subscribeData *MachineEvents, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			if machineData.FiveMinuteOldData == nil { // no compare data found (5 min old data)
				return false
			}

			idle := float64(machineData.CurrentData.CpuNodeIdleSecondsTotal) - float64(machineData.FiveMinuteOldData.CpuNodeIdleSecondsTotal)
			total := float64(machineData.CurrentData.CpuNodeSystemSecondsTotal) - float64(machineData.FiveMinuteOldData.CpuNodeSystemSecondsTotal)
			percentLoad := float64(1) - (idle / total)

			return percentLoad > subscribeData.EventThreshold
		},
		epoch,
	)
}

func collectMonitoringMachineMemoryUsage(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineMemoryUsageEventName, 10,
		// notify condition
		func(subscribeData *MachineEvents, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			memFree := float64(machineData.CurrentData.MemoryNodeBytesFree) + float64(machineData.CurrentData.MemoryNodeBytesCached) + float64(machineData.CurrentData.MemoryNodeBytesBuffers)
			memTotal := float64(machineData.CurrentData.MemoryNodeBytesTotal)
			memUsage := float64(1) - (memFree / memTotal)

			return memUsage > subscribeData.EventThreshold
		},
		epoch,
	)
}

func collectMonitoringMachine(
	notificationsByUserID map[uint64]map[types.EventName][]types.Notification,
	eventName types.EventName,
	epochWaitInBetween int,
	notifyConditionFullfilled func(subscribeData *MachineEvents, machineData *types.MachineMetricSystemUser) bool,
	epoch uint64,
) error {

	var allSubscribed []MachineEvents
	err := db.FrontendWriterDB.Select(&allSubscribed,
		`SELECT 
			us.user_id,
			max(us.id) as id,
			ENCODE((array_agg(us.unsubscribe_hash))[1], 'hex') as unsubscribe_hash,
			event_filter as machine,
			COALESCE(event_threshold, 0) as event_threshold
		FROM users_subscriptions us 
		WHERE us.event_name = $1 AND us.created_epoch <= $2 
		AND (us.last_sent_epoch < ($2 - $3) OR us.last_sent_epoch IS NULL)
		group by us.user_id, machine, event_threshold`,
		eventName, epoch, epochWaitInBetween)
	if err != nil {
		return err
	}

	rowKeys := gcp_bigtable.RowList{}
	for _, data := range allSubscribed {
		rowKeys = append(rowKeys, db.GetMachineRowKey(data.UserID, "system", data.MachineName))
	}

	machineDataOfSubscribed, err := db.BigtableClient.GetMachineMetricsForNotifications(rowKeys)
	if err != nil {
		return err
	}

	var result []MachineEvents
	for _, data := range allSubscribed {
		machineMap, found := machineDataOfSubscribed[data.UserID]
		if !found {
			continue
		}
		currentMachineData, found := machineMap[data.MachineName]
		if !found {
			continue
		}

		//logrus.Infof("currentMachineData %v | %v | %v | %v", currentMachine.CurrentDataInsertTs, currentMachine.CompareDataInsertTs, currentMachine.UserID, currentMachine.Machine)
		if notifyConditionFullfilled(&data, currentMachineData) {
			result = append(result, data)
		}
	}

	for _, r := range result {

		n := &monitorMachineNotification{
			SubscriptionID:  r.SubscriptionID,
			MachineName:     r.MachineName,
			UserID:          r.UserID,
			EventName:       eventName,
			Epoch:           epoch,
			UnsubscribeHash: r.UnsubscribeHash,
		}
		//logrus.Infof("notify %v %v", eventName, n)
		if _, exists := notificationsByUserID[r.UserID]; !exists {
			notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
			notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
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

func (n *monitorMachineNotification) GetLatestState() string {
	return ""
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

func (n *monitorMachineNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
}

type taxReportNotification struct {
	SubscriptionID  uint64
	UserID          uint64
	Epoch           uint64
	EventFilter     string
	UnsubscribeHash sql.NullString
}

func (n *taxReportNotification) GetLatestState() string {
	return ""
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
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *taxReportNotification) GetTitle() string {
	return "Income Report"
}

func (n *taxReportNotification) GetEventFilter() string {
	return n.EventFilter
}

func (n *taxReportNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
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
		if utils.Config.Chain.Config.ConfigName != "" {
			name = utils.Config.Chain.Config.ConfigName + ":" + name
		}

		err := db.FrontendWriterDB.Select(&dbResult, `
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
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

func (n *networkNotification) GetLatestState() string {
	return ""
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
	return "Beaconchain Network Issues"
}

func (n *networkNotification) GetEventFilter() string {
	return n.EventFilter
}

func (n *networkNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`Network experienced finality issues ([view chart](https://%v/charts/network_liveness)).`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func collectNetworkNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	count := 0
	err := db.WriterDb.Get(&count, `
		select count(ts) from network_liveness where (headepoch-finalizedepoch) > 3 AND ts > now() - interval '60 minutes';
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

		err := db.FrontendWriterDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
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

func (n *rocketpoolNotification) GetLatestState() string {
	return ""
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
		return `A new reward round has started. You can now claim your rewards from the previous round.`
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

		return fmt.Sprintf(`Your validator %v has been elected to be part of the next sync committee. The additional duties start at epoch %v, which is in %s and will last for about a day until epoch %v.`, extras[0], extras[1], inTime.Round(time.Second), extras[2])
	}

	return ""
}

func (n *rocketpoolNotification) GetTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return `Rocketpool Commission`
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `Rocketpool Claim Available`
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

func (n *rocketpoolNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
}

func collectRocketpoolComissionNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	fee := 0.0
	err := db.WriterDb.Get(&fee, `
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

		err := db.FrontendWriterDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

func collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName) error {
	var ts int64
	err := db.WriterDb.Get(&ts, `
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

		err := db.FrontendWriterDB.Select(&dbResult, `
			SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
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
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

func collectRocketpoolRPLCollateralNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName, epoch uint64) error {

	pubkeys, subMap, err := db.GetSubsForEventFilter(eventName)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for RocketpoolRPLCollateral %w", err)
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

		err = db.WriterDb.Select(&partial, `
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

		if sub.LastEpoch != nil {
			lastSentEpoch := *sub.LastEpoch
			if lastSentEpoch >= epoch-80 || epoch < sub.CreatedEpoch {
				continue
			}
		}

		n := &rocketpoolNotification{
			SubscriptionID:  *sub.ID,
			UserID:          *sub.UserID,
			Epoch:           epoch,
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
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
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
		return errors.New("can not cast nil to BigFloat")
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
		return fmt.Errorf("could not scan type %T into BigFloat", t)
	}

	return nil
}

func (b *BigFloat) bigFloat() *big.Float {
	return (*big.Float)(b)
}
func bigFloat(x float64) *big.Float {
	return new(big.Float).SetFloat64(x)
}

func collectSyncCommittee(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, eventName types.EventName, epoch uint64) error {

	slotsPerSyncCommittee := utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod * utils.Config.Chain.Config.SlotsPerEpoch
	currentPeriod := epoch * utils.Config.Chain.Config.SlotsPerEpoch / slotsPerSyncCommittee
	nextPeriod := currentPeriod + 1

	var validators []struct {
		PubKey string `db:"pubkey"`
		Index  uint64 `db:"validatorindex"`
	}
	err := db.WriterDb.Select(&validators, `SELECT encode(pubkey, 'hex') as pubkey, validators.validatorindex FROM sync_committees LEFT JOIN validators ON validators.validatorindex = sync_committees.validatorindex WHERE period = $1`, nextPeriod)

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
		EventFilter     string         `db:"event_filter"`
		UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	}

	err = db.FrontendWriterDB.Select(&dbResult, `
				SELECT us.id, us.user_id, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') as unsubscribe_hash
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
			Epoch:           epoch,
			EventFilter:     r.EventFilter,
			EventName:       eventName,
			ExtraData:       fmt.Sprintf("%v|%v|%v", mapping[r.EventFilter], nextPeriod*utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod, (nextPeriod+1)*utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod),
			UnsubscribeHash: r.UnsubscribeHash,
		}
		if _, exists := notificationsByUserID[r.UserID]; !exists {
			notificationsByUserID[r.UserID] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByUserID[r.UserID][n.GetEventName()]; !exists {
			notificationsByUserID[r.UserID][n.GetEventName()] = []types.Notification{}
		}
		notificationsByUserID[r.UserID][n.GetEventName()] = append(notificationsByUserID[r.UserID][n.GetEventName()], n)
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
	}

	return nil
}

type WebhookQueue struct {
	NotificationID uint64         `db:"id"`
	Url            string         `db:"url"`
	Retries        uint64         `db:"retries"`
	LastSent       time.Time      `db:"last_retry"`
	Destination    sql.NullString `db:"destination"`
	Payload        []byte         `db:"payload"`
	LastTry        time.Time      `db:"last_try"`
}
