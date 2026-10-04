# Verification — Twilio SMS OTP sender (+ email fallback)

Branch: `azure`. No push. Twilio Go SDK NOT added (net/http only).

## Commands run (from repo root) and results

### 1. `gofmt -l .`
Reports nothing (clean). Formatted the touched files with `gofmt -w` beforehand:
`internal/config/config.go`, `internal/config/config_test.go`,
`internal/modules/auth/otp_sender_twilio.go`,
`internal/modules/auth/otp_sender_email.go`,
`internal/modules/auth/otp_sender_twilio_test.go`,
`internal/app/server.go`.

### 2. `go build ./...`
Succeeds (`build OK`).

### 3. `go test ./...`
All packages pass; no failures and no DB-dependent tests were skipped. Includes:
- `internal/config` — `TestTwilioEnabled` (table-driven true/false), `TestEmailOTPEnabled`,
  `TestLoadOTPSMSTemplateDefault` (default contains `{code}`).
- `internal/modules/auth` — `TestToE164` (`+91…`, `91…`, spaces, `0091…` → single `+`,
  no spaces) and `TestTwilioOTPSenderSend`:
  - success → correct URL path (contains account SID, ends `/Messages.json`),
    Basic auth decodes to accountSID:token, form body has `To` in E.164 with `+`,
    `From`, and `Body` with the code substituted; 2xx ⇒ nil.
  - MessagingServiceSid preferred over From when set.
  - Twilio 4xx JSON error ⇒ wrapped error that contains neither the OTP code nor
    the auth token (asserts no secret leakage). No real network — custom
    `http.RoundTripper` stub.
- `internal/app` — route/OpenAPI parity test passes.

Targeted run of the new tests (`-run 'Twilio|EmailOTPEnabled|OTPSMSTemplate|ToE164' -v`)
all PASS.

### 4. go.mod / go.sum unchanged
`git status --short go.mod go.sum` → no output. `grep -i twilio go.mod go.sum` →
no matches. No new third-party module for Twilio.

## Env / docs checks
- `.env` gains empty Twilio placeholders + `OTP_SMS_TEMPLATE`; `OTP_PROVIDER=` left
  empty; MSG91 / Razorpay / SMTP values untouched.
- `.env.example` Twilio block + `OTP_PROVIDER=email` note present.
- `docs/IMPLEMENTED_FEATURES.md` and `docs/PHASE_PLAN.md` document pluggable OTP
  delivery, Twilio env vars, and the trial-number + India-DLT caveats.

## Security notes
- OTP code, Twilio auth token, and other secrets are never logged. Twilio failure
  log carries only HTTP status + Twilio error code. The returned error omits the
  code and token (asserted by test).
- HTTP client is timeout-bounded (10s default) and injectable for tests.
- `RequestOTP` neutral/non-enumerable behavior and OTP generation/verification are
  unchanged; Twilio is a transport only (no Twilio Verify).
