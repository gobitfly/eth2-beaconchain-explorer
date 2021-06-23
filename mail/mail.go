package mail

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/smtp"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"
)

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
func SendMail(to, subject, msg string, attachment []types.EmailAttachment) error {
	var err error
	if utils.Config.Frontend.Mail.SMTP.User != "" {
		fmt.Println("Email Attachments will not work with SMTP server")
		err = SendMailSMTP(to, subject, msg)
	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
		err = SendMailMailgun(to, subject, msg, attachment)
	} else {
		err = fmt.Errorf("invalid config for mail-service")
	}
	return err
}

// SendMailRateLimited sends an email to a given address with the given message.
// It will return a ratelimit-error if the configured ratelimit is exceeded.
func SendMailRateLimited(to, subject, msg string, attachment []types.EmailAttachment) error {
	if utils.Config.Frontend.MaxMailsPerEmailPerDay > 0 {
		now := time.Now()
		count, err := db.GetMailsSentCount(to, now)
		if err != nil {
			return err
		}
		if count >= utils.Config.Frontend.MaxMailsPerEmailPerDay {
			timeLeft := now.Add(time.Hour * 24).Truncate(time.Hour * 24).Sub(now)
			return &types.RateLimitError{timeLeft}
		}
	}

	err := db.CountSentMail(to)
	if err != nil {
		// only log if counting did not work
		return fmt.Errorf("error counting sent email: %v", err)
	}

	err = SendMail(to, subject, msg, attachment)
	if err != nil {
		return err
	}

	return nil
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendMailSMTP(to, subject, body string) error {
	server := utils.Config.Frontend.Mail.SMTP.Server // eg. smtp.gmail.com:587
	host := utils.Config.Frontend.Mail.SMTP.Host     // eg. smtp.gmail.com
	from := utils.Config.Frontend.Mail.SMTP.User     // eg. userxyz123@gmail.com
	password := utils.Config.Frontend.Mail.SMTP.Password
	auth := smtp.PlainAuth("", from, password, host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	err := smtp.SendMail(server, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("error sending mail via smtp: %w", err)
	}

	return nil
}

// SendMailMailgun sends an email to the given address with the given message, using mailgun.
func SendMailMailgun(to, subject, msg string, attachment []types.EmailAttachment) error {
	mg := mailgun.NewMailgun(
		utils.Config.Frontend.Mail.Mailgun.Domain,
		utils.Config.Frontend.Mail.Mailgun.PrivateKey,
	)
	message := mg.NewMessage(utils.Config.Frontend.Mail.Mailgun.Sender, subject, msg, to)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if len(attachment) > 0 {
		for _, att := range attachment {
			message.AddBufferAttachment(att.Name, att.Attachment)
		}
	}

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		logrus.WithField("resp", resp).WithField("id", id).Errorf("error sending mail via mailgun: %v", err)
		return fmt.Errorf("error sending mail via mailgun: %w", err)
	}

	return nil
}
