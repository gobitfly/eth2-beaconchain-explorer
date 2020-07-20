package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"
)

var notificationRateLimit = time.Second * 60 * 10

var notificationsByEmail = map[string]map[types.EventName][]types.Notification{}

func notificationsSender() {
	for {
		start := time.Now()
		err := collectNotifications()
		if err != nil {
			logger.Errorf("error collecting notifications: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		sendNotifications()
		logger.WithField("duration", time.Since(start)).Info("notifications completed")
		time.Sleep(time.Second * 60)
	}
}

func collectNotifications() error {
	notificationsByEmail = map[string]map[types.EventName][]types.Notification{}
	err := collectValidatorBalanceDecreasedNotifications()
	if err != nil {
		return err
	}
	return nil
}

func sendNotifications() {
	for userEmail, userNotifications := range notificationsByEmail {
		go func(userEmail string, userNotifications map[types.EventName][]types.Notification) {
			sentSubIDs := []uint64{}
			subject := "beaconcha.in: Notification"
			msg := ""
			for event, ns := range userNotifications {
				if len(msg) > 0 {
					msg += "\n"
				}
				msg += fmt.Sprintf("%s\n====\n\n", event)
				for _, n := range ns {
					msg += fmt.Sprintf("%s\n", n.GetInfo())
					sentSubIDs = append(sentSubIDs, n.GetSubscriptionID())
				}
			}
			err := utils.SendMail(userEmail, subject, msg)
			if err != nil {
				logger.Errorf("error sending notification-email: %v", err)
				return
			}

			err = db.UpdateSubscriptionsSent(sentSubIDs, time.Now())
			if err != nil {
				logger.Errorf("error updating sent-time of sent notifications: %v", err)
				return
			}
		}(userEmail, userNotifications)
	}
}

type validatorBalanceDecreasedNotification struct {
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	PrevBalance        uint64
	Balance            uint64
	SubscriptionID     uint64
}

func (n *validatorBalanceDecreasedNotification) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *validatorBalanceDecreasedNotification) GetEventName() types.EventName {
	return types.ValidatorBalanceDecreasedEventName
}

func (n *validatorBalanceDecreasedNotification) GetInfo() string {
	balance := utils.RoundDecimals(float64(n.Balance)/1e9, 9)
	prevBalance := utils.RoundDecimals(float64(n.Balance)/1e9, 9)
	return fmt.Sprintf(`The balance of validator %v decreased from %v ETH to %v ETH at epoch %v.`, n.ValidatorIndex, prevBalance, balance, n.Epoch)
}

func collectValidatorBalanceDecreasedNotifications() error {
	latestEpoch := LatestEpoch()
	// only check if there is a new epoch
	if latestEpoch == 0 {
		return nil
	}
	prevEpoch := latestEpoch - 1
	dbResult := []struct {
		SubscriptionID uint64 `db:"id"`
		Email          string `db:"email"`
		ValidatorIndex uint64 `db:"validatorindex"`
		Balance        uint64 `db:"balance"`
		PrevBalance    uint64 `db:"prevbalance"`
	}{}
	err := db.DB.Select(&dbResult, `
		WITH
			decreased_balance_validators AS (
				SELECT 
					vb.validatorindex, 
					ENCODE(v.pubkey, 'hex') AS pubkey,
					vb.balance, 
					vb2.balance AS prevbalance
				FROM validator_balances vb
				INNER JOIN validators v ON v.validatorindex = vb.validatorindex
				INNER JOIN validator_balances vb2 ON vb.validatorindex = vb2.validatorindex AND vb2.epoch = $2
				WHERE vb.epoch = $1 AND (vb.balance - vb2.balance) < 0
			)
		SELECT us.id, u.email, dbv.validatorindex, dbv.balance, dbv.prevbalance
		FROM users_subscriptions us
			INNER JOIN users u ON u.id = us.user_id
			INNER JOIN decreased_balance_validators dbv ON dbv.pubkey = us.event_filter
		WHERE (us.sent_ts IS NULL OR us.sent_ts < TO_TIMESTAMP($3))`,
		latestEpoch, prevEpoch, time.Now().Add(-notificationRateLimit).Unix())
	if err != nil {
		return err
	}
	for _, r := range dbResult {
		n := &validatorBalanceDecreasedNotification{
			SubscriptionID: r.SubscriptionID,
			ValidatorIndex: r.ValidatorIndex,
			Balance:        r.Balance,
			PrevBalance:    r.PrevBalance,
			Epoch:          latestEpoch,
		}

		if _, exists := notificationsByEmail[r.Email]; !exists {
			notificationsByEmail[r.Email] = map[types.EventName][]types.Notification{}
		}
		if _, exists := notificationsByEmail[r.Email][n.GetEventName()]; !exists {
			notificationsByEmail[r.Email][n.GetEventName()] = []types.Notification{}
		}
		notificationsByEmail[r.Email][n.GetEventName()] = append(notificationsByEmail[r.Email][n.GetEventName()], n)
	}
	return nil
}
