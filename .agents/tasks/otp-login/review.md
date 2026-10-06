# Phone + OTP login alongside email/password auth

Adds a phone + OTP login path to the `auth` module as a parallel authentication route to the existing email/password flow. A client POSTs `{phone, role}` to `/auth/otp/request` to receive a 6-digit code, then POSTs `{phone, role, code}` to `/auth/otp/verify` to receive the same `AuthResponse` (JWT access/refresh pair + user) that `/auth/login/customer` returns. The design mirrors the established lifecycle-token pattern exactly: a dedicated `otp_login_codes` table, repository methods on the existing `auth.Repository`, an `OTPStore`/`OTPSender` interface pair wired via `WithOTPStore`/`WithOTPSender` options, hashed codes via the shared `hashLifecycleToken` helper, and token issuance through the existing `issueTokens`. Codes are stored only as a sha256+secret hash, single-use, expiry-enforced, attempt-capped, and request-rate-limited. The only sender shipped is a Noop (no SMS SDK dependency), consistent with AGENTS.md.

Watch for: a request-side rate-limit gap for unregistered phone numbers — codes are only persisted when an active user matches, so the DB-backed cooldown/window counters never engage for unknown phones, leaving `/auth/otp/request` unthrottled for non-existent numbers (confirmed). This does not leak registration status and does not weaken security for real accounts; it is a defense-in-depth / abuse-amplification note, relevant once a real SMS sender replaces the Noop. Everything else — hashing, single-use, expiry, attempt cap, neutral response, status gating, token reuse — is correct and well-tested.

**Verdict**: APPROVED

## High-level view

The review scope is a single commit (`270e1bd`) on top of a far-behind `main`; diffing against `main` surfaces dozens of unrelated prior features (lifecycle auth, migrations 10–27, seed tooling). The actual OTP change is clean and confined to the auth module, config, the users-repo lookup, route inventory, OpenAPI spec, one migration, and docs. Existing email/password auth is untouched.

Security posture is strong and matches the plan. Codes live only as a keyed-sha256 hash in `code_hash CHAR(64)`; verification compares hashes and never the plaintext. Codes are single-use (consumed on success, reuse rejected via the `consumed_at IS NULL` filter), expiry is enforced, and a per-code attempt counter locks the code after the configured max. OTP login is gated on the same `StatusActive && DeactivatedAt == nil` check as password login, and the JWT pair comes from the existing `TokenManager` via `issueTokens` — no new token logic.

The request endpoint is neutral: handler and service return the identical 200 response whether or not the phone is registered, and a code is only generated/stored/sent when a matching active user exists. Rate limiting (resend cooldown + max-per-window) is DB-backed by counting recent rows for `(phone, role)`, which is the right choice given no Redis in this project. The gap is that this counting only sees rows that exist, and rows exist only for registered actives — so unregistered phones bypass the throttle entirely. With the Noop sender this is inert; it becomes a real SMS-cost/abuse vector once a provider is wired.

Route parity is satisfied: both routes are in `ExpectedAPIRoutes` and in `openapi.yaml`, and the implementer's evidence shows `TestMountedRoutesMatchExpected` and `TestOpenAPICoversExpectedRoutes` passing. The migration pair is valid and the down cleanly drops exactly what the up creates. Table-driven service tests cover every required case.

<details>
<summary>Issues (2)</summary>

1. **Unregistered-phone request throttle gap** — `/auth/otp/request` only writes/counts rows for registered active users, so the cooldown and max-per-window checks never fire for unknown phones; an attacker can call the endpoint unbounded for non-existent numbers. Harmless with the Noop sender, but add an identity-independent throttle (e.g. per-IP or per-phone attempt table independent of user existence) before wiring a real SMS provider. Non-blocking.
2. **Non-constant-time hash comparison** — `VerifyOTP` compares the computed code hash to the stored hash with `!=` rather than `hmac.Equal`/`subtle.ConstantTimeCompare`. This matches the pre-existing lifecycle-token convention in this module and the risk is low (keyed sha256, not a raw secret), so it is a consistency note, not a blocker. Non-blocking.

</details>

<details>
<summary>Details</summary>

### Review scope: one commit over a stale base

`main` sits well behind the `azure` branch, so `git diff main` is dominated by prior work (lifecycle auth endpoints, migrations `000010`–`000027`, seed tooling, Docker/Dokploy config, agent skill files). The OTP feature is isolated in commit `270e1bd`, which touches 22 files: the auth module (dto, errors, handler, model, routes, service, plus new `otp_repository.go`/`otp_service.go`/`otp_service_test.go`), `internal/config/config.go`, `internal/modules/users/repository.go` (one new lookup), `internal/app/server.go` (wiring), `internal/app/expected_routes.go`, `internal/app/spec/openapi.yaml`, `internal/shared/errors/errors.go`, the `000028_otp_login` migration pair, and docs. Reviewing the commit rather than the `main` diff is what keeps scope honest here — email/password auth and all other modules are genuinely untouched.

### Hashing, single-use, expiry, attempt cap

`RequestOTP` generates a crypto/rand numeric code, hashes it with `hashLifecycleToken(code, s.lifecycleSecret)` (sha256 keyed on the lifecycle secret), and persists only the hash in `otp_login_codes.code_hash CHAR(64)`. The plaintext leaves the process only via the sender; nothing writes it to the DB or logs. `VerifyOTP` recomputes the hash and compares — the plaintext is never retrieved.

Single-use is enforced at two layers: `GetActiveOTPByPhoneRole` filters `consumed_at IS NULL ORDER BY created_at DESC LIMIT 1`, and a successful verify calls `MarkOTPConsumed`, whose `UPDATE ... WHERE id = $1 AND consumed_at IS NULL` returns `ErrOTPInvalid` on zero rows. A reused code therefore no longer matches the active-row query and fails with `ErrOTPInvalid` (covered by `TestVerifyOTPReuseConsumed`).

Expiry is checked before the hash comparison (`now.After(stored.ExpiresAt)` → `ErrOTPExpired`). The attempt cap is checked up front (`AttemptCount >= maxAttempts` → `ErrOTPTooManyAttempts`) and the counter is incremented only on hash mismatch via `IncrementOTPAttempt` (not on expiry or already-locked states), so it tracks wrong-code guesses specifically; once the increment reaches the max, even a subsequently-correct code is rejected. `TestVerifyOTPTooManyAttempts` exercises both the lockout and the correct-code-after-lockout path.

### Neutral request response and status gating

The handler returns `200 "If the phone is registered, an OTP was sent"` unconditionally. The service returns `nil` both when `users.GetByPhoneAndRole` yields `ErrNotFound` and when the user exists but is not active — so neither the HTTP status, the body, nor an error distinguishes registered from unregistered or active from suspended. `TestRequestOTPUnknownPhoneNeutral` asserts no code is stored or sent for an unknown phone; `TestVerifyOTPInactiveUser` asserts a suspended user gets no code at request time and `ErrUserInactive` at verify time. Verify maps both expired and wrong-code to the same `401 INVALID_CREDENTIALS` message, so the client cannot distinguish "wrong" from "expired" either.

Login gating matches password login: verify loads the user via `GetByPhoneAndRole` and rejects unless `Status == StatusActive && DeactivatedAt == nil`, returning `ErrUserInactive` — the same guard the password `login` path uses.

### Token issuance reuses the existing TokenManager

On a successful verify the service calls `s.issueTokens(ctx, user)` — the identical method the password login and refresh paths use — and returns `AuthResponse{User: toUserResponse(user), Tokens: *tokens}`, byte-for-byte the same shape as `/auth/login/customer`. No parallel token-minting logic was introduced. `TestVerifyOTPSuccess` asserts both tokens are present, the user role is correct, and the code is marked consumed.

### Rate-limiting gap for unregistered phones

```
RequestOTP:
  count = CountOTPRequestsSince(phone, role, now-window)   ─┐ both read otp_login_codes
  latest = LatestOTPCreatedAt(phone, role)                 ─┘ which only has rows for
  if count >= maxPerWindow -> rate limited                    registered active users
  if latest within cooldown -> rate limited
  user = GetByPhoneAndRole(...)   // ErrNotFound -> return nil (neutral)
  if !active -> return nil
  ...insert row only here...
```

Because a row is written only after the active-user check passes, the two throttle queries always return empty for an unregistered phone. The cooldown and max-per-window limits therefore protect registered accounts but impose no limit on requests for phones that don't map to an active user. An attacker enumerating or hammering random numbers hits no ceiling. With the shipped Noop sender nothing is dispatched, so today this costs only DB reads. Once a real SMS provider sits behind `OTPSender`, this path still sends nothing for unknown phones (the send is also gated on the user existing), so it is not an SMS-cost leak for unknown numbers specifically — the practical exposure is unbounded request volume against the endpoint. A throttle keyed independently of user existence (per-IP, or a request-attempt table keyed on phone regardless of match) would close it. Flagged as non-blocking defense-in-depth; the per-registered-account limits that matter for real credentials are present and tested (`TestRequestOTPRateLimited`, `TestRequestOTPMaxPerWindow`).

### OTPSender is Noop, no SMS SDK

`server.go` wires `auth.OTPSender(auth.NoopOTPSender{})`; `Enabled()` returns false and `Send` is a no-op. `RequestOTP` only calls `Send` when `otpSender != nil && otpSender.Enabled()`, so the dev/default path generates and stores a code but dispatches nothing — no hard dependency on any SMS gateway, matching AGENTS.md ("no payment gateways / Firebase wired up yet"). `OTP_PROVIDER` is parsed into config for a future toggle but only Noop exists, as intended.

### Migration pair

`000028_otp_login.up.sql` creates `otp_login_codes` with `id UUID PK DEFAULT gen_random_uuid()`, `phone`, `role`, `code_hash CHAR(64)`, `expires_at`, `attempt_count INT DEFAULT 0`, `consumed_at` (nullable), `created_at DEFAULT NOW()`, plus `idx_otp_login_codes_phone_role` and `idx_otp_login_codes_created_at`. The `(phone, role)` index backs the active-code lookup and the count/latest queries; the `created_at` index backs the window scan. The down is `DROP TABLE IF EXISTS otp_login_codes;`, which removes the table and its indexes together — a clean inverse of the up. No FK to `users`, consistent with the neutral-response design (a code row's existence is decoupled from account existence). Not re-applied against a live DB here; evaluated by reading, consistent with the implementer's note.

### Test coverage

The table-driven suite in `otp_service_test.go` uses in-memory `mockOTPStore`, `mockOTPSender`, and the extended `mockUserStore` (now phone-indexed), with a fixed `svc.now` for deterministic expiry/cooldown. Covered: success with JWT pair issued and code consumed; expired; wrong code with attempt increment; too-many-attempts lockout including correct-code-after-lockout; reuse of a consumed code; resend-cooldown and max-per-window rate limits; unknown-phone neutral (nothing stored or sent); and inactive user at both request and verify. This is the full matrix the task asked for.

Not tested: the repository SQL itself (no live-DB integration test — the mocks stand in), and the HTTP handler error-mapping (the 429/401 status mapping in `writeServiceError` is exercised only by reading, not by a handler test). Both are acceptable given the project's test conventions and the plan's "test business logic, not handlers" directive, but the handler status mapping for the new 429 code is unverified by an automated test.

</details>

<details>
<summary>File map</summary>

- `internal/modules/auth/otp_service.go` — new: `OTPStore`/`OTPSender` interfaces, `NoopOTPSender`, `RequestOTP`/`VerifyOTP`, numeric code generator, config accessors.
- `internal/modules/auth/otp_repository.go` — new: six `otp_login_codes` DB methods + scanner, mirroring `lifecycle_repository.go`.
- `internal/modules/auth/otp_service_test.go` — new: table-driven service tests + in-memory mocks.
- `internal/modules/auth/service.go` — add `GetByPhoneAndRole` to `UserStore`; add `otp`/`otpSender` fields and `WithOTPStore`/`WithOTPSender` options.
- `internal/modules/auth/handler.go` — add `otpRequest`/`otpVerify` handlers and 401/429 error mapping.
- `internal/modules/auth/routes.go` — register `POST /auth/otp/request` and `/auth/otp/verify` (unauthenticated group).
- `internal/modules/auth/dto.go` — `OTPRequestRequest`/`OTPVerifyRequest`.
- `internal/modules/auth/model.go` — `OTPLoginCode` struct.
- `internal/modules/auth/errors.go` — OTP error values.
- `internal/modules/users/repository.go` — `GetByPhoneAndRole`.
- `internal/config/config.go` — OTP env vars + parsing (default-safe).
- `internal/app/server.go` — wire Noop sender + OTP store into `auth.NewService`.
- `internal/app/expected_routes.go` — add two routes to the inventory.
- `internal/app/spec/openapi.yaml` — add two path items (copied from `/auth/forgot-password` shape).
- `internal/shared/errors/errors.go` — `CodeTooManyRequests`.
- `migrations/000028_otp_login.up.sql` / `.down.sql` — create/drop `otp_login_codes`.
- `docs/IMPLEMENTED_FEATURES.md`, `docs/PHASE_PLAN.md` — doc updates.

Full diff: `git show 270e1bd` (the OTP feature commit on branch `azure`).

</details>
