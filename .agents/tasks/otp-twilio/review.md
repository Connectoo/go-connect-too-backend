# Twilio SMS OTP sender + SMTP email fallback behind OTPSender

Adds two new OTP delivery transports behind the existing `auth.OTPSender` interface: a `TwilioOTPSender` that POSTs to the Twilio Messages REST API over plain `net/http` (Basic auth, form body), and an `EmailOTPSender` that reuses the already-wired SMTP sender as a dev/testing fallback. The active transport is chosen in `internal/app/server.go` by `cfg.OTPProvider` (`twilio` / `email` / otherwise the dev `logging` sender in non-prod or `noop` in prod). Config grows the Twilio fields plus `OTP_SMS_TEMPLATE`, with `TwilioEnabled()` / `EmailOTPEnabled()` helpers; `.env.example` gains a documented Twilio block. OTP generation, hashing, and verification are untouched — this is a transport-only change, and no Twilio Go SDK was added.

Watch for: the Twilio send-failure path returns a 500 to the API caller rather than the neutral 200, which does not leak registration status but is a slight behavioral asymmetry (confirmed); the working-tree `.env` holds real third-party secrets (Razorpay/SMTP/FCM) but those are pre-existing and outside this commit (confirmed).

**Verdict**: APPROVED

## High-level view

The transport is selected once at startup and injected via `WithOTPSender`, so the service's `RequestOTP`/`VerifyOTP` logic is unchanged; the sender is a pure delivery seam exercised through the unchanged `Enabled()`/`Send(phone, code)` contract. When a provider is named but not configured, wiring logs a warning and falls through to the dev/noop sender rather than failing startup.

The Twilio sender is careful with secrets. The OTP code and the auth token are never logged; failure logs carry only the HTTP status and Twilio's own error code, and the returned error omits both the code and the token (asserted by test). The HTTP client is timeout-bounded (10s) and injectable for tests, non-2xx responses are parsed defensively with a bounded reader, and no panic path exists.

Phone numbers are normalized to E.164 with a single leading `+` (spaces stripped, any `+`/`00` prefix dropped then one `+` prepended) — deliberately the opposite of the MSG91 no-plus convention. Template substitution replaces `{code}`; a Messaging Service SID takes precedence over a from-number when both are set.

The email fallback is explicitly dev/testing only: the `OTPSender.Send` signature carries just the phone value, so the email sender treats that value as the destination address. Real phone-only users have no email and cannot be reached this way — documented in the type comment, not a silent gap.

Scope is tight. The commit touches only the OTP sender files, config, server wiring, env example, and docs; `go.mod`/`go.sum` carry no Twilio dependency.

<details>
<summary>Issues (2)</summary>

1. **Send-failure returns 500, not neutral 200** — a Twilio/email delivery error propagates to `writeServiceError`'s default case and surfaces as a generic 500, unlike the neutral 200 used elsewhere in `otpRequest`. It does not reveal whether the phone is registered, so it is non-blocking, but a transport outage is distinguishable from success. Optionally treat delivery failure as a logged-but-neutral response if callers must never distinguish the two. (confirmed)
2. **Real secrets in working-tree `.env`** — the untracked `.env` contains live Razorpay keys, a SendGrid SMTP API key, and a full Firebase private key. The Twilio fields are correctly empty and the file is not part of this commit (base `.env` is minimal/deleted in the diff), so this is out of scope for the Twilio task, but the live secrets should be rotated and kept out of any future commit. (confirmed)

</details>

<details>
<summary>Details</summary>

## Secret handling in the Twilio sender

The sender logs a phone-only debug breadcrumb on success and, on failure, a warning carrying only `status` and `twilio_code` — never the OTP code or the auth token. The returned error is `"twilio send failed: status %d: %s (code %d)"` built from the HTTP status and Twilio's own `message`/`code` fields parsed from the response body; it does not interpolate the token or the OTP. The accompanying test asserts the error string contains neither the code nor the token, so the no-leak property is pinned by a regression test rather than left to inspection. Basic auth is set via `req.SetBasicAuth`, so the token lives only in the request header and never reaches a log or error string.

The non-2xx branch reads the body through `io.LimitReader(resp.Body, 4096)` and ignores the unmarshal error, so a malformed or oversized Twilio error body cannot panic or blow up memory; worst case the Twilio code/message come back zero/empty and the error still reports the HTTP status. The client defaults to `&http.Client{Timeout: 10 * time.Second}` when none is injected, so a hung Twilio endpoint cannot stall the request indefinitely.

## E.164 normalization

`toE164` strips spaces, trims a leading `+`, then trims a leading `00`, then prepends exactly one `+`. The table test covers already-E.164, no-plus, spaced-with-plus, and `00`-prefixed inputs. Trimming `+` before `00` means `+0091…` collapses correctly to `+91…`. The one theoretical sharp edge is a number whose significant digits genuinely begin `00` after the country code, which would be mis-trimmed — but stored phones already carry a country code, so the leading segment is never bare subscriber digits. This matches the spec's note that the convention is intentionally the inverse of MSG91.

## Provider selection and the unchanged call-site

The switch in `server.go` lower-cases and trims `OTPProvider`, instantiates Twilio only when `TwilioEnabled()` (SID + token + either from-number or messaging-service SID) and email only when `EmailOTPEnabled()` (SMTP host set). A provider named but unconfigured logs a warning and leaves `otpSender` nil, which then resolves to the dev `logging` sender outside production or `noop` in production — matching the spec's fall-back-with-warning requirement. The selected provider name is logged (not any secret). `RequestOTP` still calls `s.otpSender.Send(phone, code)` only when `Enabled()` is true, and its rate-limiting, hashing, and neutral-response contract are byte-for-byte unchanged.

A delivery error from `Send` propagates out of `RequestOTP` and lands in `writeServiceError`'s default branch: the real error is logged server-side and the caller gets a generic 500 "Something went wrong". No provider internals or secrets reach the client. The only caveat is the status asymmetry noted in Issues — a 500 on delivery failure versus the neutral 200 on success — which does not enable account enumeration.

## Test coverage

The Twilio sender is tested with an injected `http.RoundTripper` stub (no network): success asserts the request path contains the account SID and ends in `/Messages.json`, Basic auth decodes to SID:token, and the form body carries E.164 `To`, `From`, and a `Body` containing the code; a second case asserts `MessagingServiceSid` is preferred over `From`; a 4xx case asserts a wrapped error that leaks neither the code nor the token and mentions the HTTP status. `toE164` has its own table test. Config has `TestTwilioEnabled`, `TestEmailOTPEnabled`, and the default-template test.

Not tested: the `email` provider's `Send` path and the server-wiring switch (fallback-with-warning behavior) have no direct unit test — both are simple enough to read by inspection, but the fall-through selection logic is only verified manually. `go vet ./internal/modules/auth/ ./internal/config/` was run as a narrow spot-check and is clean; the full build/test evidence in `verification.md` was accepted per instructions and not re-run.

<details>
<summary>Files changed (full diff: `git show HEAD`)</summary>

- `internal/modules/auth/otp_sender_twilio.go` — new Twilio REST sender + `toE164`.
- `internal/modules/auth/otp_sender_twilio_test.go` — RoundTripper-stubbed tests, no-leak assertions.
- `internal/modules/auth/otp_sender_email.go` — SMTP-backed dev/testing OTP fallback.
- `internal/config/config.go` — Twilio fields, `OTP_PROVIDER`/`OTP_SMS_TEMPLATE`, `TwilioEnabled()`/`EmailOTPEnabled()`.
- `internal/config/config_test.go` — config helper tests.
- `internal/app/server.go` — provider-selection switch wiring the sender.
- `.env.example` — documented Twilio block (no secrets).
- `docs/IMPLEMENTED_FEATURES.md`, `docs/PHASE_PLAN.md` — pluggable OTP delivery + Twilio/DLT caveats.
- `.agents/tasks/otp-twilio/plan.md`, `verification.md` — task artifacts.

</details>

</details>
