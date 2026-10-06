package auth

import "strings"

// EmailOTPSender delivers OTP codes over email by reusing the already-wired SMTP
// EmailSender. No SMTP dialing logic is duplicated here.
//
// DEV/TESTING FALLBACK ONLY: the OTPSender.Send signature carries only the
// phone value, so this sender treats that value as the destination email
// address. Genuine phone-only users have no email, so this transport cannot
// deliver to them — it is only useful when a tester supplies an email in the
// phone field. The primary OTP delivery path is Twilio SMS.
type EmailOTPSender struct {
	mailer   EmailSender
	template string
}

// NewEmailOTPSender builds the SMTP-backed OTP email fallback. When template is
// empty the default SMS template is applied (the same {code} token is used).
func NewEmailOTPSender(mailer EmailSender, template string) *EmailOTPSender {
	if strings.TrimSpace(template) == "" {
		template = twilioDefaultTemplate
	}
	return &EmailOTPSender{mailer: mailer, template: template}
}

// Enabled implements OTPSender; it mirrors the underlying mailer's state.
func (s *EmailOTPSender) Enabled() bool {
	return s != nil && s.mailer != nil && s.mailer.Enabled()
}

// Send implements OTPSender by emailing the code. See the type doc: the phone
// argument is treated as the recipient email address (dev/testing only).
func (s *EmailOTPSender) Send(phone, code string) error {
	body := strings.ReplaceAll(s.template, "{code}", code)
	return s.mailer.Send(phone, "Your verification code", body)
}
