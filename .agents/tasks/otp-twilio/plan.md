# Implementation Plan: Twilio SMS OTP Sender (+ email fallback)

Goal: add a Twilio SMS OTP delivery transport behind the existing `OTPSender`
interface, selected by `cfg.OTPProvider`, plus a minimal email-OTP fallback that
reuses the already-wired SMTP `EmailSender`. We keep our own OTP
generation/hash/store/expiry/verify — Twilio is a dumb SMS transport only. Twilio
Go SDK is NOT added; the sender calls the Twilio Messages REST API with net/http.

## Facts established during exploration

- `OTPSender` interface (`internal/modules/auth/otp_service.go`): `Enabled() bool`
  and `Send(phone, code string) error`. Existing impls: `NoopOTPSender`
  (Enabled=false), `LoggingOTPSender` (dev; logs code via `*slog.Logger`,
  Enabled=true). Call site in `RequestOTP`: `if s.otpSender != nil &&
  s.otpSender.Enabled() { if err := s.otpSender.Send(phone, code); err != nil {
  return err } }`. Contract unchanged.
- `EmailSender` interface (`internal/modules/auth/lifecycle_service.go`):
  `Enabled() bool` and `Send(to, subject, body string) error`. The platform SMTP
  sender `internal/platform/email.Sender` satisfies it and is already built in
  `server.go` as `emailSender` and passed via `auth.WithEmailSender`.
- Config (`internal/config/config.go`): env-based via `getEnv`/`os.Getenv`.
  Already has `OTPProvider` (`OTP_PROVIDER`, default ""), SMTP fields, and helper
  methods `FCMEnabled()` / `StorageEnabled()`. No Twilio or OTP-template fields yet.
- Provider-selection pattern to mirror: the FCM block in `server.go`
  (`pushProvider := notifications.PushProvider(notifications.NoopPushProvider{});
  if cfg.FCMEnabled() { pushProvider = notifications.NewFCMPushProvider(...) }`).
  `push_fcm.go` is the reference for an HTTP provider: struct holds
  `client *http.Client`, constructor sets `&http.Client{Timeout: 10 * time.Second}`.
- `.env` has an OTP block + a filled MSG91 block (leave untouched); `OTP_PROVIDER=`
  is empty. `.env.example` has commented OTP + MSG91 blocks.
- Build/test commands (Makefile): `make fmt` (gofmt -w .), `make test`
  (`go test $(go list ./...)`), `make run` (`go run ./cmd/api`). No DB needed for
  the unit tests below (auth tests use mocks).
- Existing auth tests are table/mock style in the `auth` package; `testConfig()`
  lives in `service_test.go`. config tests are in `config_test.go`.

## Email-fallback decision (noted gap)

`RequestOTP` only has `phone` + `role`; there is NO email in scope, and the
`OTPSender.Send(phone, code)` signature cannot carry an email. Per the task we do
NOT add a phone→email lookup. Decision: the `EmailOTPSender` treats its first
`Send` argument (the phone value) as the destination email address and emails the
code there via the injected `EmailSender`. This is ONLY meaningful as a
DEV/testing fallback where a tester supplies an email in the phone field; genuine
phone-only users have no email, so for them this transport cannot deliver. This
limitation is documented in code comments, `.env.example`, and
`IMPLEMENTED_FEATURES.md`. The PRIMARY deliverable is the Twilio SMS sender; the
email sender stays intentionally minimal and reuses the SMTP sender (no new SMTP
dialing logic).

---

# Implementation Plan

- [ ] 1. Add Twilio + OTP-template config fields and helper methods to config.
      Add struct fields `TwilioAccountSID`, `TwilioAuthToken`, `TwilioFromNumber`,
      `TwilioMessagingServiceSID` (string), and `OTPSMSTemplate` (string). In
      `Load()`, populate them: the Twilio fields via `os.Getenv("TWILIO_ACCOUNT_SID")`,
      `os.Getenv("TWILIO_AUTH_TOKEN")`, `os.Getenv("TWILIO_FROM_NUMBER")`,
      `os.Getenv("TWILIO_MESSAGING_SERVICE_SID")`; `OTPSMSTemplate` via
      `getEnv("OTP_SMS_TEMPLATE", "Your Go Connect verification code is {code}. It expires in a few minutes.")`.
      Add two helpers next to `FCMEnabled()`:
      `func (c *Config) TwilioEnabled() bool { return c.TwilioAccountSID != "" && c.TwilioAuthToken != "" && (c.TwilioFromNumber != "" || c.TwilioMessagingServiceSID != "") }`
      and `func (c *Config) EmailOTPEnabled() bool { return c.SMTPHost != "" }`.
      Files: internal/config/config.go
      Verify: `make fmt` then `go build ./internal/config/...` — compiles clean.

- [ ] 2. Add table-driven config tests for the new helpers and template default.
      Add `TestTwilioEnabled` (cases: all empty=false; SID+token+from=true;
      SID+token+messagingServiceSID=true; SID+token only, no from/MSS=false;
      missing token=false) and `TestEmailOTPEnabled` (SMTPHost set=true, empty=false)
      building `&Config{...}` literals directly (no env needed, like the helper
      inputs). Add a case asserting `Load()` sets `OTPSMSTemplate` to the default
      containing `{code}` when `OTP_SMS_TEMPLATE` is unset (set the required
      `DATABASE_URL`/`JWT_*` envs via `t.Setenv` as `TestLoad` does).
      Files: internal/config/config_test.go
      Verify: `go test ./internal/config/...` — all tests pass.

- [ ] 3. Implement the Twilio OTP sender.
      New file with `TwilioOTPSender` struct holding `accountSID string`,
      `authToken string`, `fromNumber string`, `messagingServiceSID string`,
      `template string`, `client *http.Client`, `log *slog.Logger`. Constructor
      `NewTwilioOTPSender(accountSID, authToken, fromNumber, messagingServiceSID, template string, client *http.Client, log *slog.Logger) *TwilioOTPSender`
      — if `client == nil`, default to `&http.Client{Timeout: 10 * time.Second}`;
      if `template == ""`, default to the same string as config. `Enabled()` returns
      `true`. `Send(phone, code string) error`:
        1. Normalize phone to E.164 with a helper `toE164(phone string) string`:
           strip all spaces, then ensure exactly one leading `+` (trim any leading
           `+`/`00`/whitespace, then prepend a single `+`). Twilio REQUIRES the
           leading `+` (opposite of MSG91). Keep remaining digits as-is.
        2. Build body by replacing `{code}` in the template with the code
           (`strings.ReplaceAll(template, "{code}", code)`).
        3. POST form-encoded (`application/x-www-form-urlencoded`) to
           `https://api.twilio.com/2010-04-01/Accounts/{AccountSID}/Messages.json`
           with HTTP Basic auth (`req.SetBasicAuth(accountSID, authToken)`); form
           fields: `To`=E.164 phone, `Body`=message, and either `From`=fromNumber
           OR `MessagingServiceSid`=messagingServiceSID (prefer MessagingServiceSid
           when set, else From). Use `http.NewRequestWithContext` with
           `context.Background()`.
        4. On non-2xx: decode Twilio JSON error body (`{"code":..,"message":".."}`)
           and return `fmt.Errorf("twilio send failed: status %d: %s (code %d)", ...)`.
           NEVER include the OTP code, the auth token, or the full request in the
           error or any log. Log only a provider-level failure at Warn with the
           Twilio status/code (no secrets, no OTP) if `log != nil`.
        5. On 2xx: return nil (optionally debug-log "otp sms sent" with phone only,
           never the code).
      Files: internal/modules/auth/otp_sender_twilio.go
      Verify: `go build ./internal/modules/auth/...` — compiles clean.

- [ ] 4. Implement the minimal email-OTP fallback sender.
      New file with `EmailOTPSender` struct holding `mailer EmailSender` and
      `template string`. Constructor
      `NewEmailOTPSender(mailer EmailSender, template string) *EmailOTPSender`.
      `Enabled()` returns `s != nil && s.mailer != nil && s.mailer.Enabled()`.
      `Send(phone, code string) error`: treat `phone` as the recipient email (see
      Email-fallback decision above — DEV/testing only), subject
      `"Your verification code"`, body = template with `{code}` substituted; call
      `s.mailer.Send(phone, subject, body)`. Add a doc comment stating the
      phone-only-user limitation and that this is a dev/test fallback. Do NOT add
      any SMTP dialing — reuse the injected `EmailSender` only.
      Files: internal/modules/auth/otp_sender_email.go
      Verify: `go build ./internal/modules/auth/...` — compiles clean.

- [ ] 5. Wire OTP provider selection by config in server.go.
      Add `"strings"` to imports if not present. Replace the current block:
      `otpSender := auth.OTPSender(auth.NoopOTPSender{}); if cfg.AppEnv != "production" { otpSender = auth.NewLoggingOTPSender(log) }`
      with a switch on `strings.ToLower(strings.TrimSpace(cfg.OTPProvider))`:
        - case "twilio": if `cfg.TwilioEnabled()` →
          `auth.NewTwilioOTPSender(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber, cfg.TwilioMessagingServiceSID, cfg.OTPSMSTemplate, nil, log)`;
          else `log.Warn` "OTP_PROVIDER=twilio but Twilio not configured; falling back to dev/noop" and use the default dev/noop below.
        - case "email": if `cfg.EmailOTPEnabled()` →
          `auth.NewEmailOTPSender(emailSender, cfg.OTPSMSTemplate)` (emailSender is
          already built above this block); else warn + fall back.
        - default: existing behaviour — `LoggingOTPSender` when
          `cfg.AppEnv != "production"`, else `NoopOTPSender`.
      After selection, `log.Info("otp delivery provider selected", slog.String("provider", <name>))`
      where name is "twilio"/"email"/"logging"/"noop" — never log secrets. Keep the
      existing `auth.WithOTPSender(otpSender)` wiring unchanged.
      Files: internal/app/server.go
      Verify: `go build ./...` — compiles clean.

- [ ] 6. Add a unit test for `TwilioOTPSender.Send`.
      Table-driven test in the `auth` package using an injected `*http.Client` with
      a custom `http.RoundTripper` stub (captures the request, returns a canned
      `*http.Response`) — or `httptest.NewServer` with the client's base left as
      the real Twilio URL via a RoundTripper that routes to the test server. Prefer
      the RoundTripper stub so the real Twilio URL/path is asserted. Assert:
        - request URL path contains the account SID and ends `/Messages.json`;
        - `Authorization` header is Basic (present and decodes to accountSID:token);
        - parsed form body has `To` in E.164 with a leading `+`, has `From` (or
          `MessagingServiceSid`), and `Body` containing the substituted code;
        - 2xx canned response → `Send` returns nil;
        - 4xx JSON error response → `Send` returns a wrapped error AND the error
          string contains neither the OTP code nor the auth token.
      Add phone-normalization cases via `toE164`: `"+91 98765 43210"`→`"+919876543210"`,
      `"919876543210"`→`"+919876543210"`, `"+919876543210"`→`"+919876543210"`,
      `"0091 98765"`→`"+9198765"` (single leading +, no spaces). Do not hit the
      network.
      Files: internal/modules/auth/otp_sender_twilio_test.go
      Verify: `go test ./internal/modules/auth/...` — new + existing tests pass.

- [ ] 7. Add the Twilio block to .env and .env.example (leave MSG91 untouched).
      In `.env.example` (commented): under the OTP SMS provider section add a Twilio
      block documenting `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`,
      `TWILIO_FROM_NUMBER` (or `TWILIO_MESSAGING_SERVICE_SID`), `OTP_SMS_TEMPLATE`
      (note `{code}` is substituted), and a note: set `OTP_PROVIDER=twilio` to
      activate; set `OTP_PROVIDER=email` to use the SMTP dev fallback. Add a one-line
      caveat: a Twilio TRIAL account can only send to verified numbers, and
      production SMS to Indian `+91` numbers additionally requires India DLT
      registration (regulatory, outside this code). In `.env` add the same keys with
      EMPTY placeholder values (e.g. `TWILIO_ACCOUNT_SID=`), do NOT set real secrets,
      leave `OTP_PROVIDER=` empty, and do NOT touch the existing MSG91 /
      Razorpay / SMTP (SendGrid) values.
      Files: .env.example, .env
      Verify: `grep -n TWILIO .env .env.example` shows the new keys; MSG91 lines unchanged.

- [ ] 8. Update documentation.
      In `docs/IMPLEMENTED_FEATURES.md` (Auth / OTP section around the "Not
      implemented" note) describe pluggable OTP delivery: `OTP_PROVIDER` empty =
      dev-log in non-prod / noop in prod; `twilio` = SMS via Twilio Messages API
      (own generation/verify retained); `email` = SMTP dev fallback (note phone-only
      users may have no email). List the `TWILIO_*` env vars and the trial-number +
      India-DLT caveats. In `docs/PHASE_PLAN.md` Phase 1 auth item (the OTP ✅
      bullet) note: OTP SMS delivery via Twilio is wired behind `OTPSender`
      (pending a Twilio account/number + India DLT for +91 production), with an SMTP
      email fallback available.
      Files: docs/IMPLEMENTED_FEATURES.md, docs/PHASE_PLAN.md
      Verify: `grep -n -i twilio docs/IMPLEMENTED_FEATURES.md docs/PHASE_PLAN.md` shows the additions.

- [ ] 9. Full build + test gate.
      Run the project's build and tests to confirm nothing regressed and the new
      senders + config behave.
      Files: (none)
      Verify: `make fmt` leaves no diff on committed files beyond intended changes;
      `go build ./...` succeeds; `make test` (or `go test ./internal/config/...
      ./internal/modules/auth/...`) passes. No DB required for these packages.

## Security recap
- Never log the OTP code, the Twilio Auth Token, or any secret (dev
  `LoggingOTPSender` logging the code is intentional and unchanged).
- Twilio HTTP client is timeout-bounded (10s default) and injectable for tests.
- Twilio non-2xx handled gracefully (no panic); `RequestOTP` already returns a
  neutral/non-enumerable response, and a `Send` error surfaces as a generic
  failure without leaking provider internals.
