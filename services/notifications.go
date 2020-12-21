package services

import (
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/notify"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
)

func notificationsSender() {
	for {
		// check if the explorer is not too far behind, if we set this value to close (10m) it could potentially never send any notifications
		// if IsSyncing() {
		if time.Now().Add(time.Minute * -20).After(utils.EpochToTime(LatestEpoch())) {
			logger.Info("skipping notifications because the explorer is syncing")
			time.Sleep(time.Second * 60)
			continue
		}
		start := time.Now()
		notifications := collectNotifications()
		sendNotifications(notifications)
		logger.WithField("notifications", len(notifications)).WithField("duration", time.Since(start)).Info("notifications completed")
		time.Sleep(time.Second * 60)
	}
}

func collectNotifications() map[uint64]map[types.EventName][]types.Notification {
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	var err error
	err = collectValidatorBalanceDecreasedNotifications(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting validator_balance_decreased notifications: %v", err)
	}
	err = collectValidatorGotSlashedNotifications(notificationsByUserID)
	if err != nil {
		logger.Errorf("error collecting validator_got_slashed notifications: %v", err)
	}

	// executed Proposals
	err = collectBlockProposalNotifications(notificationsByUserID, 1, types.ValidatorExecutedProposalEventName)
	if err != nil {
		logger.Errorf("error collecting validator_proposal_submitted notifications: %v", err)
	}

	// Missed proposals
	err = collectBlockProposalNotifications(notificationsByUserID, 2, types.ValidatorMissedProposalEventName)
	if err != nil {
		logger.Errorf("error collecting validator_proposal_missed notifications: %v", err)
	}

	return notificationsByUserID
}

func sendNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) {
	sendEmailNotifications(notificationsByUserID)
	sendPushNotifications(notificationsByUserID)
	// sendWebhookNotifications(notificationsByUserID)
}

func sendPushNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}

	tokensByUserID, err := db.GetUserPushTokenByIds(userIDs)
	if err != nil {
		logger.Errorf("error when sending push-notificaitons: could not get tokens: %v", err)
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
						notification.Title = n.GetTitle()
						notification.Body = n.GetInfo(false)

						message := new(messaging.Message)
						message.Notification = notification
						message.Token = userToken

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
				return
			}

			for epoch, subIDs := range sentSubsByEpoch {
				err = db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch)
				if err != nil {
					logger.Errorf("error updating sent-time of sent notifications: %v", err)
				}
			}
		}(userTokens, userNotifications)
	}

}

func sendEmailNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) {
	userIDs := []uint64{}
	for userID := range notificationsByUserID {
		userIDs = append(userIDs, userID)
	}
	emailsByUserID, err := db.GetUserEmailsByIds(userIDs)
	if err != nil {
		logger.Errorf("error when sending eamil-notificaitons: could not get emails: %v", err)
		return
	}

	for userID, userNotifications := range notificationsByUserID {
		userEmail, exists := emailsByUserID[userID]
		if !exists {
			logger.Errorf("error when sending email-notification: could not find email for user %v", userID)
			continue
		}
		go func(userEmail string, userNotifications map[types.EventName][]types.Notification) {
			sentSubsByEpoch := map[uint64][]uint64{}
			subject := fmt.Sprintf("%s: Notification", utils.Config.Frontend.SiteDomain)
			msg := ""
			for event, ns := range userNotifications {
				if len(msg) > 0 {
					msg += "\n"
				}
				msg += fmt.Sprintf("%s\n====\n\n", event)
				for _, n := range ns {
					msg += fmt.Sprintf("%s\n", n.GetInfo(true))
					e := n.GetEpoch()
					if _, exists := sentSubsByEpoch[e]; !exists {
						sentSubsByEpoch[e] = []uint64{n.GetSubscriptionID()}
					} else {
						sentSubsByEpoch[e] = append(sentSubsByEpoch[e], n.GetSubscriptionID())
					}
				}
				if event == "validator_balance_decreased" {
					msg += "\nYou will not receive any further balance decrease mails for these validators until the balance of a validator is increasing again.\n"
				}
			}
			msg += fmt.Sprintf("\nBest regards\n\n%s", utils.Config.Frontend.SiteDomain)

			err := mail.SendMailRateLimited(userEmail, subject, msg)
			if err != nil {
				logger.Errorf("error sending notification-email: %v", err)
				return
			}

			for epoch, subIDs := range sentSubsByEpoch {
				err = db.UpdateSubscriptionsLastSent(subIDs, time.Now(), epoch)
				if err != nil {
					logger.Errorf("error updating sent-time of sent notifications: %v", err)
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

func getUrlPart(validatorIndex uint64) string {
	return fmt.Sprintf(` For more information visit: https://%[2]s/validator/%[1]v.`, validatorIndex, utils.Config.Frontend.SiteDomain)
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

	var dbResult []struct {
		SubscriptionID uint64 `db:"id"`
		UserID         uint64 `db:"user_id"`
		ValidatorIndex uint64 `db:"validatorindex"`
		StartBalance   uint64 `db:"startbalance"`
		EndBalance     uint64 `db:"endbalance"`
	}

	err := db.DB.Select(&dbResult, `
		SELECT id, user_id, validatorindex, startbalance, endbalance FROM (
			SELECT 
				us.id, 
				us.user_id, 
				v.validatorindex, 
				vb0.balance AS endbalance, 
				vb3.balance AS startbalance, 
				us.last_sent_epoch,
				(SELECT MAX(epoch) FROM (
					SELECT epoch, balance-LAG(balance) OVER (ORDER BY epoch) AS diff
					FROM validator_balances 
					WHERE validatorindex = v.validatorindex AND epoch > us.last_sent_epoch AND epoch > $2 - 10
				) b WHERE diff > 0) AS lastbalanceincreaseepoch
			FROM users_subscriptions us
			INNER JOIN validators v ON ENCODE(v.pubkey, 'hex') = us.event_filter
			INNER JOIN validator_balances vb0 ON v.validatorindex = vb0.validatorindex AND vb0.epoch = $2
			INNER JOIN validator_balances vb1 ON v.validatorindex = vb1.validatorindex AND vb1.epoch = $2 - 1 AND vb1.balance > vb0.balance
			INNER JOIN validator_balances vb2 ON v.validatorindex = vb2.validatorindex AND vb2.epoch = $2 - 2 AND vb2.balance > vb1.balance
			INNER JOIN validator_balances vb3 ON v.validatorindex = vb3.validatorindex AND vb3.epoch = $2 - 3 AND vb3.balance > vb2.balance
			WHERE us.event_name = $1 AND us.created_epoch <= $2
		) a WHERE lastbalanceincreaseepoch IS NOT NULL OR last_sent_epoch IS NULL`,
		types.ValidatorBalanceDecreasedEventName, latestEpoch)
	if err != nil {
		return err
	}

	for _, r := range dbResult {
		n := &validatorBalanceDecreasedNotification{
			SubscriptionID: r.SubscriptionID,
			ValidatorIndex: r.ValidatorIndex,
			StartEpoch:     latestEpoch - 3,
			EndEpoch:       latestEpoch,
			StartBalance:   r.StartBalance,
			EndBalance:     r.EndBalance,
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

func collectBlockProposalNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification, status uint64, eventName types.EventName) error {
	latestEpoch := LatestEpoch()

	var dbResult []struct {
		SubscriptionID uint64 `db:"id"`
		UserID         uint64 `db:"user_id"`
		ValidatorIndex uint64 `db:"validatorindex"`
		Epoch          uint64 `db:"epoch"`
		Status         uint64 `db:"status"`
	}

	err := db.DB.Select(&dbResult, `
			SELECT 
				us.id, 
				us.user_id, 
				v.validatorindex, 
				pa.epoch,
				pa.status
			FROM users_subscriptions us
			INNER JOIN validators v ON ENCODE(v.pubkey, 'hex') = us.event_filter
			INNER JOIN proposal_assignments pa ON v.validatorindex = pa.validatorindex AND pa.epoch >= ($2 - 5) 
			WHERE us.event_name = $1 AND pa.status = $3 AND us.created_epoch <= $2 AND pa.epoch >= ($2 - 5) AND (us.last_sent_epoch < pa.epoch OR us.last_sent_epoch IS NULL)`,
		eventName, latestEpoch, status)
	if err != nil {
		return err
	}

	for _, r := range dbResult {

		n := &validatorProposalNotification{
			SubscriptionID: r.SubscriptionID,
			ValidatorIndex: r.ValidatorIndex,
			Epoch:          r.Epoch,
			Status:         r.Status,
			EventName:      eventName,
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

type validatorProposalNotification struct {
	SubscriptionID     uint64
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	Status             uint64 // * Can be 0 = scheduled, 1 executed, 2 missed */
	EventName          types.EventName
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

type validatorGotSlashedNotification struct {
	SubscriptionID uint64
	ValidatorIndex uint64
	Epoch          uint64
	Slasher        uint64
	Reason         string
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

func collectValidatorGotSlashedNotifications(notificationsByUserID map[uint64]map[types.EventName][]types.Notification) error {
	latestEpoch := LatestEpoch()
	if latestEpoch == 0 {
		return nil
	}

	var dbResult []struct {
		SubscriptionID uint64 `db:"id"`
		UserID         uint64 `db:"user_id"`
		ValidatorIndex uint64 `db:"validatorindex"`
		Slasher        uint64 `db:"slasher"`
		Epoch          uint64 `db:"epoch"`
		Reason         string `db:"reason"`
	}

	err := db.DB.Select(&dbResult, `
		WITH
			slashings AS (
				SELECT DISTINCT ON (slashedvalidator) * FROM (
					SELECT
						blocks.slot, 
						blocks.epoch, 
						blocks.proposer AS slasher, 
						UNNEST(ARRAY(
							SELECT UNNEST(attestation1_indices)
								INTERSECT
							SELECT UNNEST(attestation2_indices)
						)) AS slashedvalidator, 
						'Attestation Violation' AS reason
					FROM blocks_attesterslashings 
					LEFT JOIN blocks ON blocks_attesterslashings.block_slot = blocks.slot
					WHERE blocks.status = '1'
					UNION ALL
						SELECT
							blocks.slot, 
							blocks.epoch, 
							blocks.proposer AS slasher, 
							blocks_proposerslashings.proposerindex AS slashedvalidator,
							'Proposer Violation' AS reason 
						FROM blocks_proposerslashings
						LEFT JOIN blocks ON blocks_proposerslashings.block_slot = blocks.slot
						WHERE blocks.status = '1'
				) a
				ORDER BY slashedvalidator, slot
			)
		SELECT us.id, us.user_id, v.validatorindex, s.slasher, s.epoch, s.reason
		FROM users_subscriptions us
		INNER JOIN validators v ON ENCODE(v.pubkey, 'hex') = us.event_filter
		INNER JOIN slashings s ON s.slashedvalidator = v.validatorindex
		WHERE us.event_name = $1 AND us.last_sent_epoch IS NULL AND us.created_epoch < s.epoch`,
		types.ValidatorGotSlashedEventName)
	if err != nil {
		return err
	}

	for _, r := range dbResult {
		n := &validatorGotSlashedNotification{
			SubscriptionID: r.SubscriptionID,
			ValidatorIndex: r.ValidatorIndex,
			Slasher:        r.Slasher,
			Epoch:          r.Epoch,
			Reason:         r.Reason,
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
