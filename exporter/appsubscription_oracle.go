package exporter

import (
	"context"
	"encoding/base64"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Gurpartap/storekit-go"
	"github.com/awa/go-iap/appstore"
	"github.com/awa/go-iap/appstore/api"
	"github.com/awa/go-iap/playstore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
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

		googleClient, err := initGoogle()
		if googleClient == nil {
			logger.Errorf("error initializing google client: %v", err)
			return
		}

		appleClient, err := initApple()
		if err != nil {
			logger.Errorf("error initializing apple client: %v", err)
			return
		}

		for _, receipt := range receipts {
			// TODO: At some point we can drop the loop validator approach for iOS purchases and replace it with
			// the notifications approach.
			// https://developer.apple.com/documentation/appstoreservernotifications

			if receipt.Store == "ethpool" {
				continue
			}

			time.Sleep(100 * time.Millisecond)
			valid, err := VerifyReceipt(googleClient, appleClient, receipt)

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

			// In case of fe stripe, just skip updating the state since this will be handled elsewhere
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

// Does not verify stripe or ethpool payments as those are handled differently
func VerifyReceipt(googleClient *playstore.Client, appleClient *api.StoreClient, receipt *types.PremiumData) (*VerifyResponse, error) {
	if receipt.Store == "ios-appstore" {
		return verifyApple(appleClient, receipt)
	} else if receipt.Store == "android-playstore" {
		return verifyGoogle(googleClient, receipt)
	} else if receipt.Store == "manuall" {
		return verifyManuall(receipt)
	} else {
		return &VerifyResponse{
			Valid:          false,
			ExpirationDate: 0,
			RejectReason:   "invalid_store",
		}, nil
	}
}

func initGoogle() (*playstore.Client, error) {
	if len(utils.Config.Frontend.AppSubsGoogleJSONPath) == 0 {
		return nil, errors.New("google app subs json path not set")
	}

	var jsonKey []byte
	var err error
	if strings.Contains(utils.Config.Frontend.AppSubsGoogleJSONPath, ".json") {
		jsonKey, err = os.ReadFile(utils.Config.Frontend.AppSubsGoogleJSONPath)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Can not read google json key file %v", utils.Config.Frontend.AppSubsGoogleJSONPath))
		}
	} else {
		jsonKey = []byte(utils.Config.Frontend.AppSubsGoogleJSONPath)
	}

	client, err := playstore.New(jsonKey)
	return client, err
}

func initApple() (*api.StoreClient, error) {
	if len(utils.Config.Frontend.Apple.Certificate) == 0 {
		return nil, errors.New("apple certificate path not set")
	}

	var keyContent []byte
	var err error
	if strings.Contains(utils.Config.Frontend.Apple.Certificate, "BEGIN PRIVATE KEY") {
		keyContent = []byte(utils.Config.Frontend.Apple.Certificate)
	} else {
		keyContent, err = os.ReadFile(utils.Config.Frontend.Apple.Certificate)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("can not load apple certificate for file %v", utils.Config.Frontend.Apple.Certificate))
		}
	}

	return api.NewStoreClient(&api.StoreConfig{
		KeyContent: keyContent,                          // Loads a .p8 certificate
		KeyID:      utils.Config.Frontend.Apple.KeyID,   // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
		BundleID:   "in.beaconcha.mobile",               // Your app’s bundle ID
		Issuer:     utils.Config.Frontend.Apple.IssueID, // Your issuer ID from the Keys page in App Store Connect (Ex: "57246542-96fe-1a63-e053-0824d011072a")
		Sandbox:    false,                               // default is Production
	}), nil
}

func verifyGoogle(client *playstore.Client, receipt *types.PremiumData) (*VerifyResponse, error) {
	response := &VerifyResponse{
		Valid:          false,
		ExpirationDate: 0,
		RejectReason:   "",
	}

	if client == nil {
		var err error
		client, err = initGoogle()
		if err != nil {
			response.RejectReason = "gclient_init_exception"
			return response, errors.New("google client can't be initialized")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := client.VerifySubscription(ctx, "in.beaconcha.mobile", receipt.ProductID, receipt.Receipt)
	if err != nil || resp == nil {
		response.RejectReason = "invalid_state"
		return response, err
	}

	otherReceiptID, found := duplicateOrderMap[resp.OrderId]
	if found {
		if otherReceiptID != receipt.ID {
			response.RejectReason = "duplicate"
			return response, err
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

func verifyApple(apple *api.StoreClient, receipt *types.PremiumData) (*VerifyResponse, error) {
	response := &VerifyResponse{
		Valid:          false,
		ExpirationDate: 0,
		RejectReason:   "",
	}

	if apple == nil {
		var err error
		apple, err = initApple()
		if err != nil {
			response.RejectReason = "aclient_init_exception"
			return response, errors.New("apple client can't be initialized")
		}
	}

	// legacy resolver for old receipts, can be removed at some point
	if len(receipt.Receipt) > 100 {
		transactionID, err := getLegacyAppstoreTransactionIDByReceipt(receipt.Receipt, receipt.ProductID)
		if err != nil {
			utils.LogError(err, "error resolving legacy appstore receipt", 0, map[string]interface{}{"receipt": receipt.Receipt})
			response.RejectReason = "exception_legresolve"
			return response, err
		}
		receipt.Receipt = transactionID
		time.Sleep(50 * time.Millisecond) // avoid rate limiting
	}

	res, err := apple.GetALLSubscriptionStatuses(context.Background(), receipt.Receipt, nil)
	if err != nil {
		response.RejectReason = "exception"
		return response, err
	}

	if res.BundleId != "in.beaconcha.mobile" {
		response.RejectReason = "unknown_bundle"
		return response, nil
	}

	client := appstore.New()

	for _, val := range res.Data {
		for _, last := range val.LastTransactions {

			if last.Status == api.SubscriptionActive || last.Status == api.SubscriptionGracePeriod {
				token := jwt.Token{}

				err = client.ParseNotificationV2(last.SignedTransactionInfo, &token)
				if err != nil {
					response.RejectReason = "exception_parse"
					return response, nil
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					response.RejectReason = "exception_cast"
					return response, nil
				}

				productId, ok := claims["productId"].(string)
				if !ok {
					response.RejectReason = "invalid_product_id"
					return response, nil
				}
				receipt.ProductID = productId

				expiresDateFloat, ok := claims["expiresDate"].(float64)
				if !ok {
					response.RejectReason = "invalid_expires_date"
					return response, nil
				}
				expiresDateUint64 := int64(math.Round(expiresDateFloat))

				response.Valid = true
				response.ExpirationDate = expiresDateUint64
				return response, nil
			}
		}
	}

	// Return unknown here since expired would disable checking this purchase in the future.
	// I could not find any remarks in apples doc if they reuse the original transaction id if at some
	// point, fe. the user re-subs after a cancellation. So we just keep checking the receipt for now
	// https://developer.apple.com/documentation/appstoreserverapi/originaltransactionid
	response.RejectReason = "unknown"
	return response, nil
}

// Can be removed in a future release once app adoption for new v2 purchase register has reached critical mass
func getLegacyAppstoreTransactionIDByReceipt(receipt, premiumPkg string) (string, error) {
	appStoreSecret := utils.Config.Frontend.Apple.LegacyAppSubsAppleSecret
	client := storekit.NewVerificationClient().OnProductionEnv()

	receiptData, err := base64.StdEncoding.DecodeString(receipt)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, resp, err := client.Verify(ctx, &storekit.ReceiptRequest{
		ReceiptData:            receiptData,
		Password:               appStoreSecret,
		ExcludeOldTransactions: true,
	})

	if err != nil {
		return "", err
	}

	if resp.Status != 0 {
		return "", errors.New("invalid state")
	}

	if len(resp.LatestReceiptInfo) == 0 {
		return "", errors.New("not found")
	}

	for _, latestReceiptInfo := range resp.LatestReceiptInfo {
		if premiumPkg == latestReceiptInfo.ProductId {
			return latestReceiptInfo.OriginalTransactionId, nil
		}
	}

	return "", errors.New("not found")
}

func updateValidationState(receipt *types.PremiumData, validation *VerifyResponse) {
	err := db.UpdateUserSubscription(
		nil,
		receipt.ID,
		validation.Valid,
		validation.ExpirationDate,
		validation.RejectReason,
	)
	if err != nil {
		fmt.Printf("error updating subscription state %v", err)
	}

	// in case user upgrades downgrades package (fe on iOS) we can automatically update the product here too
	err = db.UpdateUserSubscriptionProduct(
		nil,
		receipt.ID,
		receipt.ProductID,
	)
	if err != nil {
		fmt.Printf("error updating subscription product id %v", err)
	}
}

type VerifyResponse struct {
	Valid          bool
	ExpirationDate int64
	RejectReason   string
}
