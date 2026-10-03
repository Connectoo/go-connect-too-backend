# Verification — Phone-only (OTP) Registration

Branch: `azure`. First iteration (no `review.json` existed).

## Commands run (from repo root) and results

1. `gofmt -l .` → **no output** (all files formatted; clean).
2. `go build ./...` → **success**. This was the main risk: `users.User.Email` and
   `users.User.PasswordHash` became `*string`, touching every call site. All compile.
3. `go test ./...` → **all packages pass** (`ok` for every package with tests;
   remaining packages have no test files). Includes:
   - `internal/app` route/OpenAPI parity test (`routes_openapi_test.go`) — green;
     no routes were added so `expected_routes.go` / `openapi.yaml` needed no changes.
   - `internal/modules/auth` — green, including the new tests below.
   - `internal/modules/users`, `internal/modules/admin`, `internal/modules/employees`,
     `internal/modules/kyc` — green (nullable-email call sites updated).
4. `go vet ./...` → clean (used to surface the `*string` test construction errors
   during development; all fixed).

### New / changed auth tests (ran with `-v`, all PASS)
- `TestVerifyOTPAutoCreatesPhoneOnlyUser` — unregistered phone → request → verify
  auto-creates a phone-only user (email nil, password_hash nil, status active),
  profile row created via the mock registrar, tokens issued, code consumed,
  `res.User.Email == ""`.
- `TestVerifyOTPExistingUserUnchanged` — seeded active user still logs in; no
  duplicate account created.
- `TestRequestOTPUnknownPhoneSendsCodeAndRateLimits` — replaces the old
  "unknown phone sends nothing" test: an unregistered phone now DOES get a code
  (precondition for auto-create), the response stays neutral (nil), and a rapid
  second request still trips `ErrOTPRateLimited` (anti SMS-bomb for unregistered
  phones — the prior review gap).
- `TestVerifyOTPInactiveUser` — an EXISTING suspended account still gets a neutral
  no-send on request and `ErrUserInactive` on verify (disabled accounts cannot be
  OTP-logged-in).
- `TestPasswordLoginRejectsPhoneOnlyAccount` — password login against a phone-only
  account (nil password hash) returns `ErrInvalidCredentials`, never a nil-deref/500.

## Migration (not applied — no DB in sandbox)

- Added paired `migrations/000029_phone_only_users.up.sql` / `.down.sql`.
  - Up: `ALTER TABLE users ALTER COLUMN email DROP NOT NULL;` and same for
    `password_hash`. Keeps `users_email_role_unique` and `users_phone_role_unique`;
    relies on Postgres treating NULLs as distinct in a UNIQUE index (multiple
    phone-only rows with NULL email coexist) — no constraint added to reject NULL
    emails.
  - Down: re-adds `SET NOT NULL` on both columns, with a comment that phone-only
    rows must be backfilled first (acceptable for dev/local rollback).
- **Skipped live migrate up/down**: Docker was not running (port 5433 closed) and
  the `migrate` CLI is not installed in this environment. Something is listening on
  5432, but per AGENTS.md that does not match the project's docker-compose DB
  (`app:app@5433`), so applying migrations there was avoided to prevent touching an
  unrelated database. The SQL is plain, reversible `ALTER COLUMN ... DROP/SET NOT NULL`
  and was reviewed by hand. The Go build/test suite does not require the migration to
  be applied.

## Summary of code changes

- `internal/modules/users/model.go` — `Email` and `PasswordHash` are now `*string`.
  Repository insert/scan already pass these by reference, so `database/sql` writes/
  reads NULL for nil pointers (no `sql.NullString` needed).
- `internal/modules/auth/otp_service.go`:
  - `RequestOTP` now generates/stores/sends a code for any phone+role after the
    rate-limit checks (so unregistered phones are rate-limited and can be
    auto-created); only an EXISTING disabled account is silently skipped.
  - `VerifyOTP` auto-creates a phone-only account via the registrar when no user
    exists for the phone+role; duplicate (concurrent verify) falls back to loading
    the existing user via `GetByPhoneAndRole`. New `createPhoneOnlyUser` helper.
- `internal/modules/auth/service.go` — `register` sets `*string` email/password;
  `login` uses `derefOr(user.PasswordHash, "")` so phone-only users fail bcrypt
  cleanly; `toUserResponse` dereferences email; added `derefOr` helper;
  `UserResponse.Email` tagged `omitempty`.
- `internal/modules/auth/lifecycle_service.go` — nil-safe email/password deref.
- `internal/modules/workers/user_email_adapter.go`, `internal/modules/admin/service.go`,
  `internal/modules/users/service.go` — nil-safe email mapping (`derefOr`).
- `internal/modules/employees/admin_repository.go`, `internal/modules/kyc/repository.go`
  — `u.email` → `COALESCE(u.email, '')` so a phone-only employee's NULL email scans
  into the non-nullable Go string without error.
- Tests updated for the `*string` model (auth, users, admin) plus the new cases above.
- Docs: `docs/IMPLEMENTED_FEATURES.md` and `docs/PHASE_PLAN.md` updated.

No new routes; request bodies unchanged, so OpenAPI/parity untouched.
