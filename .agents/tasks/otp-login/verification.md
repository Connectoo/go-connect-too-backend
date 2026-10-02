# Verification — OTP (phone) login

Iteration: FIRST (no `review.json` present at start). Implemented from `plan.md`.

All commands run from the repo root:
`/Users/mustafa/Desktop/projects/go-connect-too-backend`.

## 1. `gofmt -l .`

```
$ gofmt -w .
$ gofmt -l .
(no output)
```

Result: **PASS** — no files need formatting.

## 2. `go build ./...`

```
$ go build ./...
(no output, exit 0)
```

Result: **PASS** — the whole module compiles.

## 3. `go test ./...`

```
$ go test ./...
ok  internal/app                         (route + OpenAPI parity)
ok  internal/config
ok  internal/modules/auth                (existing + new OTP tests)
ok  internal/modules/users
... all other packages ok / [no test files]
exit 0
```

Result: **PASS** — every package with tests passes. No tests were skipped;
the auth OTP service tests are pure in-memory (no live DB needed). No live-DB
integration tests exist in these packages, so nothing was skipped for DB reasons.

Targeted runs:

```
$ go test ./internal/modules/auth/ -run OTP -v
--- PASS: TestVerifyOTPSuccess
--- PASS: TestVerifyOTPExpired
--- PASS: TestVerifyOTPWrongCode
--- PASS: TestVerifyOTPTooManyAttempts
--- PASS: TestVerifyOTPReuseConsumed
--- PASS: TestRequestOTPRateLimited
--- PASS: TestRequestOTPMaxPerWindow
--- PASS: TestRequestOTPUnknownPhoneNeutral
--- PASS: TestVerifyOTPInactiveUser
PASS

$ go test ./internal/app/ -run 'Routes|OpenAPI' -v
--- PASS: TestMountedRoutesMatchExpected
--- PASS: TestOpenAPICoversExpectedRoutes
--- PASS: TestOpenAPISpec
PASS
```

## 4. Route inventory + OpenAPI parity

- `internal/app/expected_routes.go` now lists `POST /auth/otp/request` and
  `POST /auth/otp/verify`.
- `internal/app/spec/openapi.yaml` has matching `/auth/otp/request` and
  `/auth/otp/verify` path items (copied from the `/auth/forgot-password` block shape).
- `TestMountedRoutesMatchExpected` and `TestOpenAPICoversExpectedRoutes` both pass,
  confirming the mounted routes, the inventory, and the spec are in sync.

## Migrations (reviewed by reading; no live DB in sandbox)

- `migrations/000028_otp_login.up.sql` creates table `otp_login_codes`
  (id, phone, role, code_hash CHAR(64), expires_at, attempt_count, consumed_at,
  created_at) plus two indexes. The code is stored HASHED (sha256 + secret), never
  plaintext.
- `migrations/000028_otp_login.down.sql` is `DROP TABLE IF EXISTS otp_login_codes;`,
  which cleanly drops exactly what the up migration creates (indexes drop with the table).
