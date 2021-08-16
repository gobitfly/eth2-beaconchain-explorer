package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	configPath := flag.String("config", "config.yml", "Path to the config file")
	flag.Parse()
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	logger.Infof("connected to db:          %+v", cfg.Database)

	db.MustInitFrontendDB(cfg.Frontend.Database.Username, cfg.Frontend.Database.Password, cfg.Frontend.Database.Host, cfg.Frontend.Database.Port, cfg.Frontend.Database.Name, cfg.Frontend.SessionSecret)
	defer db.FrontendDB.Close()

	logger.Infof("connected to FrontendDB:  %+v", cfg.Frontend.Database)

	var epoch uint64
	err = db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
	if err != nil {
		logger.Errorf("error retrieving latest epoch from the database: %v", err)
	}
	atomic.StoreUint64(&latestEpoch, epoch)

	m.Run()
}

func TestBalanceDecrease(t *testing.T) {
	latestEpoch := LatestEpoch()
	t.Logf("Running test Balance Decrease for epoch: %v", latestEpoch)
	result, err := db.GetValidatorsBalanceDecrease(latestEpoch)
	if err != nil {
		t.Errorf("error getting validators balance decrease %v", err)
		return
	}

	t.Logf("found %v validators losing balance", len(result))

	if len(result) > 0 {
		valOne := result[0]
		err := db.AddTestSubscription(10, types.ValidatorBalanceDecreasedEventName, valOne.Pubkey, 0, latestEpoch-1)
		if err != nil {
			t.Errorf("error creating test subscription %v", err)
			return
		}
		t.Cleanup(func() {
			_, err := db.FrontendDB.Exec("DELETE FROM users_subscriptions where user_id = 10")
			if err != nil {
				t.Errorf("error cleaning up TestBalanceDecrease err: %v", err)
				return
			}
		})
	} else {
		t.Error("error no validators are losing a balance, this test cannot complete")
		return
	}

	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	err = collectValidatorBalanceDecreasedNotifications(notificationsByUserID)
	if err != nil {
		t.Errorf("error collecting balance decrease notifications err: %v", err)
	}

	t.Logf("notifications recorded: %v", notificationsByUserID)

	subs, ok := notificationsByUserID[10]
	if !ok {
		t.Errorf("no notifications for user %v exist in %+v", 10, notificationsByUserID)
		return
	}

	t.Logf("test user has the following subs: %v", subs)

	notifications, ok := subs[types.ValidatorBalanceDecreasedEventName]
	if !ok {
		t.Errorf("no notifications for user %v exist in %+v", 10, notificationsByUserID)
		return
	}

	if len(notifications) == 0 {
		t.Errorf("error expected to receive at least one event")
		return
	}

	t.Logf("notifications for test user %v", notifications)

	expected := result[0].Pubkey
	got := notifications[0].GetEventFilter()
	if got != expected {
		t.Errorf("error unexpected event created expected: %v but got %v", expected, got)
		return
	}
}
