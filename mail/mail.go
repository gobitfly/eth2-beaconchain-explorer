package mail

import (
	"bytes"
	"context"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/smtp"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"
)

type MailTemplate struct {
	Mail   types.Email
	Domain string
}

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
func SendHTMLMail(to, subject string, msg types.Email, attachment []types.EmailAttachment) error {
	var renderer = templates.GetTemplate("mail/layout.html")

	var err error
	var body bytes.Buffer

	if utils.Config.Frontend.Mail.SMTP.User != "" {
		headers := "MIME-version: 1.0;\nContent-Type: text/html;"
		body.Write([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n%s\r\n", to, subject, headers)))
		renderer.Execute(&body, MailTemplate{Mail: msg, Domain: utils.Config.Frontend.SiteDomain})

		fmt.Println("Email Attachments will not work with SMTP server")
		err = SendMailSMTP(to, body.Bytes())
	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
		_ = renderer.ExecuteTemplate(&body, "layout", MailTemplate{Mail: msg, Domain: utils.Config.Frontend.SiteDomain})
		content := body.String()
		err = SendMailMailgun(to, subject, content, createTextMessage(msg), attachment)
	} else {
		utils.LogError(nil, "error sending reset-email: invalid config for mail-service", 0)
		err = nil
	}
	return err
}

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
func SendTextMail(to, subject, msg string, attachment []types.EmailAttachment) error {
	var err error
	if utils.Config.Frontend.Mail.SMTP.User != "" {
		fmt.Println("Email Attachments will not work with SMTP server")
		err = SendTextMailSMTP(to, subject, msg)
	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
		err = SendTextMailMailgun(to, subject, msg, attachment)
	} else {
		err = fmt.Errorf("invalid config for mail-service")
	}
	return err
}

func createTextMessage(msg types.Email) string {
	return fmt.Sprintf("%s\n\n%s\n\n― You are receiving this because you are staking on Ethermine Staking. You can manage your subscriptions at %s.", msg.Title, msg.Body, msg.SubscriptionManageURL)
}

// SendMail sends an email to the given address with the given message.
// It will use smtp if configured otherwise it will use gunmail if configured.
// func SendMail(to, subject, msg string, attachment []types.EmailAttachment) error {
// 	var err error
// 	if utils.Config.Frontend.Mail.SMTP.User != "" {
// 		fmt.Println("Email Attachments will not work with SMTP server")
// 		err = SendMailSMTP(to, subject, msg)
// 	} else if utils.Config.Frontend.Mail.Mailgun.PrivateKey != "" {
// 		err = SendMailMailgun(to, subject, msg, attachment)
// 	} else {
// 		err = fmt.Errorf("invalid config for mail-service")
// 	}
// 	return err
// }

// SendMailRateLimited sends an email to a given address with the given message.
// It will return a ratelimit-error if the configured ratelimit is exceeded.
func SendMailRateLimited(to, subject string, msg types.Email, attachment []types.EmailAttachment) error {
	if utils.Config.Frontend.MaxMailsPerEmailPerDay > 0 {
		now := time.Now()
		count, err := db.GetMailsSentCount(to, now)
		if err != nil {
			return err
		}
		if count >= utils.Config.Frontend.MaxMailsPerEmailPerDay {
			timeLeft := now.Add(utils.Day).Truncate(utils.Day).Sub(now)
			return &types.RateLimitError{TimeLeft: timeLeft}
		}
	}

	err := db.CountSentMail(to)
	if err != nil {
		// only log if counting did not work
		return fmt.Errorf("error counting sent email: %v", err)
	}

	err = SendHTMLMail(to, subject, msg, attachment)
	if err != nil {
		return err
	}

	return nil
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendMailSMTP(to string, msg []byte) error {
	server := utils.Config.Frontend.Mail.SMTP.Server // eg. smtp.gmail.com:587
	host := utils.Config.Frontend.Mail.SMTP.Host     // eg. smtp.gmail.com
	from := utils.Config.Frontend.Mail.SMTP.User     // eg. userxyz123@gmail.com
	password := utils.Config.Frontend.Mail.SMTP.Password
	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(server, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("error sending mail via smtp: %w", err)
	}

	return nil
}

// SendMailMailgun sends an email to the given address with the given message, using mailgun.
func SendMailMailgun(to, subject, msgHtml, msgText string, attachment []types.EmailAttachment) error {
	mg := mailgun.NewMailgun(
		utils.Config.Frontend.Mail.Mailgun.Domain,
		utils.Config.Frontend.Mail.Mailgun.PrivateKey,
	)
	message := mg.NewMessage(utils.Config.Frontend.Mail.Mailgun.Sender, subject, msgText, to)
	message.SetHtml(msgHtml)

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

	// mg := mailgun.NewMailgun(
	// 	utils.Config.Frontend.Mail.Mailgun.Domain,
	// 	utils.Config.Frontend.Mail.Mailgun.PrivateKey,
	// )
	// mg.SetAPIBase(mailgun.APIBaseEU)
	// niceFrom := fmt.Sprintf("%v <%v>", "Ethermine Staking", utils.Config.Frontend.Mail.Mailgun.Sender)

	// message := mg.NewMessage(niceFrom, subject, msgText, to)
	// message.SetHtml(msgHtml)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()
	// if len(attachment) > 0 {
	// 	for _, att := range attachment {
	// 		message.AddBufferAttachment(att.Name, att.Attachment)
	// 	}
	// }

	// // Send the message with a 10 second timeout
	// resp, id, err := mg.Send(ctx, message)
	// if err != nil {
	// 	logrus.WithField("resp", resp).WithField("id", id).Errorf("error sending mail via mailgun: %v", err)
	// 	return fmt.Errorf("error sending mail via mailgun: %w", err)
	// }

	// return nil
}

// SendMailSMTP sends an email to the given address with the given message, using smtp.
func SendTextMailSMTP(to, subject, body string) error {
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
func SendTextMailMailgun(to, subject, msg string, attachment []types.EmailAttachment) error {
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
