package notify

import (
	"context"
	"eth2-exporter/utils"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var logger = logrus.New().WithField("module", "notify").WithField("service", "firebase")

func SendPushBatch(messages []*messaging.Message) (*messaging.BatchResponse, error) {
	credentialsPath := utils.Config.Notifications.FirebaseCredentialsPath
	if credentialsPath == "" {
		logger.Errorf("firebase credentials path not provided, disabling push notifications")
		return nil, nil
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.Errorf("error initializing app:  %v", err)
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		logger.Errorf("error initializing messaging: %v", err)
		return nil, err
	}

	result, err := client.SendAll(context.Background(), messages)
	if err != nil {
		logger.Errorf("error sending push notifications: %v", err)
		return nil, err
	}
	for _, response := range result.Responses {
		if !response.Success {
			logger.Errorf("firebase error %v %v", response.Error, response.MessageID)
		}
	}

	logger.Infof("Successfully send %v firebase notifications. Successful: %v | Failed: %v", len(messages), result.SuccessCount, result.FailureCount)
	return result, nil
}
