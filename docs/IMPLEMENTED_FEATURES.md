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
Email + password with JWT access/refresh tokens, plus phone + OTP login
(customer and employee) that issues the same JWT token pair.

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
- `POST /auth/otp/request` — request a one-time login code for a phone + role (neutral response; rate limited)
- `POST /auth/otp/verify` — verify the code and receive the standard `AuthResponse` (JWT access/refresh + user)
- `POST /auth/resend-verification` (auth)
- `POST /auth/change-password` (auth)

> Not implemented: Google/OAuth social login. (Shipped OTP delivery uses a
> no-op/log sender in dev; a real SMS provider plugs in behind the `OTPSender` interface.)

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
