package notify

import (
	"context"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "notify").WithField("service", "firebase")

func SendPushBatch(messages []*messaging.Message) (*messaging.BatchResponse, error) {
	ctx := context.Background()
	//utils.Config.Frontend
	//opt := option.WithCredentialsFile("./run-local/firebaseAdminSdk.json")
	app, err := firebase.NewApp(context.Background(), nil) //, opt)
	if err != nil {
		logger.Errorf("error initializing app:  %v", err)
		return nil, err
	}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		logger.Infof("env %v", pair)
	}

	logger.Infof("Firebase app %v", app)

	for _, message := range messages {
		logger.Infof("Firebase messages %v", message)
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

	logger.Infof("Successfully send %v firebase notifications. Successfull: %v | Failed: %v", len(messages), result.SuccessCount, result.FailureCount)
	return result, nil
}
