package notifications

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailMessage is sent through an email provider.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// EmailProvider delivers transactional email.
type EmailProvider interface {
	Send(ctx context.Context, message EmailMessage) error
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// SMTPProvider sends email via SMTP.
type SMTPProvider struct {
	cfg SMTPConfig
}

// NewSMTPProvider creates an SMTP email provider.
func NewSMTPProvider(cfg SMTPConfig) *SMTPProvider {
	return &SMTPProvider{cfg: cfg}
}

// Send delivers an email when SMTP is configured.
func (p *SMTPProvider) Send(_ context.Context, message EmailMessage) error {
	if p == nil || strings.TrimSpace(p.cfg.Host) == "" || strings.TrimSpace(message.To) == "" {
		return nil
	}

	from := strings.TrimSpace(p.cfg.From)
	if from == "" {
		from = p.cfg.User
	}
	if from == "" {
		return fmt.Errorf("smtp from address is not configured")
	}

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)
	body := strings.TrimSpace(message.Body)
	subject := strings.TrimSpace(message.Subject)
	payload := []byte(
		"From: " + from + "\r\n" +
			"To: " + message.To + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
			body + "\r\n",
	)

	var auth smtp.Auth
	if p.cfg.User != "" {
		auth = smtp.PlainAuth("", p.cfg.User, p.cfg.Password, p.cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{message.To}, payload)
}

// NoopEmailProvider discards email until SMTP is configured.
type NoopEmailProvider struct{}

// Send implements EmailProvider.
func (NoopEmailProvider) Send(context.Context, EmailMessage) error {
	return nil
}
