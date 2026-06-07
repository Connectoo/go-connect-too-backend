# Go Connect — Full Platform Inventory

Master reference for building all web UIs. Base URL: `http://localhost:8080/api/v1`  
Swagger (dev): `http://localhost:8080/api/v1/docs/`

---

## 1. Web apps overview

| App | Folder | Port | Auth endpoint | Role | Status |
|-----|--------|------|---------------|------|--------|
| Public website | `web/website` | 3000 | None (public) | — | **Partial** |
| Customer portal | `web/customer` | 3002 | `POST /auth/login/customer` | `customer` | **Not started** |
| Employee portal | `web/employee` | 3003 | `POST /auth/login/employee` | `employee` | **Not started** |
| Admin dashboard | `web/admin` | 3001 | `POST /auth/login/admin` | `admin` | **Partial** |

**Env for all apps:** `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1`

---

## 2. Domain status enums (for UI badges/filters)

### Booking statuses
`pending` → `accepted` → `in_progress` → `completed`  
Also: `rejected`, `cancelled`, `no_show`

| Status | Customer can cancel? | Customer can reschedule? | Employee actions |
|--------|---------------------|--------------------------|------------------|
| pending | Yes | Yes | accept, reject, cancel |
| accepted | Yes | Yes | start, cancel, no-show, reschedule |
| in_progress | No | No | complete, no-show |
| completed | No | No | — (customer can review) |
| rejected | No | No | — |
| cancelled | No | No | — |
| no_show | No | No | — |

### Employee verification
`pending` | `approved` | `rejected`

### KYC status
`pending` | `approved` | `rejected`

### User account status
`active` | `inactive` | `suspended`

### User roles
`customer` | `employee` | `admin`

### Subscription status
`pending` | `active` | `expired` | `cancelled`

### Payment status
`pending` | `success` | `failed`

### Support ticket status
`open` | `in_progress` | `resolved` | `closed`

### Support ticket priority
`low` | `normal` | `high` | `urgent`

### User report status
`open` | `resolved`

---

## 3. Complete API list (137 endpoints)

Legend: **Auth** = Bearer JWT required | **Role** = required role

---

### 3.1 Health
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/health` | No | — | DB health check |

---

### 3.2 Auth (14 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/auth/register/customer` | No | — | Customer signup |
| POST | `/auth/register/employee` | No | — | Provider signup |
| POST | `/auth/login/customer` | No | — | Customer login |
| POST | `/auth/login/employee` | No | — | Provider login |
| POST | `/auth/login/admin` | No | — | Admin login |
| POST | `/auth/refresh` | No | — | Refresh access token |
| POST | `/auth/logout` | No | — | Revoke refresh token |
| GET | `/auth/me` | Yes | Any | Current user from token |
| POST | `/auth/forgot-password` | No | — | Send reset email |
| POST | `/auth/reset-password` | No | — | Reset with token |
| POST | `/auth/verify-email` | No | — | Verify email token |
| POST | `/auth/resend-verification` | Yes | Any | Resend verify email |
| POST | `/auth/change-password` | Yes | Any | Change password |

---

### 3.3 Public discovery (6 endpoints)
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/public/home` | No | Homepage: featured categories, providers, services, stats |
| GET | `/public/categories` | No | Active categories |
| GET | `/public/providers` | No | Provider list (query: limit, category_id) |
| GET | `/public/providers/{id}` | No | Provider public profile |
| GET | `/public/services` | No | Service list (query: category_id, limit) |
| GET | `/public/services/{id}` | No | Service detail |

---

### 3.4 Search (2 endpoints)
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/search/services` | No | Search services (query: q, category_id, page, limit) |
| GET | `/search/employees` | No | Search providers (query: q, page, limit) |

---

### 3.5 Categories (4 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/categories` | No | — | Public category list |
| POST | `/admin/categories` | Yes | admin | Create category |
| PUT | `/admin/categories/{id}` | Yes | admin | Update category |
| DELETE | `/admin/categories/{id}` | Yes | admin | Delete category |

---

### 3.6 Users / customer profile (7 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/users/me` | Yes | customer | Get profile |
| PUT | `/users/me` | Yes | customer | Update name, phone |
| PATCH | `/users/me/deactivate` | Yes | customer | Deactivate account |
| GET | `/users/addresses` | Yes | customer | List saved addresses |
| POST | `/users/addresses` | Yes | customer | Add address |
| PUT | `/users/addresses/{id}` | Yes | customer | Update address |
| DELETE | `/users/addresses/{id}` | Yes | customer | Delete address |

---

### 3.7 Employees / providers (9 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/employees/{id}/public-profile` | No | — | Public provider profile |
| GET | `/employee/profile` | Yes | employee | Own profile |
| PUT | `/employee/profile` | Yes | employee | Update profile (bio, skills, location, photo) |
| GET | `/admin/employees` | Yes | admin | List providers (filter: verification_status, q, page) |
| GET | `/admin/employees/{id}` | Yes | admin | Provider detail |
| PATCH | `/admin/employees/{id}/approve` | Yes | admin | Approve profile |
| PATCH | `/admin/employees/{id}/reject` | Yes | admin | Reject profile |
| PATCH | `/admin/employees/{id}/suspend` | Yes | admin | Suspend provider |

---

### 3.8 KYC (6 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/employee/kyc` | Yes | employee | Submit ID + address proof URLs |
| GET | `/employee/kyc` | Yes | employee | View own KYC status |
| GET | `/admin/kyc` | Yes | admin | KYC review queue |
| GET | `/admin/kyc/{id}` | Yes | admin | KYC detail |
| PATCH | `/admin/kyc/{id}/approve` | Yes | admin | Approve KYC |
| PATCH | `/admin/kyc/{id}/reject` | Yes | admin | Reject KYC (reason) |

---

### 3.9 Services (11 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/services` | No | — | Browse active services |
| GET | `/services/{id}` | No | — | Service detail |
| GET | `/employees/{id}/services` | No | — | Services by provider |
| POST | `/employee/services` | Yes | employee | Create service |
| GET | `/employee/services` | Yes | employee | List own services |
| PUT | `/employee/services/{id}` | Yes | employee | Update service |
| DELETE | `/employee/services/{id}` | Yes | employee | Delete service |
| PATCH | `/employee/services/{id}/status` | Yes | employee | Activate/deactivate |
| GET | `/admin/services` | Yes | admin | List all services |
| PATCH | `/admin/services/{id}/activate` | Yes | admin | Force activate |
| PATCH | `/admin/services/{id}/deactivate` | Yes | admin | Force deactivate |

---

### 3.10 Availability (5 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/employees/{id}/availability` | No | — | Public slot list for booking |
| GET | `/employee/availability` | Yes | employee | Own schedule |
| POST | `/employee/availability` | Yes | employee | Add slot (day, start, end) |
| PUT | `/employee/availability/{id}` | Yes | employee | Update slot |
| DELETE | `/employee/availability/{id}` | Yes | employee | Delete slot |

---

### 3.11 Bookings (18 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/bookings` | Yes | customer | Create booking |
| POST | `/bookings/rebook` | Yes | customer | Rebook from past booking |
| GET | `/bookings` | Yes | customer | List own bookings |
| GET | `/bookings/{id}` | Yes | customer | Booking detail |
| GET | `/bookings/{id}/rebook-preview` | Yes | customer | Can rebook? + suggested slot |
| PATCH | `/bookings/{id}/cancel` | Yes | customer | Cancel (optional reason) |
| PATCH | `/bookings/{id}/reschedule` | Yes | customer | Reschedule date/time |
| GET | `/employee/bookings` | Yes | employee | Provider inbox |
| PATCH | `/employee/bookings/{id}/accept` | Yes | employee | Accept pending |
| PATCH | `/employee/bookings/{id}/reject` | Yes | employee | Reject pending |
| PATCH | `/employee/bookings/{id}/start` | Yes | employee | Mark in progress |
| PATCH | `/employee/bookings/{id}/complete` | Yes | employee | Mark completed |
| PATCH | `/employee/bookings/{id}/cancel` | Yes | employee | Cancel |
| PATCH | `/employee/bookings/{id}/no-show` | Yes | employee | Mark no-show |
| PATCH | `/employee/bookings/{id}/reschedule` | Yes | employee | Reschedule |
| GET | `/admin/bookings` | Yes | admin | List all (filter: status, page) |
| GET | `/admin/bookings/{id}` | Yes | admin | Booking detail |
| PATCH | `/admin/bookings/{id}/status` | Yes | admin | Force status change |

**Create booking body:**
```json
{
  "service_id": "uuid",
  "booking_date": "2026-06-15",
  "start_time": "10:00",
  "end_time": "11:00",
  "customer_notes": "optional"
}
```

---

### 3.12 Reviews (5 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/employees/{id}/reviews` | No | — | Public reviews for provider |
| POST | `/bookings/{id}/review` | Yes | customer | Leave review after completed booking |
| GET | `/employee/reviews` | Yes | employee | Own reviews |
| POST | `/employee/reviews/{id}/reply` | Yes | employee | Reply to review |

---

### 3.13 Subscriptions (9 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/subscription-plans` | No | — | Public plan list |
| POST | `/employee/subscriptions/create-order` | Yes | employee | Create Razorpay order |
| POST | `/employee/subscriptions/verify-payment` | Yes | employee | Verify Razorpay payment |
| POST | `/employee/subscriptions/cancel` | Yes | employee | Cancel subscription |
| POST | `/employee/subscriptions/change-plan` | Yes | employee | Change plan |
| PATCH | `/employee/subscriptions/auto-renew` | Yes | employee | Toggle auto-renew |
| GET | `/employee/subscriptions/current` | Yes | employee | Current subscription |
| POST | `/admin/subscription-plans` | Yes | admin | Create plan |
| PUT | `/admin/subscription-plans/{id}` | Yes | admin | Update plan |
| GET | `/admin/subscriptions` | Yes | admin | All subscriptions |

---

### 3.14 Payments (4 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/employee/payments` | Yes | employee | Own payment history |
| GET | `/admin/payments` | Yes | admin | All payments |
| POST | `/admin/payments/{id}/refund` | Yes | admin | Issue refund |
| POST | `/webhooks/razorpay` | No (signed) | — | Razorpay webhook (backend only) |

---

### 3.15 Notifications (4 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/notifications` | Yes | Any | In-app notification list |
| PATCH | `/notifications/{id}/read` | Yes | Any | Mark one read |
| PATCH | `/notifications/read-all` | Yes | Any | Mark all read |
| POST | `/device-tokens` | Yes | Any | Register FCM device token |

---

### 3.16 Chat (4 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/chat/conversations` | Yes | customer/employee | Conversation list |
| GET | `/chat/conversations/{id}/messages` | Yes | customer/employee | Message history |
| POST | `/chat/conversations/{id}/messages` | Yes | customer/employee | Send message |
| PATCH | `/chat/conversations/{id}/messages/{messageId}/read` | Yes | customer/employee | Read receipt |

---

### 3.17 WebSocket (1 endpoint)
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/ws` | Yes (JWT) | Real-time events (bookings, chat, notifications) |

---

### 3.18 Support tickets (6 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/support/tickets` | Yes | customer | Create ticket |
| GET | `/support/tickets` | Yes | customer | List own tickets |
| GET | `/admin/support/tickets` | Yes | admin | All tickets |
| GET | `/admin/support/tickets/{id}` | Yes | admin | Ticket detail + messages |
| PATCH | `/admin/support/tickets/{id}` | Yes | admin | Update status/priority |
| POST | `/admin/support/tickets/{id}/messages` | Yes | admin | Staff reply |

---

### 3.19 Reports & moderation (7 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/reports` | Yes | Any | Report user/content |
| GET | `/admin/reports` | Yes | admin | Report queue |
| PATCH | `/admin/reports/{id}/resolve` | Yes | admin | Resolve report |
| GET | `/admin/reports/export` | Yes | admin | CSV export |
| GET | `/admin/reviews` | Yes | admin | Review moderation queue |
| PATCH | `/admin/reviews/{id}/approve` | Yes | admin | Approve review |
| PATCH | `/admin/reviews/{id}/hide` | Yes | admin | Hide review |

---

### 3.20 Admin platform (8 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/admin/dashboard/summary` | Yes | admin | KPI dashboard |
| GET | `/admin/users` | Yes | admin | User list |
| GET | `/admin/users/{id}` | Yes | admin | User detail |
| PATCH | `/admin/users/{id}/suspend` | Yes | admin | Suspend user |
| PATCH | `/admin/users/{id}/activate` | Yes | admin | Activate user |
| GET | `/admin/settings` | Yes | admin | Platform settings |
| PUT | `/admin/settings` | Yes | admin | Update settings |

**Settings keys:** `general`, `providers` (JSON blobs)

---

### 3.21 Analytics (7 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| GET | `/employee/analytics/summary` | Yes | employee | Provider KPIs |
| GET | `/employee/analytics/bookings` | Yes | employee | Bookings by status/date |
| GET | `/employee/analytics/reviews` | Yes | employee | Review stats |
| GET | `/admin/analytics/summary` | Yes | admin | Platform KPIs |
| GET | `/admin/analytics/revenue` | Yes | admin | Revenue over time |
| GET | `/admin/analytics/bookings` | Yes | admin | Bookings over time |
| GET | `/admin/analytics/categories` | Yes | admin | Category breakdown |

---

### 3.22 File uploads (2 endpoints)
| Method | Path | Auth | Role | Purpose |
|--------|------|------|------|---------|
| POST | `/uploads/presign` | Yes | Any | Get S3/MinIO presigned upload URL |
| DELETE | `/uploads/{id}` | Yes | Any | Delete uploaded file |

**Used for:** profile photos, KYC documents, review media

---

## 4. External integrations

| Integration | Backend code | Config needed | Without config | Used by UI |
|-------------|-------------|---------------|----------------|------------|
| PostgreSQL | ✅ Built | `DATABASE_URL` | App won't start | All |
| JWT auth | ✅ Built | `JWT_*` secrets | Required | All apps |
| S3/MinIO storage | ✅ Built | `STORAGE_PROVIDER=s3`, S3 vars | Uploads disabled | Employee KYC, profile photo |
| Razorpay | ✅ Built | `RAZORPAY_*` keys | Subscription checkout fails | Employee subscription only |
| SMTP email | ✅ Built | `SMTP_*` vars | Emails skipped silently | Forgot password, verify email |
| FCM push | ✅ Built | `FCM_*` vars | Noop (in-app still works) | Mobile/web push later |
| WebSocket | ✅ Built | JWT on connect | Works if auth valid | Chat, live notifications |
| In-app notifications | ✅ Built | None | Always works | Customer + employee |
| Event worker | ✅ Built | None | Always works | Background (no UI) |

---

## 5. NOT built (do not build UI for these yet)

From `docs/DEFERRED_SCOPE.md`:

| Feature | Notes |
|---------|-------|
| Customer booking payment / checkout | Bookings have no Razorpay — amount is stored but not charged |
| Wallet, saved cards, COD | No API |
| Provider payouts / bank linking | No API |
| Invoice PDF/email | Module reserved, no routes |
| Elasticsearch / advanced search | Postgres ILIKE only |
| SMS notifications | No API |
| Phone OTP login | No API |
| Social login (Google/Facebook) | No API |
| 2FA / session management | No API |
| CMS (FAQ, blog) | No API |
| Marketing campaigns | No API |
| Feature flags / multi-admin RBAC | Single admin role only |
| GDPR export/delete | No API |
| Maps / geocoding | Location is text + lat/lng fields only |
| Favorites / saved providers | No API |
| Guest booking | Auth required for bookings |
| Recurring booking series | Rebook is one-off only |

---

## 6. Full UI screen list — all 4 apps

Status: ✅ Done | 🔶 Partial | ❌ Not started

---

### 6.1 Public website (`web/website` — port 3000)

| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| W1 | `/` | Homepage (hero, featured categories/providers/services) | `GET /public/home` | ✅ |
| W2 | `/categories` | Category grid | `GET /public/categories` | ✅ |
| W3 | `/providers` | Provider listing | `GET /public/providers` | ✅ |
| W4 | `/providers/[id]` | Provider profile + services + reviews | `/public/providers/{id}`, `/employees/{id}/services`, `/employees/{id}/reviews` | ✅ |
| W5 | `/services/[id]` | Service detail + "Book" CTA | `GET /public/services/{id}` | ✅ |
| W6 | `/search` | Search results (services + providers tabs) | `GET /search/services`, `/search/employees` | ❌ |
| W7 | `/about` | About page | Static | ✅ |
| W8 | `/contact` | Contact page | Static | ✅ |
| W9 | `/privacy` | Privacy policy | Static | ✅ |
| W10 | `/terms` | Terms of service | Static | ✅ |
| W11 | `/login` | Customer login | `POST /auth/login/customer` | 🔶 Move to customer app |
| W12 | `/register` | Customer register | `POST /auth/register/customer` | 🔶 Move to customer app |

**Website links to build:**
- "Book now" on W5 → `http://localhost:3002/book/[serviceId]`
- "Sign in" in header → `http://localhost:3002/login`
- Remove mock fallbacks in `lib/mocks.ts` once API is stable

---

### 6.2 Customer portal (`web/customer` — port 3002) — NEW APP

#### Auth & account
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C1 | `/login` | Sign in | `POST /auth/login/customer` | ❌ |
| C2 | `/register` | Sign up | `POST /auth/register/customer` | ❌ |
| C3 | `/forgot-password` | Request reset email | `POST /auth/forgot-password` | ❌ |
| C4 | `/reset-password` | Reset with token (from email link) | `POST /auth/reset-password` | ❌ |
| C5 | `/verify-email` | Email verification | `POST /auth/verify-email` | ❌ |

#### Dashboard & profile
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C6 | `/dashboard` | Overview (upcoming bookings, quick actions) | `GET /bookings?status=pending,accepted` | ❌ |
| C7 | `/profile` | Edit name, phone | `GET/PUT /users/me` | ❌ |
| C8 | `/settings/security` | Change password | `POST /auth/change-password` | ❌ |
| C9 | `/settings/account` | Deactivate account | `PATCH /users/me/deactivate` | ❌ |
| C10 | `/settings/notifications` | Resend verification | `POST /auth/resend-verification` | ❌ |

#### Addresses
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C11 | `/addresses` | Saved addresses list | `GET /users/addresses` | ❌ |
| C12 | `/addresses/new` | Add address form | `POST /users/addresses` | ❌ |
| C13 | `/addresses/[id]/edit` | Edit address | `PUT /users/addresses/{id}` | ❌ |

#### Bookings
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C14 | `/book/[serviceId]` | Booking wizard (date → time slot → notes → confirm) | `GET /services/{id}`, `GET /employees/{id}/availability`, `POST /bookings` | ✅ |
| C15 | `/bookings` | My bookings list (tabs: upcoming / past / cancelled) | `GET /bookings` | ✅ |
| C16 | `/bookings/[id]` | Booking detail + status timeline | `GET /bookings/{id}` | ✅ |
| C17 | `/bookings/[id]/cancel` | Cancel confirmation dialog | `PATCH /bookings/{id}/cancel` | ❌ |
| C18 | `/bookings/[id]/reschedule` | Reschedule form | `PATCH /bookings/{id}/reschedule` | ❌ |
| C19 | `/bookings/[id]/rebook` | Rebook from completed booking | `GET /bookings/{id}/rebook-preview`, `POST /bookings/rebook` | ❌ |
| C20 | `/bookings/[id]/review` | Leave review (rating + text + optional media via upload) | `POST /bookings/{id}/review`, `POST /uploads/presign` | ❌ |

#### Communication
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C21 | `/notifications` | Notification inbox | `GET /notifications`, mark read | ❌ |
| C22 | `/chat` | Conversations list | `GET /chat/conversations` | ❌ |
| C23 | `/chat/[id]` | Message thread | messages CRUD + WebSocket `/ws` | ❌ |

#### Support & safety
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| C24 | `/support` | My support tickets | `GET /support/tickets` | ❌ |
| C25 | `/support/new` | Create ticket | `POST /support/tickets` | ❌ |
| C26 | `/support/[id]` | Ticket detail + messages | `GET /support/tickets` (detail if added) | ❌ |
| C27 | `/report` | Report user/content modal | `POST /reports` | ❌ |

**Customer sidebar nav:**
```
Dashboard | Bookings | Messages | Notifications | Profile | Addresses | Support
```

**Customer app scaffold files to create:**
```
web/customer/
├── app/login/page.tsx
├── app/register/page.tsx
├── app/(portal)/layout.tsx          ← sidebar + auth guard
├── app/(portal)/dashboard/page.tsx
├── app/(portal)/bookings/...
├── lib/api-client.ts                ← copy from admin
├── lib/auth.ts                      ← customer_access_token cookie
├── middleware.ts
├── services/customer.ts
├── hooks/use-customer.ts
└── .env.example
```

---

### 6.3 Employee portal (`web/employee` — port 3003) — NEW APP

#### Auth & onboarding
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E1 | `/login` | Sign in | `POST /auth/login/employee` | ❌ |
| E2 | `/register` | Sign up | `POST /auth/register/employee` | ❌ |
| E3 | `/forgot-password` | Request reset | `POST /auth/forgot-password` | ❌ |
| E4 | `/reset-password` | Reset password | `POST /auth/reset-password` | ❌ |
| E5 | `/onboarding` | Step wizard: profile → KYC → services → availability | Multiple | ❌ |

#### Profile & KYC
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E6 | `/profile` | Business profile (bio, skills, languages, location, photo) | `GET/PUT /employee/profile`, `POST /uploads/presign` | ✅ (photo upload deferred) |
| E7 | `/kyc` | KYC submission (ID proof + address proof upload) | `POST/GET /employee/kyc`, `POST /uploads/presign` | ❌ |
| E8 | `/settings/security` | Change password | `POST /auth/change-password` | ❌ |

#### Services
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E9 | `/services` | My services list | `GET /employee/services` | ✅ |
| E10 | `/services/new` | Create service | `POST /employee/services` | ✅ |
| E11 | `/services/[id]/edit` | Edit service | `PUT /employee/services/{id}` | ✅ |
| E12 | `/services/[id]` | Service detail + activate/deactivate toggle | `PATCH /employee/services/{id}/status` | ✅ |

#### Availability
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E13 | `/availability` | Weekly schedule manager (calendar/grid) | `GET/POST/PUT/DELETE /employee/availability` | ✅ |

#### Bookings
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E14 | `/dashboard` | KPI summary | `GET /employee/analytics/summary` | ✅ |
| E15 | `/bookings` | Booking inbox (tabs: pending / active / completed) | `GET /employee/bookings` | ✅ |
| E16 | `/bookings/[id]` | Booking detail + action buttons | All PATCH employee booking endpoints | ✅ |

**Employee booking actions on E16:**
- pending → Accept / Reject
- accepted → Start / Cancel / Reschedule / No-show
- in_progress → Complete / No-show

#### Reviews
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E17 | `/reviews` | Reviews list | `GET /employee/reviews` | ❌ |
| E18 | `/reviews/[id]` | Review detail + reply form | `POST /employee/reviews/{id}/reply` | ❌ |

#### Subscription & payments
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E19 | `/subscription` | Current plan + status | `GET /employee/subscriptions/current`, `GET /subscription-plans` | ❌ |
| E20 | `/subscription/plans` | Plan comparison + choose plan | `GET /subscription-plans` | ❌ |
| E21 | `/subscription/checkout` | Razorpay checkout | `POST /create-order`, `POST /verify-payment` | ❌ |
| E22 | `/subscription/manage` | Cancel, change plan, auto-renew toggle | cancel, change-plan, auto-renew | ❌ |
| E23 | `/payments` | Payment history | `GET /employee/payments` | ❌ |

#### Analytics
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E24 | `/analytics` | Charts dashboard | `/employee/analytics/bookings`, `/reviews` | ❌ |

#### Communication
| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| E25 | `/notifications` | Notification inbox | `GET /notifications` | ❌ |
| E26 | `/chat` | Conversations list | `GET /chat/conversations` | ❌ |
| E27 | `/chat/[id]` | Message thread | messages + WebSocket | ❌ |

**Employee sidebar nav:**
```
Dashboard | Bookings | Services | Availability | Reviews | Subscription | Analytics | Profile | KYC
```

**Onboarding banner:** show on all pages until `verification_status = approved` AND KYC `approved`

---

### 6.4 Admin dashboard (`web/admin` — port 3001)

| # | Route | Screen | APIs | Status |
|---|-------|--------|------|--------|
| A1 | `/login` | Admin sign in | `POST /auth/login/admin` | ✅ |
| A2 | `/dashboard` | KPI cards (8 metrics) | `GET /admin/dashboard/summary` | ✅ |
| A3 | `/employees` | Provider approval queue | list, approve, reject | ✅ |
| A4 | `/employees/[id]` | Provider detail + suspend | `GET /admin/employees/{id}`, suspend | ❌ |
| A5 | `/categories` | Category CRUD | full CRUD | ✅ |
| A6 | `/bookings` | Bookings list (read-only) | `GET /admin/bookings` | ✅ |
| A7 | `/bookings/[id]` | Booking detail + status override | `GET /admin/bookings/{id}`, `PATCH .../status` | ✅ |
| A8 | `/kyc` | KYC review queue | `GET /admin/kyc`, approve/reject | ❌ |
| A9 | `/users` | User management | list, suspend, activate | ❌ |
| A10 | `/users/[id]` | User detail | `GET /admin/users/{id}` | ❌ |
| A11 | `/services` | Service moderation | list, activate/deactivate | ❌ |
| A12 | `/payments` | Payments list + refund | list, refund | ❌ |
| A13 | `/subscriptions` | Subscriptions list | `GET /admin/subscriptions` | ❌ |
| A14 | `/subscription-plans` | Plan management | create, update plans | ❌ |
| A15 | `/analytics` | Charts (revenue, bookings, categories) | `/admin/analytics/*` | ❌ |
| A16 | `/settings` | Platform settings editor | `GET/PUT /admin/settings` | ❌ |
| A17 | `/reviews` | Review moderation | list, approve, hide | ❌ |
| A18 | `/reports` | User reports queue + CSV export | list, resolve, export | ❌ |
| A19 | `/support` | Support tickets | list, detail, reply, update | ❌ |
| A20 | `/support/[id]` | Ticket thread | messages + status update | ❌ |

**Admin sidebar — full nav to add:**
```
Dashboard | Employees | KYC | Bookings | Users | Categories | Services |
Payments | Subscriptions | Analytics | Reviews | Reports | Support | Settings
```

---

## 7. Master status matrix

| Feature | Backend | Website | Customer | Employee | Admin |
|---------|---------|---------|----------|----------|-------|
| Auth (login/register) | ✅ | 🔶 | ❌ | ❌ | ✅ |
| Public discovery | ✅ | ✅ | — | — | — |
| Search | ✅ | ❌ | — | — | — |
| User profile | ✅ | — | ❌ | — | — |
| Saved addresses | ✅ | — | ❌ | — | — |
| Employee profile | ✅ | read-only | — | ❌ | partial |
| KYC | ✅ | — | — | ❌ | ❌ |
| Categories | ✅ | read | — | — | ✅ |
| Services browse | ✅ | ✅ | — | — | ❌ |
| Services manage | ✅ | — | — | ❌ | ❌ |
| Availability | ✅ | — | ❌ (book flow) | ❌ | — |
| Create booking | ✅ | — | ❌ | — | — |
| Booking lifecycle | ✅ | — | ❌ | ❌ | partial |
| Reviews | ✅ | read-only | ❌ | ❌ | ❌ |
| Subscriptions/Razorpay | ✅ | — | — | ❌ | ❌ |
| Payments/refunds | ✅ | — | — | ❌ | ❌ |
| Notifications | ✅ | — | ❌ | ❌ | — |
| Chat + WebSocket | ✅ | — | ❌ | ❌ | — |
| Support tickets | ✅ | — | ❌ | — | ❌ |
| Reports/moderation | ✅ | — | ❌ | — | ❌ |
| Analytics | ✅ | — | — | ❌ | ❌ |
| Platform settings | ✅ | — | — | — | ❌ |
| File uploads | ✅ | — | ❌ | ❌ | — |
| Admin user mgmt | ✅ | — | — | — | ❌ |

**Totals:**
- Backend APIs: **137 endpoints** ✅
- Website screens: **12** (8 done, 2 to move, 2 missing)
- Customer screens: **27** (all pending)
- Employee screens: **27** (all pending)
- Admin screens: **20** (4 done, 1 partial, 15 pending)

---

## 8. Recommended build order

### Sprint 1 — App shells
1. Scaffold `web/customer` (port 3002) — auth, layout, middleware
2. Scaffold `web/employee` (port 3003) — auth, layout, middleware
3. Update `web/README.md` with 4-app setup

### Sprint 2 — Core marketplace loop
4. Customer: booking wizard (C14) + my bookings (C15–C16)
5. Employee: profile (E6) + services (E9–E12) + availability (E13)
6. Employee: booking inbox (E15–E16)
7. Admin: booking detail (A7)

### Sprint 3 — Account & onboarding
8. Customer: profile + addresses (C7, C11–C13)
9. Employee: KYC (E7) + onboarding wizard (E5)
10. Admin: KYC queue (A8) + employee detail (A4)
11. Website: search page (W6) + link "Book" to customer app

### Sprint 4 — Engagement
12. Customer: reviews (C20), rebook (C19), notifications (C21)
13. Employee: reviews (E17–E18), analytics (E24)
14. Both: chat (C22–C23, E26–E27) + WebSocket

### Sprint 5 — Money & ops
15. Employee: subscription + Razorpay (E19–E23)
16. Admin: payments, subscriptions, analytics (A12–A15)
17. Admin: users, services, reviews, reports, support (A9–A11, A17–A20)
18. Admin: settings (A16)

### Sprint 6 — Auth extras & polish
19. Customer + employee: forgot/reset password, verify email
20. Customer: support tickets (C24–C26)
21. Remove website mock fallbacks
22. Cross-app navigation links in headers

---

## 9. Copy-paste prompt — scaffold both new apps

```
Create two Next.js 15 apps in this monorepo:

1. web/customer (port 3002) — customer portal
2. web/employee (port 3003) — provider portal

Match web/admin exactly: Tailwind, shadcn-style UI components, React Query,
zod, react-hook-form, lucide-react.

Each app needs:
- lib/api-client.ts (copy from web/admin)
- lib/auth.ts with role-specific cookie (customer_access_token / employee_access_token)
- middleware.ts protecting all routes except auth pages
- services/auth.ts calling the correct login/register endpoint
- app/login/page.tsx, app/register/page.tsx
- app/(portal)/layout.tsx with sidebar
- app/(portal)/dashboard/page.tsx (placeholder)
- .env.example with NEXT_PUBLIC_API_URL
- package.json with "dev": "next dev --turbopack -p 3002" (or 3003)

Customer auth: POST /auth/login/customer, role must be "customer"
Employee auth: POST /auth/login/employee, role must be "employee"

Update web/README.md with all 4 apps and ports.
```

---

This is the complete inventory: **137 APIs**, **86 UI screens** across 4 apps, integration status, deferred scope, and build order. Save this as your master prompt doc, or tell me if you want it written to a file like `docs/UI_INVENTORY.md`.