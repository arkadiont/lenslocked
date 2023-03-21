package models

import (
	"fmt"
	"github.com/go-mail/mail/v2"
)

const (
	DefaultSender = "support@lenslocked.com"
)

type Email struct {
	From      string
	To        string
	Subject   string
	PlainText string
	Html      string
}

type EmailService interface {
	Send(email Email) error
	ForgotPassword(to, resetURL string) error
}

type emailService struct {
	DefaultSender string
	dialer        *mail.Dialer
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

func NewEmailService(config SMTPConfig) EmailService {
	es := emailService{
		dialer: mail.NewDialer(config.Host, config.Port, config.User, config.Pass),
	}
	return &es
}

func (es *emailService) Send(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("To", email.To)
	es.setFrom(msg, email)
	msg.SetHeader("Subject", email.Subject)
	switch {
	case email.PlainText != "" && email.Html != "":
		msg.SetBody("text/plain", email.PlainText)
		msg.AddAlternative("text/html", email.Html)
	case email.PlainText != "":
		msg.SetBody("text/plain", email.PlainText)
	case email.Html != "":
		msg.SetBody("text/html", email.Html)
	}
	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	return nil
}

func (es *emailService) ForgotPassword(to, resetURL string) error {
	msg := "To reset your password, please visit the following link:"
	email := Email{
		To:        to,
		Subject:   "Reset your password",
		PlainText: fmt.Sprintf("%s %s", msg, resetURL),
		Html:      fmt.Sprintf(`<p>%s <a href="%s">%s</a></p>`, msg, resetURL, resetURL),
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot pass: %w", err)
	}
	return nil
}

func (es *emailService) setFrom(msg *mail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.SetHeader("From", from)
}
