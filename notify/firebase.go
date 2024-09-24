package notify

import (
	"context"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var logger = logrus.New().WithField("module", "notify").WithField("service", "firebase")

func isRelevantError(response *messaging.SendResponse) bool {
	if !response.Success && response.Error != nil {
		// Ignore https://stackoverflow.com/questions/58308835/using-firebase-for-notifications-getting-app-instance-has-been-unregistered
		// Errors since they indicate that the user token is expired
		if !strings.Contains(response.Error.Error(), "registration-token-not-registered") &&
			!strings.Contains(response.Error.Error(), "Requested entity was not found.") &&
			!strings.Contains(response.Error.Error(), "Request contains an invalid argument.") {
			return true
		}
	}
	return false
}

func SendPushBatch(messages []*messaging.Message, dryRun bool) error {
	credentialsPath := utils.Config.Notifications.FirebaseCredentialsPath
	if credentialsPath == "" {
		logger.Errorf("firebase credentials path not provided, disabling push notifications")
		return nil
	}

	ctx := context.Background()
	var opt option.ClientOption

	if strings.Contains(credentialsPath, ".json") && len(credentialsPath) < 200 {
		opt = option.WithCredentialsFile(credentialsPath)
	} else {
		opt = option.WithCredentialsJSON([]byte(credentialsPath))
	}

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.Errorf("error initializing app:  %v", err)
		return err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		logger.Errorf("error initializing messaging: %v", err)
		return err
	}

	var waitBeforeTryInSeconds = []time.Duration{0, 2, 4, 8, 16}
	var resultSuccessCount, resultFailureCount int = 0, 0
	var result *messaging.BatchResponse

	// badgeValue := 0
	// for i := range messages {
	// 	if messages[i].APNS == nil && messages[i].APNS.Payload == nil {
	// 		messages[i].APNS.Payload.Aps.Badge = &badgeValue
	// 	}
	// }

	currentMessages := messages
	tries := 0
	for _, s := range waitBeforeTryInSeconds {
		time.Sleep(s * time.Second)
		tries++
		if dryRun {
			result, err = client.SendEachDryRun(context.Background(), currentMessages)
		} else {
			result, err = client.SendEach(context.Background(), currentMessages)
		}
		if err != nil {
			logger.Errorf("error sending push notifications: %v", err)
			return err
		}

		resultSuccessCount += result.SuccessCount
		resultFailureCount += result.FailureCount

		newMessages := make([]*messaging.Message, 0, result.FailureCount)
		if result.FailureCount > 0 {
			for i, response := range result.Responses {
				logger.Info(response)
				if isRelevantError(response) {
					logger.Infof("retrying message %d", i)
					newMessages = append(newMessages, currentMessages[i])
					resultFailureCount--
				}
			}
		}

		currentMessages = newMessages
		if len(currentMessages) == 0 {
			break // no more messages to be proceeded
		}
	}

	if len(currentMessages) > 0 {
		for _, response := range result.Responses {
			if isRelevantError(response) {
				logger.WithError(response.Error).WithField("MessageID", response.MessageID).Errorf("firebase error")
				resultFailureCount++
			}
		}
	}

	logger.Infof("sent %d firebase notifications in %d of %d tries. successful: %d | failed: %d", len(messages), tries, len(waitBeforeTryInSeconds), resultSuccessCount, resultFailureCount)
	return nil
}
