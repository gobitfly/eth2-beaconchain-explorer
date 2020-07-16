package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"
)

var notifications = map[int64]map[types.EventName][]types.Notification{}

func notificationsSender() {
	for {
		var err error
		err = updateSubscriptions()
		if err != nil {
			logger.Errorf("error updating subscriptions: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}
		collectNofitications()
		sendNotifications()
		time.Sleep(time.Second * 60)
	}
}

func updateSubscriptions() error {
	updateValidatorBalanceDecreasedSubscriptions()
}

func collectNotifcations() {
	notifications = map[int64]map[types.EventName]types.Notification{}
	collectValidatorBalanceDecreasedNotifications()
}

func sendNotifications() {
	for userID, userNotifications := range notifications {
		go func(userID, userNotifications) {
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
	ValidatorIndex uint64
	Epoch          uint64
	OldBalance     uint64
	NewBalance     uint64
}

func (n *validatorBalanceDecreasedNotification) EventName() types.EventName {
	return types.ValidatorBalanceDecreasedEventName
}

func (n *validatorBalanceDecreasedNotification) Info() string {
	return fmt.Sprintf(`the balance of validator %v decreased from %v to %v at epoch %v`, n.ValidatorIndex, n.Epoch, n.OldBalance, n.NewBalance)
}

var validatorBalanceDecreasedSubscriptions = map[uint64][]int64{}
var validatorBalances = map[uint64]uint64{}

func updateValidatorBalanceDecreasedSubscriptions() error {
	validatorBalanceDecreasedSubscriptions = map[uint64][]int64{}

	subs, err := db.GetSubscriptions(ValidatorBalanceDecreased)
	if err != nil {
		return err
	}
	for _, s := range subs {
		if s.ValidatorIndex == nil {
			continue
		}
		_, exists := validatorBalanceDecreasedSubscriptions[*s.ValidatorIndex]
		if !exists {
			validatorBalanceDecreasedSubscriptions[*s.ValidatorIndex] = []int64{s.UserID}
		} else {
			validatorBalanceDecreasedSubscriptions[*s.ValidatorIndex] = append(validatorBalanceDecreasedSubscriptions[*s.ValidatorIndex], s.UserID)
		}
	}
	return nil
}

func collectValidatorBalanceDecreasedNotifications() error {
	latestEpoch := LatestEpoch()

	newValidatorBalances := []struct {
		Index   uint64 `db:"validatorindex"`
		Balance uint64 `db:"balance"`
	}{}

	err := db.DB.Select(&validatorBalances, "SELECT validatorindex, balance FROM validator_balances WHERE epoch = $1", latestEpoch)
	if err != nil {
		return err
	}

	for _, v := range newValidatorBalances {
		oldValidatorBalance, exists := validatorBalances[v.Index]
		if !exists {
			validatorBalances[v.Index] = v.Balance
		} else {
			if oldValidatorBalance <= v.Balance {
				continue
			}
			if _, exists = validatorBalanceDecreasedSubscriptions[v.Index]; !exists {
				continue
			}
			if len(validatorBalanceDecreasedSubscriptions[v.Index]) == 0 {
				continue
			}
			n = validatorBalanceDecreasedNotification{
				ValidatorIndex: v.Index,
				Epoch:          latestEpoch,
				OldBalance:     oldValidatorBalance,
				NewBalance:     v.Balance,
			}
			for _, userID := range validatorBalanceDecreasedSubscriptions[v.Index] {
				_, exists := notifications[userID]
				if !exists {
					notifications[userID] = map[types.EventName]types.Notification{}
				}
				_, exists = notifications[userID][types.ValidatorBalanceDecreasedEventName]
				if !exists {
					notifications[userID][types.ValidatorBalanceDecreasedEventName] = []types.Notification{}
				}
			}
		}
	}

	return nil
}
