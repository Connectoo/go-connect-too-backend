# Implementation Plan — OTP (phone) login

Add phone + OTP login as a NEW auth path alongside the existing email/password + JWT auth.
A user POSTs their phone + role to request a 6-digit code, then POSTs phone + role + code to
verify and receive the same `AuthResponse` (JWT access/refresh + user) the password login returns.
Must work for customer and employee roles. Email/password auth is untouched. Google/OAuth is out of scope.

## Design decisions (chosen, with rationale)

- **Follow the lifecycle-token pattern, not a new mechanism.** OTP gets its own table
  (`otp_login_codes`), its own model struct, repository methods on the existing `auth.Repository`,
  an `OTPStore` interface + `WithOTPStore` option, and an `OTPSender` interface + `WithOTPSender`
  option — mirroring `LifecycleStore` / `WithLifecycleStore` and `EmailSender` / `WithEmailSender`
  exactly. Rationale: the task mandates reusing this pattern and it keeps handlers thin / service
  owning logic / repo owning DB.
- **Store a hash of the code, not plaintext.** Use the same `hashLifecycleToken(plain, secret)`
  helper (sha256 of code+secret) already used for reset/verification tokens, keyed on
  `s.lifecycleSecret`. Rationale: consistent with the existing token-hashing convention in this
  module; avoids introducing bcrypt-per-code cost and a second hashing scheme. The code lookup is
  by `(phone, role)` with hash compared in the service, so a plain sha256 (not per-row-salted bcrypt)
  is appropriate and matches the lifecycle tokens.
- **Lookup user by phone+role.** The `users` table already enforces phone-unique-per-role
  (`users_phone_role_unique`), so add `users.Repository.GetByPhoneAndRole` mirroring
  `GetByEmailAndRole`. Rationale: OTP identifies the account by (phone, role), exactly the unique key.
- **Neutral request response.** `POST /auth/otp/request` always returns the same 200 success
  regardless of whether the phone is registered (mirrors `ForgotPassword`, which returns nil when
  the user is not found). The OTP is only generated/stored/sent when a matching active user exists.
- **Rate limiting lives in the OTP store + service.** Resend cooldown and max-per-window are
  enforced by counting recent rows for (phone, role) via a repo query; the service returns
  `ErrOTPRateLimited`. Rationale: no Redis in this project (per AGENTS.md), so DB-backed counting
  is the idiomatic choice here.
- **Add one new error code `CodeTooManyRequests` → HTTP 429.** The shared errors package has no
  rate-limit code; add it so the handler can map `ErrOTPRateLimited`. Rationale: 429 is the correct
  status and the existing codes don't cover it.
- **Reuse `TokenManager.issueTokens`.** Verify success calls the existing `s.issueTokens(ctx, user)`
  — no new token logic.
- **Login status check mirrors password login.** Reject unless `user.Status == StatusActive` and
  `user.DeactivatedAt == nil`, returning `ErrUserInactive` (same as `login`).

## Config defaults (new env vars, following existing config style)

- `OTP_CODE_LENGTH` (int, default `6`)
- `OTP_CODE_TTL_MINUTES` (int, default `5`) → `OTPCodeTTL time.Duration`
- `OTP_RESEND_COOLDOWN_SECONDS` (int, default `60`) → `OTPResendCooldown time.Duration`
- `OTP_MAX_PER_WINDOW` (int, default `5`) — max requests per phone+role per window
- `OTP_WINDOW_MINUTES` (int, default `15`) → `OTPRequestWindow time.Duration`
- `OTP_MAX_ATTEMPTS` (int, default `5`) — max verify attempts per code
- `OTP_PROVIDER` (string, default `""`; `""`/`noop`/`log` → NoopOTPSender)

---

## Steps

- [ ] 1. Add the rate-limit error code to the shared errors package.
      Add `CodeTooManyRequests = "TOO_MANY_REQUESTS"` to the const block.
      Files: `internal/shared/errors/errors.go`
      Verify: `make fmt` then `go build ./...` succeeds.

- [ ] 2. Create the OTP migration (table stores a HASH of the code, never plaintext).
      Create `000028_otp_login.up.sql` with table `otp_login_codes`:
      `id UUID PK DEFAULT gen_random_uuid()`, `phone TEXT NOT NULL`, `role TEXT NOT NULL`,
      `code_hash CHAR(64) NOT NULL`, `expires_at TIMESTAMPTZ NOT NULL`,
      `attempt_count INT NOT NULL DEFAULT 0`, `consumed_at TIMESTAMPTZ`,
      `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
      Add `CREATE INDEX idx_otp_login_codes_phone_role ON otp_login_codes (phone, role);`
      and `CREATE INDEX idx_otp_login_codes_created_at ON otp_login_codes (created_at);`
      (no FK to users — request must not reveal existence; a code row may exist without matching user,
      though in this design we only insert when a user matches, keep it decoupled anyway).
      Create `000028_otp_login.down.sql`: `DROP TABLE IF EXISTS otp_login_codes;`.
      Files: `migrations/000028_otp_login.up.sql`, `migrations/000028_otp_login.down.sql`
      Verify: with Postgres up (see AGENTS.md DATABASE_URL override),
      `DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make migrate-up`
      applies cleanly; the down file drops exactly what the up creates (reviewed by reading).

- [ ] 3. Add OTP config fields and parsing in config.
      Add the fields listed in "Config defaults" to the `Config` struct and parse them in `Load()`
      following the existing `strconv.Atoi(getEnv(...))` + `time.Duration` style used for JWT TTLs.
      Add a helper `func (c *Config) OTPSenderKind() string` or just read `OTP_PROVIDER` in server.go
      (prefer keeping it a plain field `OTPProvider string`).
      Files: `internal/config/config.go`
      Verify: `go test ./internal/config/...` passes (existing `config_test.go` still green; defaults
      must not require new env vars).

- [ ] 4. Add `users.Repository.GetByPhoneAndRole` mirroring `GetByEmailAndRole`.
      Query `SELECT <userColumns> FROM users WHERE phone = $1 AND role = $2`, scan with `scanUserRow`.
      Files: `internal/modules/users/repository.go`
      Verify: `go build ./...` succeeds (full test run happens in later steps).

- [ ] 5. Add the OTP model struct.
      Add `OTPLoginCode` struct to `model.go`: `ID uuid.UUID`, `Phone string`, `Role string`,
      `CodeHash string`, `ExpiresAt time.Time`, `AttemptCount int`, `ConsumedAt *time.Time`,
      `CreatedAt time.Time`.
      Files: `internal/modules/auth/model.go`
      Verify: `go build ./...` succeeds.

- [ ] 6. Add OTP errors to the auth error set.
      Add `ErrOTPInvalid`, `ErrOTPExpired`, `ErrOTPTooManyAttempts`, `ErrOTPRateLimited`,
      `ErrOTPNotConfigured` to `errors.go`.
      Files: `internal/modules/auth/errors.go`
      Verify: `go build ./...` succeeds.

- [ ] 7. Add the OTP repository methods on `auth.Repository` (DB access only).
      In a new file `otp_repository.go`:
      - `CreateOTPCode(ctx, *OTPLoginCode) error` — INSERT.
      - `GetActiveOTPByPhoneRole(ctx, phone, role string) (*OTPLoginCode, error)` — latest
        unconsumed row for (phone, role) ordered by `created_at DESC LIMIT 1`; return
        `ErrOTPInvalid` on `sql.ErrNoRows`.
      - `IncrementOTPAttempt(ctx, id uuid.UUID) (int, error)` — `UPDATE ... SET attempt_count =
        attempt_count + 1 WHERE id = $1 RETURNING attempt_count`.
      - `MarkOTPConsumed(ctx, id uuid.UUID, at time.Time) error` — set `consumed_at` where NULL;
        0 rows → `ErrOTPInvalid`.
      - `CountOTPRequestsSince(ctx, phone, role string, since time.Time) (int, error)` — COUNT rows
        with `created_at >= since` for rate limiting.
      - `LatestOTPCreatedAt(ctx, phone, role string) (*time.Time, error)` — most recent `created_at`
        for cooldown check (nil if none).
      Follow the exact style of `lifecycle_repository.go` (ExecContext, QueryRowContext, scanner helper).
      Files: `internal/modules/auth/otp_repository.go`
      Verify: `go build ./...` succeeds.

- [ ] 8. Define the OTP service dependencies, interfaces, options, and DTOs.
      In a new file `otp_service.go` (service logic) and additions to `dto.go`:
      - `dto.go`: add `OTPRequestRequest{ Phone string `json:"phone"`; Role string `json:"role"` }`
        and `OTPVerifyRequest{ Phone string; Role string; Code string }` with json tags.
      - `otp_service.go`: define `OTPStore` interface (the six repo methods from step 7),
        `OTPSender` interface `{ Enabled() bool; Send(phone, code string) error }`, and
        `NoopOTPSender` (logs/no-ops, `Enabled() bool` returns false) matching
        `notifications.NoopPushProvider`.
      - Add service fields `otp OTPStore`, `otpSender OTPSender`, plus OTP timing/limits pulled from
        `cfg` (TTL, length, cooldown, window, maxPerWindow, maxAttempts); set them in `NewService`
        from `cfg` (default-safe) OR read from `cfg` directly in the methods.
      - Add options `WithOTPStore(OTPStore) ServiceOption` and `WithOTPSender(OTPSender) ServiceOption`
        following `WithLifecycleStore` / `WithEmailSender`.
      Files: `internal/modules/auth/otp_service.go`, `internal/modules/auth/dto.go`,
      and `internal/modules/auth/service.go` (add the two fields to the `Service` struct)
      Verify: `go build ./...` succeeds.

- [ ] 9. Implement `RequestOTP` and `VerifyOTP` service methods (business logic).
      In `otp_service.go`:
      - `RequestOTP(ctx, OTPRequestRequest) error`:
        validate phone+role non-empty (else `ErrValidation`); `nil` store → `ErrOTPNotConfigured`.
        Rate limit: if `CountOTPRequestsSince(window)` >= maxPerWindow OR last `created_at` within
        cooldown → `ErrOTPRateLimited`. Look up user via `users.GetByPhoneAndRole`; on
        `users.ErrNotFound` return `nil` (NEUTRAL — do not reveal). Only when a matching user exists
        and is active: generate a numeric code of configured length (crypto/rand), hash with
        `hashLifecycleToken(code, s.lifecycleSecret)`, insert an `OTPLoginCode` with
        `ExpiresAt = now + TTL`, and if `otpSender != nil && otpSender.Enabled()` send it (Noop in dev).
        Always return `nil` on the happy/neutral path.
      - `VerifyOTP(ctx, OTPVerifyRequest) (*AuthResponse, error)`:
        validate inputs; fetch latest active code via `GetActiveOTPByPhoneRole` (→ `ErrOTPInvalid`
        if none). If `now.After(ExpiresAt)` → `ErrOTPExpired`. If `AttemptCount >= maxAttempts` →
        `ErrOTPTooManyAttempts`. Compare `hashLifecycleToken(req.Code, secret)` to `CodeHash`; on
        mismatch call `IncrementOTPAttempt` and return `ErrOTPInvalid` (or `ErrOTPTooManyAttempts` if
        the increment reaches the max). On match: load user via `GetByPhoneAndRole`; enforce
        `Status == StatusActive && DeactivatedAt == nil` else `ErrUserInactive`; `MarkOTPConsumed`;
        then `tokens, _ := s.issueTokens(ctx, user)` and return the same `AuthResponse` shape as login.
      Add a numeric-code generator helper (crypto/rand, zero-padded to length) in `otp_service.go`.
      Files: `internal/modules/auth/otp_service.go`
      Verify: `go build ./...` succeeds; unit tests added in step 13 pass.

- [ ] 10. Add the OTP HTTP handlers and route wiring (handlers thin).
      In `handler.go`: add `otpRequest` and `otpVerify` handlers following `forgotPassword` /
      `loginCustomer` style (decodeJSON → svc call → response). `otpRequest` returns a neutral
      `response.JSON(w, 200, "If the phone is registered, an OTP was sent", nil)`.
      `otpVerify` returns `response.JSON(w, 200, "OTP login successful", res)`.
      Extend `writeServiceError` to map: `ErrOTPExpired`/`ErrOTPInvalid` → 401
      `CodeInvalidCredentials` (or `CodeInvalidToken`; pick `CodeInvalidCredentials` to match login
      "invalid" semantics, do NOT distinguish expired vs wrong to the client beyond message),
      `ErrOTPTooManyAttempts` → 429 `CodeTooManyRequests`, `ErrOTPRateLimited` → 429
      `CodeTooManyRequests`, `ErrOTPNotConfigured` → 500 `CodeInternalError`.
      In `routes.go`: add `r.Post("/otp/request", h.otpRequest)` and
      `r.Post("/otp/verify", h.otpVerify)` in the unauthenticated group (next to login routes).
      Files: `internal/modules/auth/handler.go`, `internal/modules/auth/routes.go`
      Verify: `go build ./...` succeeds.

- [ ] 11. Wire the OTP store and sender in server.go following the existing options pattern.
      Select the sender: `otpSender := auth.OTPSender(auth.NoopOTPSender{})` by default (optionally
      switch on `cfg.OTPProvider` later; only Noop/log exists now — do NOT add a 3rd-party SMS SDK).
      Pass `auth.WithOTPStore(authRepo)` and `auth.WithOTPSender(otpSender)` to the existing
      `auth.NewService(...)` call. `authRepo` already satisfies `OTPStore` via step 7.
      Note `GetByPhoneAndRole` must be reachable from the service's user store — the service uses the
      `UserStore` interface; add `GetByPhoneAndRole` to a dedicated lookup. SIMPLEST: have the OTP
      service hold a reference to the concrete `*users.Repository` via a small `OTPUserLookup`
      interface `{ GetByPhoneAndRole(ctx, phone, role string) (*users.User, error) }`, wired through
      a `WithOTPUserLookup(userRepo)` option, OR extend the `UserStore` interface with
      `GetByPhoneAndRole` (preferred: extend `UserStore`, since `*users.Repository` already implements
      it after step 4 and the mock in tests can add it easily). Choose extending `UserStore`.
      Files: `internal/app/server.go`, and `internal/modules/auth/service.go` (add
      `GetByPhoneAndRole` to the `UserStore` interface)
      Verify: `go build ./...` succeeds.

- [ ] 12. Register the two new routes in the inventory and OpenAPI spec (the parity test enforces this).
      Add `"POST /auth/otp/request"` and `"POST /auth/otp/verify"` to `ExpectedAPIRoutes`.
      Add `/auth/otp/request` and `/auth/otp/verify` path items to `openapi.yaml` copying the exact
      block shape used by `/auth/forgot-password` (post, tags: [API], summary, 200 SuccessEnvelope,
      400 ValidationError, 401 UnauthorizedError, 500 InternalError).
      Files: `internal/app/expected_routes.go`, `internal/app/spec/openapi.yaml`
      Verify: `go test ./internal/app/...` passes `TestMountedRoutesMatchExpected` and
      `TestOpenAPICoversExpectedRoutes`.

- [ ] 13. Add table-driven unit tests for the OTP service (test business logic, not handlers).
      In `otp_service_test.go`, extend the existing mocks: add `GetByPhoneAndRole` to `mockUserStore`
      (index users by phone+role) and create a `mockOTPStore` implementing `OTPStore` (in-memory
      slice/map with the six methods) and a `mockOTPSender` capturing the last sent code.
      Build the service via a helper like `newTestService` but with `WithOTPStore` + `WithOTPSender`
      and a fixed `svc.now`. Cover cases:
      success (correct code → JWT pair issued, code marked consumed);
      expired code → `ErrOTPExpired`;
      wrong code → `ErrOTPInvalid` and attempt_count incremented;
      too many attempts → `ErrOTPTooManyAttempts`;
      reuse of a consumed code → `ErrOTPInvalid`;
      rate-limited request (exceed max-per-window / within cooldown) → `ErrOTPRateLimited`;
      unknown phone on request → returns nil AND no code stored/sent (neutral);
      inactive/suspended user on verify → `ErrUserInactive`.
      Files: `internal/modules/auth/otp_service_test.go`
      Verify: `go test ./internal/modules/auth/...` — all new and existing tests pass.

- [ ] 14. Run the full format + test suite to confirm nothing regressed.
      Files: none (verification only)
      Verify: `make fmt` makes no further changes on a second run; `make test` (i.e. `go test ./...`)
      is fully green, including the route/openapi parity tests and the auth package.

- [ ] 15. Update the docs.
      In `docs/IMPLEMENTED_FEATURES.md` Auth section: add `POST /auth/otp/request` and
      `POST /auth/otp/verify` bullets and change the "Not implemented" note to drop mobile OTP
      (keep Google/OAuth as not implemented).
      In `docs/PHASE_PLAN.md`: under "Already done (✅)" note OTP login shipped, and in the Phase 1
      "Auth model decision" item mark the OTP portion done (OTP shipped; Google still open) — e.g.
      change its status marker and strike/annotate the OTP sub-bullet, leaving the Google sub-bullet.
      Files: `docs/IMPLEMENTED_FEATURES.md`, `docs/PHASE_PLAN.md`
      Verify: re-read both files; the OTP endpoints are listed and the Phase 1 auth item reflects
      OTP done / Google open. (Docs-only; `make test` already green from step 14.)

## Notes / assumptions

- No Redis/SMS SDK is wired in this project (AGENTS.md), so rate limiting is DB-backed and the only
  OTP sender shipped is the Noop/log implementation. A real SMS provider is a later wiring change
  behind the `OTPSender` interface and `OTP_PROVIDER` toggle.
- Code hashing reuses the module's existing sha256+secret token hashing (`hashLifecycleToken`) for
  consistency; this is a lookup-by-(phone,role) design, not a bcrypt-per-row design.
- Expired-vs-wrong code is not distinguished to the client beyond the message to avoid leaking
  which codes exist; both map to 401.
- Admin login via OTP is not added (admin is a web app, out of scope per the original request);
  OTP is customer + employee only, matching the `role` values accepted.
