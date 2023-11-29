package exporter

import (
	"context"
	"encoding/base64"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Gurpartap/storekit-go"
	"github.com/awa/go-iap/playstore"
)

var duplicateOrderMap map[string]uint64 = make(map[string]uint64)

func checkSubscriptions() {
	if !utils.Config.Frontend.VerifyAppSubs {
		return
	}
	for {
		duplicateOrderMap = make(map[string]uint64)
		start := time.Now()

		receipts, err := db.GetAllAppSubscriptions()

		if err != nil {
			logger.Errorf("error retrieving subscription data from db: %v", err)
			return
		}

		googleClient := initGoogle()

		for _, receipt := range receipts {

			valid, err := VerifyReceipt(googleClient, receipt)

			if receipt.Store == "manuall" {
				valid, err = verifyManuall(receipt)
			}

			if receipt.Store == "ethpool" {
				continue
			}

			if err != nil {
				// error might indicate a connection problem, ignore validation response
				// for this iteration
				if strings.Contains(err.Error(), "expired") {
					err = db.SetSubscriptionToExpired(nil, receipt.ID)
					if err != nil {
						logger.Errorf("subscription set expired failed for [%v]: %v", receipt.ID, err)
					}
					continue
				}
				logger.Warnf("subscription verification failed in service for [%v]: %v", receipt.ID, err)
				continue
			}

			if valid.RejectReason == "invalid_store" {
				continue

			}
			updateValidationState(receipt, valid)
		}

		logger.WithField("subscriptions", len(receipts)).WithField("duration", time.Since(start)).Info("subscription update completed")
		time.Sleep(time.Hour * 4)
	}
}

func verifyManuall(receipt *types.PremiumData) (*VerifyResponse, error) {
	valid := receipt.ExpiresAt.Unix() > time.Now().Unix()
	return &VerifyResponse{
		Valid:          valid,
		ExpirationDate: receipt.ExpiresAt.Unix(),
		RejectReason:   rejectReason(valid),
	}, nil
}

func VerifyReceipt(googleClient *playstore.Client, receipt *types.PremiumData) (*VerifyResponse, error) {
	if receipt.Store == "ios-appstore" {
		return verifyApple(receipt)
	} else if receipt.Store == "android-playstore" {
		return verifyGoogle(googleClient, receipt)
	} else {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "invalid_store",
		}, nil
	}
}

func initGoogle() *playstore.Client {
	jsonKey, err := os.ReadFile(utils.Config.Frontend.AppSubsGoogleJSONPath)
	if err != nil {
		log.Fatal(err)
	}

	client, _ := playstore.New(jsonKey)
	return client
}

func verifyGoogle(client *playstore.Client, receipt *types.PremiumData) (*VerifyResponse, error) {
	if client == nil {
		client = initGoogle()
		if client == nil {
			return &VerifyResponse{
				Valid:          false,
				ExpirationDate: 0,
				RejectReason:   "gclient_init_exception",
			}, errors.New("google client can't be initialized")
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

	otherReceiptID, found := duplicateOrderMap[resp.OrderId]
	if found {
		if otherReceiptID != receipt.ID {
			return &VerifyResponse{
				Valid:          false,
				ExpirationDate: 0,
				RejectReason:   "duplicate",
			}, err
		}
	}

	duplicateOrderMap[resp.OrderId] = receipt.ID

	now := time.Now().Unix() * 1000
	valid := resp.ExpiryTimeMillis > now
	canceled := resp.UserCancellationTimeMillis > 0
	var reason string = rejectReason(valid)
	if canceled {
		if resp.CancelReason == 0 {
			reason = "user_canceled"
		} else if resp.CancelReason == 1 {
			reason = "system_canceled"
		} else if resp.CancelReason == 2 {
			reason = "canceled_replaced"
		} else if resp.CancelReason == 3 {
			reason = "developer_canceled"
		}
	}

	return &VerifyResponse{
		Valid:          valid && !canceled,
		ExpirationDate: resp.ExpiryTimeMillis / 1000,
		RejectReason:   reason,
	}, nil
}

func rejectReason(valid bool) string {
	if valid {
		return ""
	}
	return "expired"
}

func verifyApple(receipt *types.PremiumData) (*VerifyResponse, error) {
	appStoreSecret := utils.Config.Frontend.AppSubsAppleSecret
	client := storekit.NewVerificationClient().OnProductionEnv()

	receiptData, err := base64.StdEncoding.DecodeString(receipt.Receipt)
	if err != nil {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "exception_decode",
		}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, resp, err := client.Verify(ctx, &storekit.ReceiptRequest{
		ReceiptData:            receiptData,
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
		logger.Errorf("invalid_state %v", resp.Status)
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

		otherReceiptID, found := duplicateOrderMap[latestReceiptInfo.OriginalTransactionId]
		if found {
			if otherReceiptID != receipt.ID {
				return &VerifyResponse{
					Valid:          false,
					ExpirationDate: 0,
					RejectReason:   "duplicate",
				}, err
			}
		}

		duplicateOrderMap[latestReceiptInfo.OriginalTransactionId] = receipt.ID

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
	err := db.UpdateUserSubscription(nil, receipt.ID, validation.Valid, validation.ExpirationDate, validation.RejectReason)
	if err != nil {
		fmt.Printf("error updating subscription state %v", err)
	}
}

type VerifyResponse struct {
	Valid          bool
	ExpirationDate int64
	RejectReason   string
}
