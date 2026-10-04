# Phone-only (OTP) registration via auto-create on verify

Extends the existing phone+OTP login path so an account can be created with nothing but a phone number: `RequestOTP` now issues a code for any phone+role (not just registered actives), and `VerifyOTP` auto-provisions a phone-only account on the first valid code when none exists, then issues the standard token pair. To support accounts with no email or password, `users.User.Email` and `users.User.PasswordHash` become `*string`, with every read site made nil-safe (`derefOr` in Go, `COALESCE(u.email,'')` in the admin/KYC SQL). Migration 000029 drops the `NOT NULL` on those two columns. Email/password registration and login stay on their existing path unchanged.

Watch for: nothing blocking. The one closed-loop win worth noting — the unregistered-phone SMS-bomb gap flagged in the prior OTP-login review is now fixed: rate-limit checks run before the user lookup, so an unknown phone increments the per-phone+role counters and trips the cooldown/window limits (confirmed). Phone-only password login fails cleanly with `ErrInvalidCredentials` via an empty-hash bcrypt compare, not a nil deref (confirmed). One minor non-blocking inconsistency: the attempt-cap counter still writes rows nowhere relevant for the disabled-account skip, but that path is correct by design.

**Verdict**: APPROVED

## High-level view

The design choice is to add no new route. `RequestOTP`/`VerifyOTP` already existed from the OTP-login feature; this change relaxes `RequestOTP` to send codes for unregistered phones and teaches `VerifyOTP` to auto-create on first verify. That keeps route inventory, `expected_routes.go`, and the OpenAPI path set unchanged — only request-body schemas were added to the spec so Swagger renders the OTP fields.

The nullable-column change is the widest-reaching part. `Email` and `PasswordHash` on `users.User` go from `string` to `*string`; `database/sql` writes/reads NULL for nil pointers through the existing insert and scan paths with no `sql.NullString` plumbing. The risk was call-site breakage, and the recorded `go build ./...` success is the evidence that every reader was updated — the two admin/KYC list queries that scan `u.email` into a non-nullable Go string were switched to `COALESCE(u.email,'')` so a phone-only employee row doesn't fail the scan.

Security posture holds. Phone-only accounts have a nil password hash; password login passes `derefOr(hash,"")` to bcrypt, which mismatches and maps to `ErrInvalidCredentials` — never a panic or 500. OTP codes remain keyed-sha256 hashed, single-use, expiry-checked, and attempt-capped exactly as before; none of that was weakened. The SMS-bomb gap from the prior review is closed because the throttle no longer depends on a user row existing.

Auto-create is transaction-safe. It goes through the same `AccountRegistrar` the email/password register path uses, which wraps the user insert and the role profile insert in one `RunInTx`. Concurrent verifies are made idempotent by `users_phone_role_unique`: a losing racer gets `ErrDuplicatePhone` and reloads the existing user instead of erroring.

Migration 000029 is a clean paired `DROP NOT NULL` / `SET NOT NULL` with the required caveat that the down needs phone-only rows backfilled first. Not applied against a live DB (none available in the sandbox); evaluated by reading, consistent with the plan which never required applying it for build/test.

<details>
<summary>Issues (0 blocking)</summary>

No blocking findings. Non-blocking observations carried forward:

1. **Unregistered-phone throttle (now fixed)** — the prior OTP-login review's defense-in-depth gap is resolved by this change; recorded here only to confirm closure, no action needed.
2. **Non-constant-time hash comparison** — `VerifyOTP` still compares code hashes with `!=` rather than `subtle.ConstantTimeCompare`. Pre-existing from the OTP-login commit, out of scope for this diff, low risk (keyed sha256). Not introduced here.

</details>

<details>
<summary>Details</summary>

### No new route; the existing two-endpoint flow absorbs registration

The deliberate design decision (documented in the plan) is to reuse `/auth/otp/request` and `/auth/otp/verify` rather than add `POST /auth/register/phone`. The mobile flow is unchanged from the client's view — request a code, verify it — and the account simply materializes on the first successful verify for a phone+role with no row. Because no route was added, `internal/app/expected_routes.go` and the OpenAPI path set need no change, and the route-parity test (`routes_openapi_test.go`) stays green per the recorded evidence. The only spec edit is fleshing out the two OTP operations with real `OTPRequestRequest`/`OTPVerifyRequest` request-body schemas (they were generic `API`-tagged stubs), which matches the Go DTOs field-for-field. That is request-body parity, not path parity, and is the correct scope for this change.

### Nullable Email/PasswordHash and the call-site sweep

The type change on `users.User` is mechanical but broad. The insert in `CreateInTx` passes `user.Email` and `user.PasswordHash` directly; as `*string` these write NULL when nil through `database/sql` with no extra plumbing. `scanUserRow` scans into `&user.Email` / `&user.PasswordHash`, now `**string`, which sets the field to nil on a NULL column. The readers were updated consistently:

- auth `register` takes the address of a normalized email string and of the bcrypt hash, so email/password accounts always store non-nil pointers.
- auth `login` and `ChangePassword` wrap the stored hash in `derefOr(hash, "")` before `security.CheckPassword`.
- `toUserResponse` (auth and admin) and `toProfileResponse` (users) dereference email with `derefOr(..., "")`.
- the workers email adapter returns `""` for a nil email.
- the employees admin repo and kyc repo switch `u.email` to `COALESCE(u.email, '')` because they scan into a non-nullable `string` — a phone-only employee's NULL email would otherwise fail the scan.

The claim that *every* reader was caught rests on the recorded `go build ./...` success, which is the right signal for a Go type change: an unconverted call site would not compile. I did not re-run it (per instructions) but spot-read the enumerated sites and they are each nil-safe.

### Password login against a phone-only account

```
login:  user = GetByEmailAndRole(...)       // phone-only user has nil hash
        CheckPassword(derefOr(hash,""), pw)  // bcrypt.CompareHashAndPassword("", pw)
                                             //   -> error -> ErrInvalidCredentials
```

`bcrypt.CompareHashAndPassword` on an empty hash returns a non-nil error (malformed hash), which `login` maps to `ErrInvalidCredentials`. No nil pointer is dereferenced because `derefOr` collapses the nil to `""` first. `TestPasswordLoginRejectsPhoneOnlyAccount` seeds a nil-email, nil-hash active user under an email key and asserts exactly `ErrInvalidCredentials` from `LoginCustomer` — the required "phone-only accounts can never be password-logged-in, never a 500" property. Note phone-only users have no email key in production, so `GetByEmailAndRole` would normally `ErrNotFound` first; the test forces the found-but-nil-hash branch to prove the deeper guard holds regardless.

### Request-side rate limit now covers unregistered phones

The prior OTP-login review flagged that the cooldown/window checks only counted rows that existed, and rows existed only for registered actives, so unknown phones bypassed the throttle. This change reorders `RequestOTP`:

```
RequestOTP:
  count  = CountOTPRequestsSince(phone, role, now-window)   ─┐ run FIRST, before any
  latest = LatestOTPCreatedAt(phone, role)                  ─┘ user lookup
  if count  >= maxPerWindow        -> ErrOTPRateLimited
  if latest within cooldown        -> ErrOTPRateLimited
  user = GetByPhoneAndRole(...)    // ErrNotFound is fine -> proceed
  if user != nil && !active        -> return nil (neutral, no code)
  ...generate + CreateOTPCode (row now exists for this phone+role) + Send...
```

Because an unregistered phone now gets a stored code, its `(phone, role)` counters increment and the next rapid request trips the cooldown. `TestRequestOTPUnknownPhoneSendsCodeAndRateLimits` asserts exactly this: one code stored and sent for the unknown phone, neutral `nil` error, and `ErrOTPRateLimited` on an immediate second request. The response stays non-enumerable — the handler returns the same "If the phone is registered, an OTP was sent" 200 regardless. The disabled-account case is the one silent skip: an existing suspended/deactivated user returns neutral `nil` without storing or sending, so a disabled account cannot be OTP-logged-in. Note this means a disabled account is not rate-limited (no row written), but that path sends nothing, so there is no SMS-bomb exposure there.

### Auto-create is transaction-safe and idempotent

`createPhoneOnlyUser` builds a `users.User` with nil email/password and dispatches to `registrar.RegisterCustomer`/`RegisterEmployee` by role (unknown role → `ErrValidation`, mapped to 400). The registrar wraps the user insert and the role profile insert in a single `database.RunInTx`, so a phone-only account never ends up with a user row but no profile row. Concurrency safety rests on `users_phone_role_unique`: `CreateInTx` maps a `23505` on that constraint to `users.ErrDuplicatePhone`, which `createPhoneOnlyUser` catches and resolves by reloading via `GetByPhoneAndRole`. So two simultaneous verifies for the same new phone produce one account; the loser issues tokens for the winner's row rather than failing. `TestVerifyOTPAutoCreatesPhoneOnlyUser` confirms the created row has nil email, nil password hash, active status, a profile via the mock registrar, tokens issued, and the code consumed; `TestVerifyOTPExistingUserUnchanged` confirms an existing user logs in with no duplicate created.

Ordering is correct: the code hash is validated and the expiry/attempt-cap checks pass *before* any user is created, so a wrong or expired code never provisions an account. `MarkOTPConsumed` runs after the user is resolved and before token issuance. One edge worth naming: if user creation succeeds but `MarkOTPConsumed` or `issueTokens` then errors, the account persists while the code stays unconsumed — a retry with the same still-valid code would find the now-existing user and log in, which is benign (no duplicate, the account was going to be created anyway).

### OTP primitives unchanged

Hashing (`hashLifecycleToken`, keyed sha256), single-use (`consumed_at IS NULL` filter + `MarkOTPConsumed` guarded update), expiry (`now.After(ExpiresAt)`), and the attempt cap (`AttemptCount >= maxAttempts`, incremented only on hash mismatch) are byte-for-byte the base OTP-login logic; this diff does not touch them. No OTP code is written to logs — the handler and service never log the plaintext, and the generated code leaves the process only via the (Noop) sender.

### Migration 000029

Up drops `NOT NULL` on `users.email` and `users.password_hash`, with a comment explaining that `users_email_role_unique` and `users_phone_role_unique` are kept and that Postgres treats NULLs as distinct in a unique index, so multiple phone-only rows (NULL email) coexist — the intended behavior, and the reason no constraint is added to reject NULL emails. Down re-adds `SET NOT NULL` on both, carrying the required warning that phone-only rows must be backfilled first or the down fails on existing NULLs. The pair is a valid inverse for a dev/local rollback. Not applied to a live DB (Docker/`migrate` CLI unavailable in the sandbox); the SQL is plain `ALTER COLUMN` and was read by hand, matching the verification note.

### Test coverage

New/updated auth tests recorded as passing: `TestVerifyOTPAutoCreatesPhoneOnlyUser`, `TestVerifyOTPExistingUserUnchanged`, `TestRequestOTPUnknownPhoneSendsCodeAndRateLimits` (replaces the old "unknown phone sends nothing" test), `TestVerifyOTPInactiveUser` (existing-disabled still neutral/`ErrUserInactive`), and `TestPasswordLoginRejectsPhoneOnlyAccount`. The mock user store was updated to key phone-only (nil-email) users by phone and to return `ErrDuplicatePhone` on a phone collision, which is what lets the auto-create and concurrency assertions be meaningful.

Not tested: the repository SQL and the migration against a live Postgres (mocks stand in, consistent with project convention), and the HTTP handler status mapping for the OTP endpoints (read-verified only). The concurrent-verify race is covered by the duplicate-reload unit path but not by an actual parallel-goroutine test — acceptable given the uniqueness constraint carries the real guarantee.

</details>

<details>
<summary>File map</summary>

- `internal/modules/users/model.go` — `Email`/`PasswordHash` → `*string`.
- `internal/modules/users/repository.go` — insert passes pointers directly (NULL on nil); duplicate-phone mapping already present.
- `internal/modules/users/scan.go` — scans into `**string` (NULL → nil), no change needed beyond the type.
- `internal/modules/auth/otp_service.go` — `RequestOTP` sends for unregistered phones (throttle before lookup); `VerifyOTP` auto-creates on no-user; new `createPhoneOnlyUser` with duplicate-reload.
- `internal/modules/auth/service.go` — `register` stores `*string` email/hash; `login` nil-safe hash; `derefOr` helper; `toUserResponse` deref.
- `internal/modules/auth/lifecycle_service.go` — nil-safe email/hash derefs.
- `internal/modules/auth/dto.go` — `UserResponse.Email` tagged `omitempty`.
- `internal/modules/admin/service.go`, `internal/modules/users/service.go`, `internal/modules/workers/user_email_adapter.go` — nil-safe email mapping.
- `internal/modules/employees/admin_repository.go`, `internal/modules/kyc/repository.go` — `u.email` → `COALESCE(u.email,'')`.
- `internal/app/spec/openapi.yaml` — OTP request-body schemas (`OTPRequestRequest`/`OTPVerifyRequest`); no path changes.
- `migrations/000029_phone_only_users.{up,down}.sql` — drop/restore NOT NULL on email + password_hash.
- `*_test.go` (auth, users, admin) — `*string` model, new phone-only cases.
- `docs/IMPLEMENTED_FEATURES.md`, `docs/PHASE_PLAN.md` — doc updates.

Full diff: `git diff 270e1bd..HEAD` on branch `azure` (phone-only commits `14d7c7c`, `01bcd09`).

</details>
