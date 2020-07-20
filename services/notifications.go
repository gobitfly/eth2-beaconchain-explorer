package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"
)

var notificationRateLimit = time.Second * 60 * 10

var notifications = map[uint64]map[types.EventName][]types.Notification{}
var subscriptions = map[types.EventName]map[string][]*types.Subscription{}
var subscriptionsByUser = map[uint64][]*types.Subscription{}

func notificationsSender() {
	for {
		start := time.Now()
		var err error
		err = updateSubscriptions()
		if err != nil {
			logger.Errorf("error updating subscriptions: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}
		collectNotifications()
		sendNotifications()
		logger.WithField("duration", time.Since(start)).Info("notifications completed")
		time.Sleep(time.Second * 60)
	}
}

func updateSubscriptions() error {
	now := time.Now()

	subscriptions = map[types.EventName]map[string][]*types.Subscription{}
	subscriptionsByUser = map[uint64][]*types.Subscription{}

	subs, err := db.GetSubscriptions(db.GetSubscriptionsFilter{})
	if err != nil {
		return err
	}
	for _, s := range subs {
		// if we already sent a notification for this subscription skip it
		if s.Sent != nil && (*s.Sent).Add(notificationRateLimit).Before(now) {
			continue
		}
		if _, exists := subscriptions[s.EventName]; !exists {
			subscriptions[s.EventName] = map[string][]*types.Subscription{}
		}
		if _, exists := subscriptions[s.EventName][s.EventFilter]; !exists {
			subscriptions[s.EventName][s.EventFilter] = []*types.Subscription{s}
		} else {
			subscriptions[s.EventName][s.EventFilter] = append(subscriptions[s.EventName][s.EventFilter], s)
		}
	}
	return nil
}

func collectNotifications() {
	notifications = map[uint64]map[types.EventName][]types.Notification{}
	collectValidatorBalanceDecreasedNotifications()
}

func sendNotifications() {
	for userID, userNotifications := range notifications {
		go func(userID uint64, userNotifications map[types.EventName][]types.Notification) {
			email, err := db.GetUserEmailById(userID)
			if err != nil {
				logger.Errorf("error getting email of user: %v", err)
				return
			}
			subject := "beaconcha.in: Notification"
			msg := ""
			for event, ns := range userNotifications {
				if len(msg) > 0 {
					msg += "\n"
				}
				msg += fmt.Sprintf("%s\n====\n\n", event)
				for _, n := range ns {
					msg += fmt.Sprintf("%s\n", n.Info())
				}
			}
			err = utils.SendMail(email, subject, msg)
			if err != nil {
				logger.Errorf("error sending notification-email: %v", err)
				return
			}

			if _, exists := subscriptionsByUser[userID]; !exists {
				return
			}

			sentNotificationIDs := []uint64{}
			for _, s := range subscriptionsByUser[userID] {
				sentNotificationIDs = append(sentNotificationIDs, s.ID)
			}
			err = db.UpdateSubscriptionsSent(sentNotificationIDs, time.Now())
			if err != nil {
				logger.Errorf("error updating sent-time of sent notifications: %v", err)
				return
			}
		}(userID, userNotifications)
	}
}

type validatorBalanceDecreasedNotification struct {
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	PrevBalance        uint64
	Balance            uint64
}

func (n validatorBalanceDecreasedNotification) EventName() types.EventName {
	return types.ValidatorBalanceDecreasedEventName
}

func (n validatorBalanceDecreasedNotification) Info() string {
	balance := utils.RoundDecimals(float64(n.Balance)/1e9, 9)
	prevBalance := utils.RoundDecimals(float64(n.Balance)/1e9, 9)
	return fmt.Sprintf(`The balance of validator %v decreased from %v ETH to %v ETH at epoch %v.`, n.ValidatorIndex, prevBalance, balance, n.Epoch)
}

func collectValidatorBalanceDecreasedNotifications() error {
	subs, exists := subscriptions[types.ValidatorBalanceDecreasedEventName]
	if !exists {
		return nil
	}

	latestEpoch := LatestEpoch()

	// only check if there is a new epoch
	if latestEpoch == 0 {
		return nil
	}

	prevEpoch := latestEpoch - 1

	validators := []struct {
		Index       uint64 `db:"validatorindex"`
		PublicKey   string `db:"pubkey"`
		Balance     uint64 `db:"balance"`
		PrevBalance uint64 `db:"balance_prev"`
		BalanceDiff int64  `db:"balance_diff"`
	}{}

	err := db.DB.Select(&validators, `
		SELECT 
			vb.validatorindex, 
			encode(v.pubkey, 'hex') as pubkey,
			vb.balance, 
			vb2.balance as balance_prev,
			vb.balance - vb2.balance as balance_diff
		FROM validator_balances vb
		INNER JOIN validators v ON v.validatorindex = vb.validatorindex
		INNER JOIN validator_balances vb2 ON vb.validatorindex = vb2.validatorindex AND vb2.epoch = $2
		WHERE vb.epoch = $1 AND (vb.balance - vb2.balance) < 0`, latestEpoch, prevEpoch)
	if err != nil {
		return err
	}

	for _, v := range validators {
		if _, exists := subs[v.PublicKey]; !exists {
			continue
		}
		if len(subs[v.PublicKey]) == 0 {
			continue
		}
		n := &validatorBalanceDecreasedNotification{
			ValidatorIndex:     v.Index,
			ValidatorPublicKey: v.PublicKey,
			Epoch:              latestEpoch,
			Balance:            v.Balance,
			PrevBalance:        v.PrevBalance,
		}
		for _, s := range subs[v.PublicKey] {
			if _, exists = notifications[s.UserID]; !exists {
				notifications[s.UserID] = map[types.EventName][]types.Notification{}
			}
			_, exists = notifications[s.UserID][types.ValidatorBalanceDecreasedEventName]
			if !exists {
				notifications[s.UserID][types.ValidatorBalanceDecreasedEventName] = []types.Notification{n}
			} else {
				notifications[s.UserID][types.ValidatorBalanceDecreasedEventName] = append(notifications[s.UserID][types.ValidatorBalanceDecreasedEventName], n)
			}

			if _, exists = subscriptionsByUser[s.UserID]; !exists {
				subscriptionsByUser[s.UserID] = []*types.Subscription{s}
			} else {
				subscriptionsByUser[s.UserID] = append(subscriptionsByUser[s.UserID], s)
			}
		}
	}

	return nil
}
