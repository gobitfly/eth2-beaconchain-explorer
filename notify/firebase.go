package notify

import (
	"context"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var logger = logrus.New().WithField("module", "notify").WithField("service", "firebase")

func isRelevantError(response *messaging.SendResponse) bool {
	if !response.Success && response.Error != nil {
		// Ignore https://stackoverflow.com/questions/58308835/using-firebase-for-notifications-getting-app-instance-has-been-unregistered
		// Errors since they indicate that the user token is expired
		if !strings.Contains(response.Error.Error(), "registration-token-not-registered") {
			return true
		}
	}
	return false
}

func SendPushBatch(messages []*messaging.Message) error {
	credentialsPath := utils.Config.Notifications.FirebaseCredentialsPath
	if credentialsPath == "" {
		logger.Errorf("firebase credentials path not provided, disabling push notifications")
		return nil
	}

	ctx := context.Background()
	var opt option.ClientOption

	if strings.HasPrefix(credentialsPath, "projects/") {
		x, err := utils.AccessSecretVersion(credentialsPath)
		if err != nil {
			return fmt.Errorf("error getting firebase config from secret store: %v", err)
		}
		opt = option.WithCredentialsJSON([]byte(*x))
	} else {
		opt = option.WithCredentialsFile(credentialsPath)
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

	currentMessages := messages
	tries := 0
	for _, s := range waitBeforeTryInSeconds {
		time.Sleep(s * time.Second)
		tries++

		result, err = client.SendAll(context.Background(), currentMessages)
		if err != nil {
			logger.Errorf("error sending push notifications: %v", err)
			return err
		}

		resultSuccessCount += result.SuccessCount
		resultFailureCount += result.FailureCount

		newMessages := make([]*messaging.Message, 0, result.FailureCount)
		if result.FailureCount > 0 {
			for i, response := range result.Responses {
				if isRelevantError(response) {
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
