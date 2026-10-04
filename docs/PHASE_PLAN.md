# Backend Phase Plan

Backend-only delivery plan. This describes what the **Go backend** has done and what
it must build per phase. It is not a design/UI mapping. Admin is a separate web app;
its endpoints already exist and are listed in `IMPLEMENTED_FEATURES.md`.

Legend:
- ✅ **Done** — implemented and wired in `internal/app/server.go`.
- 🟡 **Partial** — code exists but needs changes/completion.
- ❌ **To build** — not implemented.

---

## Phase 1 — Core marketplace backend

Goal: back the full customer→booking→service→payment→review loop and the provider
onboarding/job loop with production-ready APIs.

### Already done (✅)
These capabilities are complete and need no Phase 1 work:

- **Auth (email/password + JWT):** register/login (customer, employee, admin),
  refresh, logout, `me`, email verify/resend, forgot/reset/change password.
- **Auth (phone OTP registration + login):** `POST /auth/otp/request` +
  `POST /auth/otp/verify` for customer and employee — hashed single-use codes with
  expiry, attempt caps, and per-phone rate limiting (incl. unregistered phones);
  verify auto-creates a phone-only account on first success and issues the same JWT
  pair. Email/password remains a parallel path. (Google/OAuth still open.)
- **Discovery:** `public/home`, categories, providers, services; `search/services`,
  `search/employees` (location-aware).
- **Provider profile:** `employee/profile` get/update, public profile, badges,
  profile views, profile files.
- **KYC:** provider submit + status sync (admin review lives in the admin app).
- **Services:** provider CRUD + status; public listing.
- **Availability:** provider CRUD + public read.
- **Bookings:** full lifecycle — create, cancel, reschedule, rebook; provider
  accept/reject/start/complete/cancel/reschedule/no-show.
- **Reviews/Ratings/Badges:** submit review, provider reply, rating refresh,
  badge awarding.
- **Chat + realtime:** conversations, messages, read receipts, `GET /ws`.
- **Notifications:** list, read/read-all, device tokens; in-app/push/email/WS delivery.
- **Support:** ticket create/list.
- **Uploads:** S3 presign + delete.

### To build / finish in Phase 1

1. **Customer booking payment (❌ — highest priority)**
   The payments module currently only lists employee payments and does admin refunds;
   the Razorpay gateway is wired for subscriptions. Needs:
   - Create payment order for a **booking** (not a subscription).
   - Verify payment and transition the booking to a paid state (idempotent).
   - Record the payment against the booking; emit a payment event.
   - Support the intended methods (online via Razorpay; cash settled at completion).
   - Reuse the existing `payments.RazorpayGateway` and `POST /webhooks/razorpay`.

2. **Auth model decision (🚧 OTP registration + login shipped; Google open)**
   Backend is email/password + JWT, now plus phone-only OTP registration and login.
   - OTP: ✅ shipped — phone send-code + verify-code endpoints, OTP store with
     expiry/rate limit; verify **auto-creates a phone-only account** (no email/
     password) when the phone+role is new, then issues the same JWT pair. Email/
     password stays a parallel path. OTP SMS delivery via Twilio is wired behind
     the `OTPSender` interface, selected by `OTP_PROVIDER` (pending a Twilio
     account/number + India DLT for `+91` production); an SMTP email fallback is
     available, and dev still uses the no-op/log sender.
   - Google: ❌ open — OAuth token verification endpoint, link/create user, issue JWT pair.

3. **Quote flow decision (❌ if required)**
   Today a booking uses the service's fixed `total_amount`; there is no quote entity.
   If provider-proposed quotes are required:
   - Quote entity (request → provider proposes amount → customer accepts/rejects).
   - Tie an accepted quote to booking creation / `total_amount`.
   If not required, no work — booking at listed price stands.

4. **Live location tracking (🟡)**
   `location` + `websocket` modules provide primitives. Confirm and, if missing,
   add a provider live-location broadcast + a customer subscribe path over WS for the
   "on the way" tracking use case.

### Phase 1 exit criteria
- A customer can pay for a booking end-to-end with idempotent verification.
- Auth model is settled (either confirmed email/password, or OTP/Google shipped).
- Quote decision is settled (built or explicitly dropped).
- Live tracking path confirmed working or explicitly deferred.
- `go test ./...` green; new endpoints in `expected_routes.go` + `openapi.yaml`.

---

## Phase 2 — Provider subscriptions

Backend is **already built** ahead of need. No new backend work required to start;
revisit only for refinements when Phase 2 activates.

- Plans listing, create-order, verify-payment, cancel, change-plan, auto-renew, current.
- Admin plan management + subscriptions listing.
- Razorpay gateway + `POST /webhooks/razorpay`.

Possible Phase 2 refinements (not yet needed): plan proration, grace periods,
dunning/retry on failed renewals, subscription-gated feature checks.

---

## Phase 3 / later — already built, activate when needed

Implemented and tested; no additional backend work to use them. Pull into a phase
when the product needs them.

- **Reports + moderation:** `POST /reports`, admin resolve/export, review hide/approve.
- **Rebook:** `POST /bookings/rebook`, `GET /bookings/{id}/rebook-preview`.
- **Reschedule / no-show:** booking lifecycle endpoints (both sides).
- **Deep analytics:** `employee/analytics/bookings`, `/reviews`; admin analytics.
- **Review reply:** `POST /employee/reviews/{id}/reply`.
- **Account deactivation:** `PATCH /users/me/deactivate`.
- **Badges:** awarding wired via review events.

---

## Cross-cutting backend standards (every phase)
- Run `gofmt` and `go test ./...` before done.
- Keep handlers thin, services business-focused, repositories DB-focused.
- Keep new routes registered and documented: `internal/app/expected_routes.go` +
  `internal/app/spec/openapi.yaml` (the OpenAPI route test enforces parity).
- Payment-touching code must be idempotent, verify server-side, and never trust
  client-reported payment status.
