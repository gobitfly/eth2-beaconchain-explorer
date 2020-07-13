package utils

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"
)

// SendMail sends an email to the given address with the given message.
func SendMail(to, subject, msg string) error {
	if Config.Frontend.Mail.SMTP.User != "" {
		return SendMailSMTP(to, subject, msg)
	} else if Config.Frontend.Mail.Gunmail.PrivateKey != "" {
		return SendMailGunmail(to, subject, msg)
	}
	return fmt.Errorf("invalid config for mail-service")
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendMailSMTP(to, subject, msg string) error {
	server := Config.Frontend.Mail.SMTP.Server // eg. smtp.gmail.com:587
	host := Config.Frontend.Mail.SMTP.Host     // eg. smtp.gmail.com
	from := Config.Frontend.Mail.SMTP.User     // eg. userxyz123@gmail.com
	password := Config.Frontend.Mail.SMTP.Password
	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(server, auth, from, []string{to}, []byte(msg))
	if err != nil {
		logrus.Errorf("error sending mail via smtp: %v", err)
	}

	return err
}

// SendMailGunmail sends an email to the given address with the given message, using gunmail.
func SendMailGunmail(to, subject, msg string) error {
	mg := mailgun.NewMailgun(
		Config.Frontend.Mail.Gunmail.Domain,
		Config.Frontend.Mail.Gunmail.PrivateKey,
	)
	message := mg.NewMessage(Config.Frontend.Mail.Gunmail.Sender, subject, msg, to)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		logrus.WithField("resp", resp).WithField("id", id).Errorf("error sending mail via mailgun: %v", err)
	}

	return err
}
