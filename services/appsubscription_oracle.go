package services

import (
	"context"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
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

func VerifyReceipt(googleClient *playstore.Client, receipt *types.PremiumData) (*VerifyResponse, error) {
	if receipt.Store == "ios-appstore" {
		return VerifyApple(receipt)
	} else if receipt.Store == "android-playstore" {
		return VerifyGoogle(googleClient, receipt)
	} else {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "invalid_store",
		}, nil
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

func VerifyGoogle(client *playstore.Client, receipt *types.PremiumData) (*VerifyResponse, error) {
	if client == nil {
		client = initGoogle()
		if client == nil {
			return &VerifyResponse{
				Valid:          false,
				ExpirationDate: 0,
				RejectReason:   "gclient_init_exception",
			}, errors.New("Google client can't be initialized")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := client.VerifySubscription(ctx, "in.beaconcha.mobile", receipt.ProductID, receipt.Receipt)
	if err != nil || resp == nil {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "invalid_state",
		}, err
	}

	now := time.Now().Unix() * 1000
	valid := resp.ExpiryTimeMillis > now

	return &VerifyResponse{
		Valid:          valid,
		ExpirationDate: resp.ExpiryTimeMillis / 1000,
		RejectReason:   rejectReason(valid),
	}, nil
}

func rejectReason(valid bool) string {
	if valid {
		return ""
	} else {
		return "expired"
	}
}

func VerifyApple(receipt *types.PremiumData) (*VerifyResponse, error) {
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
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "exception",
		}, err
	}

	if resp.Status != 0 {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "invalid_state",
		}, nil
	}

	if len(resp.LatestReceiptInfo) == 0 {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "possible_jailbreak",
		}, nil
	}

	for _, latestReceiptInfo := range resp.LatestReceiptInfo {
		productID := latestReceiptInfo.ProductId

		if receipt.ProductID == productID {
			expiresAtMs := latestReceiptInfo.ExpiresDateMs
			if expiresAtMs == 0 {
				return &VerifyResponse{
					Valid:          false,
					ExpirationDate: 0,
					RejectReason:   "expires_0",
				}, nil
			}

			valid := expiresAtMs > time.Now().Unix()*1000

			return &VerifyResponse{
				Valid:          valid,
				ExpirationDate: expiresAtMs / 1000,
				RejectReason:   rejectReason(valid),
			}, nil
		}
	}

	return &VerifyResponse{
		Valid:          false,
		ExpirationDate: 0,
		RejectReason:   "unknown",
	}, nil
}

func updateValidationState(receipt *types.PremiumData, validation *VerifyResponse) {
	err := db.UpdateUserSubscription(receipt.ID, validation.Valid, validation.ExpirationDate, validation.RejectReason)
	if err != nil {
		fmt.Printf("error updating subscription state %v", err)
	}
}

type VerifyResponse struct {
	Valid          bool
	ExpirationDate int64
	RejectReason   string
}
