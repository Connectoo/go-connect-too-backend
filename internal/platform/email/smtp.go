package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Config holds SMTP connection settings.
type Config struct {
	Host string
	User string
	Pass string
	From string
}

// Sender delivers plain-text email messages.
type Sender struct {
	cfg Config
}

// NewSender creates an SMTP email sender.
func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

// Enabled reports whether SMTP is configured.
func (s *Sender) Enabled() bool {
	return s != nil && s.cfg.Host != "" && s.cfg.From != ""
}

// Send delivers a plain-text email to one recipient.
func (s *Sender) Send(to, subject, body string) error {
	if !s.Enabled() {
		return fmt.Errorf("smtp is not configured")
	}

	msg := strings.Builder{}
	msg.WriteString("From: ")
	msg.WriteString(s.cfg.From)
	msg.WriteString("\r\nTo: ")
	msg.WriteString(to)
	msg.WriteString("\r\nSubject: ")
	msg.WriteString(subject)
	msg.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(body)

	addr := s.cfg.Host
	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, hostFromAddr(addr))
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg.String()))
}

func hostFromAddr(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		return addr[:idx]
	}
	return addr
}
