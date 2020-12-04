package notify

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "firebase")

func SendPushBatch(messages []*messaging.Message) (*messaging.BatchResponse, error) {
	ctx := context.Background()
	//opt := option.WithCredentialsFile("./run-local/firebaseAdminSdk.json")
	app, err := firebase.NewApp(context.Background(), nil) //, opt)
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

	logger.Infof("Successfully send %v firebase notifications. Successfull: %v | Failed: %v", len(messages), result.SuccessCount, result.FailureCount)
	return result, nil
}
