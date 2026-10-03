# Implementation Plan — Phone-only (OTP) Registration

Goal: let accounts be created with ONLY a phone number. The mobile path becomes
`enter phone -> request OTP -> verify OTP`; on first successful verify for an
unregistered phone+role, the account is auto-created and tokens are issued.
Email/password registration and login keep working unchanged (parallel path).

## Design decisions (made during exploration)

- **Auto-create on verify, no new route.** `RequestOTP` already exists and will be
  changed to send a code even for unregistered phones; `VerifyOTP` will auto-create
  the account on first valid verify. A dedicated `POST /auth/register/phone` adds a
  route + openapi + expected_routes churn for no behavioral gain, so we do NOT add
  one. Rationale: the task says add a route only if cleaner; auto-create reuses the
  existing two-endpoint flow the mobile apps already call. **No route changes ⇒ no
  `expected_routes.go` change and no openapi path/parity change.**
- **Nullable representation.** `users.User.Email` and `users.User.PasswordHash`
  become `*string`. Rationale: lets scan/insert pass NULL directly via `database/sql`
  (a `*string` nil scans/writes NULL), and keeps `UserResponse.Email` omitempty JSON
  stable. Call sites that read these are few (verified below) and mechanical.
- **Password login against phone-only accounts.** `security.CheckPassword` is called
  with the stored hash; for a phone-only user the hash is nil/empty, so bcrypt returns
  a mismatch error, which `login`/`ChangePassword` already map to `ErrInvalidCredentials`.
  We only need to guard against dereferencing a nil `*string` — pass `""` when nil so
  bcrypt fails cleanly rather than panicking. No 500.
- **Request-side rate limit for unregistered phones.** Current `RequestOTP` runs the
  rate-limit checks (cooldown + max-per-window) BEFORE the user lookup, then returns
  neutral `nil` for unknown phones WITHOUT storing a code — so repeated requests for an
  unregistered phone never increment the count and the limit never engages (prior review
  flag). Fix: for phone-only auto-create, an unregistered phone must now actually get a
  code generated/stored/sent, so the per-phone+role count increments naturally and the
  limit engages. Keep the neutral response shape (same message, no enumeration).
- **Concurrency / idempotency.** Auto-create relies on `users_phone_role_unique`. If a
  concurrent verify already created the row, `CreateInTx` returns `users.ErrDuplicatePhone`;
  handle it by re-loading the existing user via `GetByPhoneAndRole` and proceeding to
  issue tokens.
- **Default name.** Phone-only users are created with an empty `Name` (`""`). The users
  table keeps `name NOT NULL`; `""` satisfies it. A later profile-edit flow can set it.
  (Optional `name` field could be threaded from `OTPVerifyRequest`, but keep minimal:
  empty name, do not add request fields.)

## Verification (run after each step unless noted; all must pass at the end)

- `gofmt -l .` — prints nothing (no unformatted files). `make fmt` to auto-fix.
- `go build ./...` — compiles clean.
- `go test ./...` — all tests pass. (`make test` runs the configured package set.)

Migrations are only exercised against a live DB; the plan does not require applying
them for the Go build/test to pass. If a DB is available, verify with
`DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make migrate-up`
then `make migrate-down`/re-up to confirm the pair is reversible (see AGENTS.md).

---

- [ ] 1. Add migration `000029_phone_only_users` to make `users.email` and
      `users.password_hash` nullable.
      Up: `ALTER TABLE users ALTER COLUMN email DROP NOT NULL;` and
      `ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;`. Add a SQL comment
      noting that `users_phone_role_unique` and `users_email_role_unique` are kept, and
      that Postgres treats NULLs as distinct in a UNIQUE index, so multiple phone-only
      rows with NULL email coexist (desired) — do NOT add any constraint rejecting NULL
      emails. Down: re-add NOT NULL: `ALTER TABLE users ALTER COLUMN email SET NOT NULL;`
      and `ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;`, with a comment
      that down requires backfilling email/password_hash for phone-only rows first
      (acceptable for a dev/rollback op).
      Files: migrations/000029_phone_only_users.up.sql, migrations/000029_phone_only_users.down.sql
      Verify: `go build ./...` still passes (no Go impact yet). If DB available, run
      migrate-up then migrate-down and confirm no error.

- [ ] 2. Change `users.User.Email` and `users.User.PasswordHash` to `*string`.
      Files: internal/modules/users/model.go
      Verify: `go build ./...` will fail until steps 3–8 land — expected; this step is
      the type change only. Build after step 8.

- [ ] 3. Update users repository insert/update/scan for the nullable columns.
      In `CreateInTx` pass `user.Email` and `user.PasswordHash` directly (they are now
      `*string`, which `database/sql` writes as NULL when nil). `scan.go`
      `scanUserRow` already scans into `&user.Email` / `&user.PasswordHash`; with
      `*string` fields these now scan NULL safely (no `sql.NullString` needed because the
      field is a pointer). `userColumns` in lifecycle_repository.go is unchanged (same
      column list). Confirm `UpdatePassword` still writes a plain string param — it takes
      a `string passwordHash` argument, unaffected.
      Files: internal/modules/users/repository.go, internal/modules/users/scan.go
      Verify: build after step 8.

- [ ] 4. Update auth DTO/response + mapping for nullable Email.
      `UserResponse.Email` stays `string` with `omitempty`? It cannot hold nil; change
      `toUserResponse` (service.go) to dereference `user.Email` when non-nil, else `""`.
      Keep `UserResponse.Email string json:"email"` (empty string omitted only if tagged
      omitempty — add `omitempty` to keep phone-only responses clean). Mirror the same
      nil-safe deref in `internal/modules/users/service.go` `toProfileResponse` and
      `internal/modules/admin/service.go` `toUserResponse`.
      Files: internal/modules/auth/dto.go (add `omitempty` to UserResponse.Email),
      internal/modules/auth/service.go (toUserResponse), internal/modules/users/service.go
      (toProfileResponse), internal/modules/admin/service.go (toUserResponse)
      Verify: build after step 8.

- [ ] 5. Fix the two auth call sites that read `user.Email` / `user.PasswordHash`
      directly so they are nil-safe.
      - service.go `register`: it SETS `Email` and `PasswordHash`; change to assign
        pointers — build the email string then take its address, and store a non-nil
        `*string` password hash. (Email/password register always has both.)
      - service.go `login`: `security.CheckPassword(user.PasswordHash, req.Password)` —
        pass `derefOr(user.PasswordHash, "")` so a phone-only user (nil hash) fails bcrypt
        → `ErrInvalidCredentials`, never a nil deref.
      - lifecycle_service.go `ChangePassword`: same nil-safe deref for `CheckPassword`.
      - lifecycle_service.go `ForgotPassword` / `ResendVerification`: `s.mailer.Send(user.Email, ...)`
        — these only run for email/password users located by `GetByEmailAndRole`, but pass
        `derefOr(user.Email, "")` to be safe.
      Add a small unexported helper `derefOr(p *string, fallback string) string` in
      service.go (or reuse in both files via one definition).
      Files: internal/modules/auth/service.go, internal/modules/auth/lifecycle_service.go
      Verify: build after step 8.

- [ ] 6. Fix the workers email adapter to be nil-safe.
      `internal/modules/workers/user_email_adapter.go` `GetByID` returns `user.Email`;
      change to return `derefOr(user.Email, "")` (define a tiny local helper or inline
      the nil check). This path emails users; a nil email returns `""` and the caller’s
      Enabled()/empty-recipient guard handles it.
      Files: internal/modules/workers/user_email_adapter.go
      Verify: build after step 8.

- [ ] 7. Make the admin/KYC listing SQL NULL-safe for `u.email`.
      Both `internal/modules/employees/admin_repository.go` and
      `internal/modules/kyc/repository.go` scan `u.email` into a non-nullable Go
      `string` (`AdminListItem.UserEmail` / the kyc list item). A phone-only EMPLOYEE
      would make these scans fail on NULL. Change the two SELECTs from `u.email` to
      `COALESCE(u.email, '') AS email` so NULL scans as `""`. No Go struct change needed.
      Files: internal/modules/employees/admin_repository.go, internal/modules/kyc/repository.go
      Verify: build after step 8; `go test ./internal/modules/employees/... ./internal/modules/kyc/...`.

- [ ] 8. Implement auto-create + unregistered-phone OTP send in the auth OTP service.
      In `RequestOTP` (otp_service.go): remove the "neutral return when user not found /
      inactive" early-outs that skip code generation; instead ALWAYS generate+store+send
      a code after the rate-limit checks pass, for any phone+role, so the per-phone+role
      rate limit engages for unregistered phones. Keep the handler's neutral response.
      (Keep the StatusActive gate only for an EXISTING user that is suspended/inactive —
      for those, still return neutral `nil` WITHOUT sending, so a disabled account can't
      be OTP-logged-in; a brand-new phone has no such row and proceeds.)
      In `VerifyOTP` (otp_service.go): after the code hash matches and before issuing
      tokens, load the user via `GetByPhoneAndRole`; on `users.ErrNotFound` CREATE a
      phone-only user via the registrar:
        - build `*users.User{ ID: uuid.New(), Name: "", Email: nil, Phone: &phone,
          PasswordHash: nil, Role: role, Status: StatusActive, CreatedAt/UpdatedAt: now }`
        - call `s.registrar.RegisterCustomer` or `RegisterEmployee` by role (default →
          `ErrValidation`), so the customers/employees profile row is created in the same
          tx.
        - if the registrar returns `users.ErrDuplicatePhone` (concurrent verify), re-load
          with `GetByPhoneAndRole` and continue.
      Then run the existing status check, `MarkOTPConsumed`, `issueTokens`, and return
      `AuthResponse`. Keep OTP hashing/single-use/expiry/attempt-cap untouched. Do not log
      the code.
      Files: internal/modules/auth/otp_service.go
      Verify: `go build ./...` passes now (all type changes resolved);
      `go test ./internal/modules/auth/...`.

- [ ] 9. Update the OTP service tests for the new behavior.
      In otp_service_test.go:
      - Update `seedOTPUser` to set `Email` as a `*string` (helper `strptr("otp@example.com")`)
        and `PasswordHash` nil (or a pointer) to match the new model.
      - Fix `TestRequestOTPUnknownPhoneNeutral`: an unknown phone now DOES generate and
        send a code (that is the new auto-create precondition). Rename/retarget it to
        assert that an unknown phone returns nil AND a code is stored/sent, and that the
        rate limit still engages on a rapid second request (`ErrOTPRateLimited`).
      - Keep `TestVerifyOTPInactiveUser` behavior for an EXISTING suspended user (still
        neutral on request; `ErrUserInactive` on verify when flipped).
      - Add `TestVerifyOTPAutoCreatesPhoneOnlyUser`: no seeded user → RequestOTP →
        VerifyOTP with the sent code → assert a user now exists via
        `store.GetByPhoneAndRole`, tokens are issued, `res.User.Role == role`,
        `res.User.Email == ""`, and the code is marked consumed. Assert the registrar
        created a profile (extend `mockRegistrar` to record which role was registered, or
        assert via `store.byPhone`).
      - Add `TestVerifyOTPExistingUserUnchanged`: with a seeded active user, verify still
        returns that user and does not create a duplicate.
      Follow the existing table-driven / helper style.
      Files: internal/modules/auth/otp_service_test.go
      Verify: `go test ./internal/modules/auth/...` passes.

- [ ] 10. Update password-login-vs-phone-only test and the shared auth test mocks for
      the nullable model.
      In service_test.go:
      - Update `mockUserStore.Create` and `userStoreKey` usage: `user.Email` is now
        `*string`; key off `derefOr(user.Email, "")` so a phone-only user (nil email)
        keys on `""`+role without colliding across phones. Index phone-only users by
        phone in `byPhone` (already handled when `copy.Phone != nil`).
      - Update register-path tests (`TestRegisterCustomerSuccess`, etc.) if they read
        `res.User.Email` — they assert role/tokens, so likely only the seed/create mock
        needs the pointer change.
      - Add `TestPasswordLoginRejectsPhoneOnlyAccount`: seed a phone-only user directly in
        the mock (nil Email, nil PasswordHash, active) under some email key, then call
        `LoginCustomer` with any password and assert `ErrInvalidCredentials` (never a
        panic / 500).
      Files: internal/modules/auth/service_test.go
      Verify: `go test ./internal/modules/auth/...` passes.

- [ ] 11. Sweep remaining users-module and cross-module tests for the `*string` change.
      Any test constructing `users.User{Email: "x", PasswordHash: "y"}` as plain strings
      must switch to pointers (add a `strptr` test helper where needed). Likely spots:
      users module tests, admin module tests. Build-driven: fix whatever `go build ./...`
      and `go vet ./...` flag.
      Files: (as surfaced) internal/modules/users/*_test.go, internal/modules/admin/*_test.go,
      internal/modules/employees/service_test.go if it builds a users.User
      Verify: `go build ./...` and `go test ./...` both pass.

- [ ] 12. Full verification pass.
      Files: none
      Verify: `gofmt -l .` prints nothing; `go build ./...` clean; `go test ./...` all
      green. If a DB is available, apply `make migrate-up` with the AGENTS.md DATABASE_URL
      override and confirm the server boots (`make run` + `curl /api/v1/health`).

- [ ] 13. Docs: update IMPLEMENTED_FEATURES.md Auth section.
      Describe phone-only registration via OTP (account auto-created on first successful
      `POST /auth/otp/verify` for an unregistered phone+role; tokens issued), and state
      that email/password register + login remain a parallel path (admin and existing
      users). Note request-side rate limiting now engages for unregistered phones.
      Files: docs/IMPLEMENTED_FEATURES.md
      Verify: manual read — Auth section mentions auto-create-on-verify and the parallel
      email/password path.

- [ ] 14. Docs: update PHASE_PLAN.md Phase 1 auth item.
      Mark phone-only OTP registration + login as done (auto-create on verify); keep
      Google/OAuth listed as open.
      Files: docs/PHASE_PLAN.md
      Verify: manual read — Phase 1 "Auth model decision" reflects OTP registration done,
      Google still open.

## Notes / assumptions

- OpenAPI/route parity: no routes added, so `internal/app/expected_routes.go` and
  `internal/app/spec/openapi.yaml` need NO path changes. The OTP request bodies in the
  spec are summary-only (no requestBody schema), and `RegisterRequest`/`UserResponse`
  schemas describe the email/password path, which is unchanged — leaving them as-is keeps
  `routes_openapi_test.go` green. (Optional: add a `nullable: true` to `UserResponse.email`
  in openapi for accuracy; not required for tests.)
- `security.CheckPassword` is never weakened; phone-only accounts simply have no hash and
  therefore fail password login by design.
- Keep handlers thin (no handler changes needed — DTOs and routes are unchanged); all new
  logic lives in the auth service, per AGENTS.md layering.
