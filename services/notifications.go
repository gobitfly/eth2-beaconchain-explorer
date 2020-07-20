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
	err := updateValidatorBalanceDecreasedSubscriptions()
	if err != nil {
		return err
	}
	return nil
}

func collectNotifications() {
	// reset notifications
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
		}(userID, userNotifications)
	}
}

type validatorBalanceDecreasedNotification struct {
	ValidatorIndex     uint64
	ValidatorPublicKey string
	Epoch              uint64
	OldBalance         uint64
	NewBalance         uint64
}

func (n validatorBalanceDecreasedNotification) EventName() types.EventName {
	return types.ValidatorBalanceDecreasedEventName
}

func (n validatorBalanceDecreasedNotification) Info() string {
	return fmt.Sprintf(`the balance of validator %v decreased from %v to %v at epoch %v`, n.ValidatorIndex, n.Epoch, n.OldBalance, n.NewBalance)
}

var validatorBalanceDecreasedSubscriptions = map[string][]uint64{}
var validatorBalances = map[string]uint64{}

func updateValidatorBalanceDecreasedSubscriptions() error {
	now := time.Now()
	validatorBalanceDecreasedSubscriptions = map[string][]uint64{}

	filter := db.GetSubscriptionsFilter{
		EventNames: &[]types.EventName{types.ValidatorBalanceDecreasedEventName},
	}

	subs, err := db.GetSubscriptions(filter)
	if err != nil {
		return err
	}
	for _, s := range subs {
		if s.ValidatorPublicKey == nil {
			continue
		}
		// if we already sent a notification for this validator skip it
		if s.LastNotification != nil && (*s.LastNotification).Add(notificationRateLimit).Before(now) {
			continue
		}
		_, exists := validatorBalanceDecreasedSubscriptions[*s.ValidatorPublicKey]
		if !exists {
			validatorBalanceDecreasedSubscriptions[*s.ValidatorPublicKey] = []uint64{s.UserID}
		} else {
			validatorBalanceDecreasedSubscriptions[*s.ValidatorPublicKey] = append(validatorBalanceDecreasedSubscriptions[*s.ValidatorPublicKey], s.UserID)
		}
	}
	return nil
}

var collectValidatorBalanceDecreasedNotificationsLastEpoch = uint64(0)

func collectValidatorBalanceDecreasedNotifications() error {
	latestEpoch := LatestEpoch()

	// only check if there is a new epoch
	if latestEpoch == 0 || latestEpoch == collectValidatorBalanceDecreasedNotificationsLastEpoch {
		return nil
	}
	collectValidatorBalanceDecreasedNotificationsLastEpoch = latestEpoch

	newValidatorBalances := []struct {
		Index     uint64 `db:"validatorindex"`
		PublicKey string `db:"pubkey"`
		Balance   uint64 `db:"balance"`
	}{}

	err := db.DB.Select(&validatorBalances, `
		SELECT vb.validatorindex, vb.balance, encode(v.pubkey, 'hex')
		FROM validator_balances vb
		INNER JOIN validators v ON v.validatorindex = vb.validatorindex
		WHERE epoch = $1`, latestEpoch)
	if err != nil {
		return err
	}

	for _, v := range newValidatorBalances {
		oldValidatorBalance, exists := validatorBalances[v.PublicKey]
		if !exists {
			validatorBalances[v.PublicKey] = v.Balance
		} else {
			if oldValidatorBalance <= v.Balance {
				continue
			}
			if _, exists = validatorBalanceDecreasedSubscriptions[v.PublicKey]; !exists {
				continue
			}
			if len(validatorBalanceDecreasedSubscriptions[v.PublicKey]) == 0 {
				continue
			}
			n := validatorBalanceDecreasedNotification{
				ValidatorIndex:     v.Index,
				ValidatorPublicKey: v.PublicKey,
				Epoch:              latestEpoch,
				OldBalance:         oldValidatorBalance,
				NewBalance:         v.Balance,
			}
			for _, userID := range validatorBalanceDecreasedSubscriptions[v.PublicKey] {
				_, exists := notifications[userID]
				if !exists {
					notifications[userID] = map[types.EventName][]types.Notification{}
				}
				_, exists = notifications[userID][types.ValidatorBalanceDecreasedEventName]
				if !exists {
					notifications[userID][types.ValidatorBalanceDecreasedEventName] = []types.Notification{n}
				} else {
					notifications[userID][types.ValidatorBalanceDecreasedEventName] = append(notifications[userID][types.ValidatorBalanceDecreasedEventName], n)
				}
			}
		}
	}

	return nil
}
