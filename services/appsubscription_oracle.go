package services

import (
	"context"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"io/ioutil"
	"log"
	"time"

	"github.com/Gurpartap/storekit-go"
	"github.com/awa/go-iap/playstore"
)

func checkSubscriptions() {
	if !utils.Config.Frontend.VerifyAppSubs {
		return
	}
	for {
		start := time.Now()

		receipts, err := db.GetAllAppSubscriptions()

		if err != nil {
			logger.Errorf("error retrieving subscription data from db: %v", err)
			return
		}

		googleClient := initGoogle()

		for _, receipt := range receipts {
			valid, err := VerifyReceipt(googleClient, receipt)

			if err != nil {
				// error might indicate a connection problem, ignore validation response
				// for this iteration
				continue
			}
			updateValidationState(receipt, valid)
		}

		logger.WithField("subscriptions", len(receipts)).WithField("duration", time.Since(start)).Info("subscription update completed")
		time.Sleep(time.Second * 60 * 60 * 4) // 4h
	}
}

func VerifyReceipt(googleClient *playstore.Client, receipt *types.PremiumData) (bool, error) {
	if receipt.Store == "ios-appstore" {
		return VerifyApple(receipt)
	} else if receipt.Store == "android-playstore" {
		return VerifyGoogle(googleClient, receipt)
	} else {
		return false, nil
	}
}

func initGoogle() *playstore.Client {
	jsonKey, err := ioutil.ReadFile(utils.Config.Frontend.AppSubsGoogleJSONPath)
	if err != nil {
		log.Fatal(err)
	}

	client, _ := playstore.New(jsonKey)
	return client
}

func VerifyGoogle(client *playstore.Client, receipt *types.PremiumData) (bool, error) {
	if client == nil {
		client = initGoogle()
		if client == nil {
			return false, errors.New("Google client can't be initialized")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := client.VerifySubscription(ctx, "in.beaconcha.mobile", receipt.ProductID, receipt.Receipt)
	if err != nil || resp == nil {
		// todo check wheter we always get an expirytime or if an invalid subscription also throws an error
		return false, nil // err
	}

	return resp.ExpiryTimeMillis > time.Now().Unix(), nil
}

func VerifyApple(receipt *types.PremiumData) (bool, error) {
	appStoreSecret := utils.Config.Frontend.AppSubsAppleSecret
	client := storekit.NewVerificationClient().OnSandboxEnv() // TODO switch to production
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, resp, err := client.Verify(ctx, &storekit.ReceiptRequest{
		ReceiptData:            []byte(receipt.Receipt),
		Password:               appStoreSecret,
		ExcludeOldTransactions: true,
	})

	if err != nil {
		return false, err
	}

	if resp.Status != 0 {
		return false, nil
	}

	if len(resp.LatestReceiptInfo) == 0 {
		return false, nil
	}

	for _, latestReceiptInfo := range resp.LatestReceiptInfo {
		productID := latestReceiptInfo.ProductId

		if receipt.ProductID == productID {
			expiresAtMs := latestReceiptInfo.ExpiresDateMs
			if expiresAtMs == 0 {
				return false, nil
			}

			return expiresAtMs > time.Now().Unix(), nil
		}
	}

	return false, nil
}

func updateValidationState(receipt *types.PremiumData, valid bool) {
	db.UpdateUserSubscription(receipt.ID, valid)
}
