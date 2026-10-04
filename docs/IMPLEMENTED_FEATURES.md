# Implemented Features — Full Backend Reference

Complete inventory of what the Go backend implements, by module. Source of truth:
`internal/modules/*` and the mounted routes in `internal/app/server.go`
(see also `internal/app/expected_routes.go` and `internal/app/spec/openapi.yaml`).

All routes are under the `/api/v1` prefix. Roles: **customer**, **employee**
(provider), **admin**. `public` routes are unauthenticated.

---

## Platform

| Module | What it does |
|---|---|
| `config` | Env-based configuration (`internal/config`). |
| `app` | HTTP server, router, middleware chain, OpenAPI docs (`/api/v1/docs/`). |
| `events` | In-process event dispatcher (publishes domain events). |
| `workers` | Async notification worker + booking event publisher. |
| `websocket` | WebSocket hub + handler for realtime chat/notifications. |
| `storage` | S3-backed file storage (presigned uploads), enabled when S3 is configured. |
| `location` | Location primitives used by search/tracking. |
| `dashboard` | Admin dashboard aggregation. |

Middleware: RequestID, RealIP, Recoverer, CORS, request logging, JWT authentication,
role enforcement, admin audit logging.

---

## Auth (`auth`)
Two parallel account paths, both issuing the same JWT access/refresh token pair:

- **Phone-only (OTP) registration + login** (customer and employee): enter phone →
  `POST /auth/otp/request` → `POST /auth/otp/verify`. If no account exists for that
  phone + role, it is **auto-created on the first successful verify** (email and
  password left empty) and tokens are issued; existing accounts just log in. No
  separate registration call is needed for the mobile flow.
- **Email + password** registration and login remain fully supported in parallel
  (used by admin and existing users). A phone-only account has no password, so
  password login against it fails with invalid-credentials (never a 500).

- `POST /auth/register/customer`
- `POST /auth/register/employee`
- `POST /auth/login/customer`
- `POST /auth/login/employee`
- `POST /auth/login/admin`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `POST /auth/verify-email`
- `POST /auth/otp/request` — request a one-time login code for a phone + role (neutral response; rate limited per phone + role **including unregistered phones**, so a code is sent even for a new phone to enable auto-create on verify)
- `POST /auth/otp/verify` — verify the code and receive the standard `AuthResponse` (JWT access/refresh + user); **auto-creates a phone-only account** if none exists for that phone + role
- `POST /auth/resend-verification` (auth)
- `POST /auth/change-password` (auth)

### OTP delivery (pluggable via `OTP_PROVIDER`)

The backend always generates, hashes, stores, expires, and verifies the OTP
itself; the provider is only a delivery transport. `OTP_PROVIDER` selects it:

- empty (default) — dev/log sender in non-production (logs the code, never sends
  it), no-op in production.
- `twilio` — SMS via the Twilio Messages REST API (`net/http`, no SDK). Requires
  `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, and either `TWILIO_FROM_NUMBER` or
  `TWILIO_MESSAGING_SERVICE_SID`. `OTP_SMS_TEMPLATE` substitutes the code into
  `{code}`. Phones are normalized to E.164 with a leading `+`.
- `email` — SMTP fallback reusing the configured `SMTP_*` sender. DEV/TESTING
  only: it emails the code to the address supplied in the phone field, so
  genuine phone-only users (who have no email) cannot be reached this way.

Caveats: a Twilio **trial** account can only send to verified numbers, and
production SMS to Indian (`+91`) numbers additionally requires India DLT
registration (regulatory, outside this codebase).

> Not implemented: Google/OAuth social login.

---

## Public (`public`)
Unauthenticated discovery.

- `GET /public/home`
- `GET /public/categories`
- `GET /public/providers`
- `GET /public/providers/{id}`
- `GET /public/services`
- `GET /public/services/{id}`

---

## Users (`users`)
Account + address book.

- `GET /users/me`
- `PUT /users/me`
- `PATCH /users/me/deactivate`
- `GET /users/addresses`
- `POST /users/addresses`
- `PUT /users/addresses/{id}`
- `DELETE /users/addresses/{id}`

---

## Search (`search`)
- `GET /search/services`
- `GET /search/employees` (location-aware)

---

## Employees / Providers (`employees`)
Provider profile; integrates badges, KYC status, profile views, profile files.

- `GET /employees/{id}/public-profile`
- `GET /employee/profile`
- `PUT /employee/profile`

---

## KYC (`kyc`)
Provider identity verification; syncs verification status back to employee record.

- `POST /employee/kyc`
- `GET /employee/kyc`
- `GET /admin/kyc`, `GET /admin/kyc/{id}`
- `PATCH /admin/kyc/{id}/approve`, `/reject`

---

## Categories (`categories`)
- `GET /categories`
- Admin: `POST /admin/categories`, `PUT /admin/categories/{id}`, `DELETE /admin/categories/{id}`

---

## Services (`services`)
Provider service listings.

- `GET /services`, `GET /services/{id}`
- `GET /employees/{id}/services`
- `POST /employee/services`, `GET /employee/services`
- `PUT /employee/services/{id}`, `DELETE /employee/services/{id}`
- `PATCH /employee/services/{id}/status`
- Admin: `GET /admin/services`, `PATCH /admin/services/{id}/activate`, `/deactivate`

---

## Availability (`availability`)
- `GET /employees/{id}/availability`
- `GET/POST /employee/availability`
- `PUT/DELETE /employee/availability/{id}`

---

## Bookings (`bookings`)
Full lifecycle. Booking amount = the service's fixed price (no quote negotiation).

Customer:
- `POST /bookings`, `POST /bookings/rebook`
- `GET /bookings`, `GET /bookings/{id}`
- `GET /bookings/{id}/rebook-preview`
- `PATCH /bookings/{id}/cancel`, `/reschedule`

Provider:
- `GET /employee/bookings`
- `PATCH /employee/bookings/{id}/accept`, `/reject`, `/start`, `/complete`, `/cancel`, `/reschedule`, `/no-show`

Admin:
- `GET /admin/bookings`, `GET /admin/bookings/{id}`, `PATCH /admin/bookings/{id}/status`

---

## Reviews & Ratings (`reviews`, `ratings`, `badges`)
- `GET /employees/{id}/reviews`
- `POST /bookings/{id}/review`
- `GET /employee/reviews`
- `POST /employee/reviews/{id}/reply`
- Admin moderation: `GET /admin/reviews`, `PATCH /admin/reviews/{id}/approve`, `/hide`
- Ratings are refreshed from reviews; badges are awarded via review events.

---

## Payments & Subscriptions (`payments`, `subscriptions`, `webhooks`)
Razorpay gateway. Payment flow is currently oriented to **provider subscriptions**
and admin refunds — there is no customer per-booking checkout endpoint.

Payments:
- `GET /employee/payments`
- Admin: `GET /admin/payments`, `POST /admin/payments/{id}/refund`

Subscriptions (Phase 2 provider feature):
- `GET /subscription-plans`
- `POST /employee/subscriptions/create-order`, `/verify-payment`, `/cancel`, `/change-plan`
- `PATCH /employee/subscriptions/auto-renew`
- `GET /employee/subscriptions/current`
- Admin: `POST /admin/subscription-plans`, `PUT /admin/subscription-plans/{id}`, `GET /admin/subscriptions`

Webhooks:
- `POST /webhooks/razorpay`

---

## Notifications & Chat (`notifications`, `chat`, `websocket`)
Notifications:
- `GET /notifications`
- `PATCH /notifications/{id}/read`, `PATCH /notifications/read-all`
- `POST /device-tokens` (FCM push, when configured)

Chat:
- `GET /chat/conversations`
- `GET /chat/conversations/{id}/messages`
- `POST /chat/conversations/{id}/messages`
- `PATCH /chat/conversations/{id}/messages/{messageId}/read`

Realtime:
- `GET /ws`

Delivery channels: in-app, push (FCM), email (SMTP), WebSocket.

---

## Support (`support`)
- `POST /support/tickets`, `GET /support/tickets`
- Admin: `GET /admin/support/tickets`, `GET /admin/support/tickets/{id}`, `PATCH /admin/support/tickets/{id}`, `POST /admin/support/tickets/{id}/messages`

---

## Reports & Moderation (`reports`, `moderation`)
- `POST /reports`
- Admin: `GET /admin/reports`, `PATCH /admin/reports/{id}/resolve`, `GET /admin/reports/export`
- Moderation service ties review + report moderation together.

---

## Analytics (`analytics`)
Provider:
- `GET /employee/analytics/summary`, `/bookings`, `/reviews`

Admin:
- `GET /admin/analytics/summary`, `/revenue`, `/bookings`, `/categories`

---

## Storage / Uploads (`storage`)
- `POST /uploads/presign`
- `DELETE /uploads/{id}`

Mounted only when S3 is configured.

---

## Admin (`admin`, `settings`, `dashboard`)
Separate web app surface.

- `GET /admin/dashboard/summary`
- `GET /admin/users`, `GET /admin/users/{id}`, `PATCH /admin/users/{id}/suspend`, `/activate`
- `GET /admin/employees`, `GET /admin/employees/{id}`, `PATCH .../approve`, `/reject`, `/suspend`
- `GET /admin/settings`, `PUT /admin/settings`
- Plus admin routes listed under KYC, categories, services, bookings, payments,
  subscriptions, reviews, reports, support, analytics above.
- All admin actions are audit-logged.

---

## Modules with no HTTP routes (internal/support)
`events`, `workers`, `websocket` (hub), `location`, `storage` (infra), `moderation`
(service layer), `dashboard` (aggregation) — these back other modules rather than
exposing their own standalone endpoints beyond what's listed.

---

## Known gaps (not implemented)
- Customer per-booking payment checkout (card/UPI/cash for a booking).
- Google/OAuth social login. (Mobile OTP login is implemented; see Auth section.)
- Quote negotiation (provider proposes a price per request; customer accepts).
- Dedicated live-location tracking broadcast endpoint (primitives exist in `location` + `websocket`).
