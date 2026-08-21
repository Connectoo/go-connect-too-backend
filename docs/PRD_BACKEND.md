# Go Connect Too — Backend Product Requirements Document

**Document owner:** Product and Backend Engineering  
**Status:** Draft  
**Product:** Go Connect Too service marketplace  
**System:** Go modular monolith API  
**Primary clients:** Unified customer/provider mobile app and admin web panel  
**API prefix:** `/api/v1`

## 1. Purpose

This document defines the product requirements for the Go Connect Too backend. The backend is the source of truth for identity, authorization, marketplace data, booking state, provider onboarding, subscriptions, payments, communication, moderation, and operational reporting.

The backend shall remain a modular monolith. Business modules may evolve independently inside one deployable Go application, but this PRD does not authorize a microservice split.

## 2. Product summary

Go Connect Too is a two-sided service marketplace:

- Customers discover local service providers, book available services, communicate with providers, manage bookings, and submit reviews.
- Providers register, complete their business profile and KYC, publish services, manage availability, process bookings, communicate with customers, and subscribe to a platform plan.
- Administrators approve providers and KYC, moderate marketplace content, manage users and bookings, operate subscriptions and refunds, answer support requests, configure the platform, and review analytics.

The mobile experience is delivered through one application binary. The backend continues to enforce distinct `customer` and `employee` roles. In product-facing copy, `employee` is presented as **Provider**.

## 3. Goals

1. Provide a secure, versioned API for all customer, provider, and administrator workflows.
2. Make the booking lifecycle authoritative, consistent, and auditable.
3. Support provider onboarding from registration through approval and service publication.
4. Support provider subscription monetization through Razorpay.
5. Deliver reliable in-app, push, email, WebSocket, and chat events where configured.
6. Give administrators the controls required to operate and moderate the marketplace.
7. Preserve data integrity under retries, concurrent booking actions, duplicate webhooks, and delayed external callbacks.
8. Provide sufficient observability, auditability, and test coverage for production operation.

## 4. Non-goals for the first release

The following are not part of the initial release unless separately approved:

- Customer booking checkout, wallet, saved cards, or cash on delivery.
- Provider payouts, bank account linking, settlement reconciliation, or tax invoices.
- Recurring booking series, counter-offers, or formal dispute resolution.
- Elasticsearch, geocoding, maps, service-area polygons, or route tracking.
- Favorites, saved searches, guest booking, or account linking across roles.
- Phone OTP, social sign-in, two-factor authentication, or device session management.
- SMS delivery, user-controlled quiet hours, or notification template management.
- Multi-admin RBAC, CMS, marketing campaigns, feature flags, or GDPR automation.

Booking records may store a quoted amount, but the first release must not represent that amount as a successfully charged customer payment.

## 5. Actors and authorization

### 5.1 Customer

A customer can manage only their own profile, addresses, bookings, reviews, conversations, notifications, reports, and support tickets. A customer cannot call provider or admin operations.

### 5.2 Provider

The API role is `employee`; the product label is Provider. A provider can manage only their own profile, KYC submission, services, availability, bookings, reviews, subscription, payment history, analytics, conversations, and notifications.

Provider capabilities may be gated by account status, profile completeness, verification status, KYC status, and subscription limits.

### 5.3 Administrator

An administrator can use protected administrative endpoints to operate the marketplace. The first release has one `admin` role with super-admin permissions. Admin accounts cannot be suspended through normal user-management operations.

### 5.4 Public user

An unauthenticated user can browse active categories, providers, services, reviews, availability, and search results. Authentication is required before a booking or report is created.

## 6. Functional requirements

### 6.1 Authentication and account lifecycle

- **BE-AUTH-001:** The system shall provide separate registration endpoints for customers and providers.
- **BE-AUTH-002:** The system shall provide role-scoped login endpoints for customers, providers, and administrators.
- **BE-AUTH-003:** Successful authentication shall issue a short-lived access token and a rotating refresh token.
- **BE-AUTH-004:** The authenticated identity and role shall be derived from the signed token, never from a client-supplied role on a protected request.
- **BE-AUTH-005:** Refresh token rotation shall invalidate the previously used refresh token.
- **BE-AUTH-006:** Logout shall revoke the submitted refresh token.
- **BE-AUTH-007:** The system shall support email verification, resend verification, forgot password, reset password, and authenticated password change.
- **BE-AUTH-008:** Password reset and email verification responses shall not reveal whether an email is registered.
- **BE-AUTH-009:** An account status of `suspended` or `inactive` shall prevent protected business actions.
- **BE-AUTH-010:** The same email may exist once per role. Customer and provider identities with the same email remain separate accounts in the first release.

### 6.2 Public discovery and search

- **BE-DISC-001:** Public home data shall include featured categories, providers, services, and platform statistics suitable for a mobile home screen.
- **BE-DISC-002:** Only active categories, active services, and approved/eligible providers shall appear in public discovery.
- **BE-DISC-003:** Users shall be able to browse and filter services by category.
- **BE-DISC-004:** Users shall be able to view a provider profile, offered services, rating summary, reviews, and public availability.
- **BE-DISC-005:** Search shall support text search for services and providers with pagination.
- **BE-DISC-006:** Search shall use PostgreSQL-backed matching for the first release. The API must not promise autocomplete, distance ranking, or advanced facets.
- **BE-DISC-007:** Public list endpoints shall enforce bounded page sizes and deterministic ordering.

### 6.3 Customer profile and addresses

- **BE-CUST-001:** Customers shall be able to read and update their name and phone number.
- **BE-CUST-002:** Customers shall be able to deactivate their own account.
- **BE-CUST-003:** Customers shall be able to create, list, update, and delete saved addresses.
- **BE-CUST-004:** Address operations shall be ownership-scoped.
- **BE-CUST-005:** Deactivation shall preserve records needed for bookings, payments, moderation, and audit history.

### 6.4 Provider profile, verification, and KYC

- **BE-PROV-001:** Providers shall be able to read and update their business profile, including bio, skills, languages, location, and profile image.
- **BE-PROV-002:** The system shall calculate or expose profile completeness for onboarding and publication gates.
- **BE-PROV-003:** Providers shall be able to submit identity and address evidence using references to uploaded files.
- **BE-PROV-004:** A provider shall be able to view the current KYC status and rejection reason.
- **BE-PROV-005:** Administrators shall be able to list, inspect, approve, or reject KYC submissions.
- **BE-PROV-006:** Administrators shall be able to approve, reject, or suspend provider profiles.
- **BE-PROV-007:** Rejections shall require a reason suitable for display to the provider.
- **BE-PROV-008:** Provider verification and KYC transitions shall create audit records and notifications.
- **BE-PROV-009:** Protected KYC documents shall not be exposed through public provider responses.

#### Provider eligibility gate

The first-release gate is normative:

- An active authenticated provider may edit their profile, submit KYC, compare subscription plans, and configure availability during onboarding.
- Creating a draft service requires a complete provider profile.
- Activating a service and appearing in public discovery require an active account, approved provider verification, approved KYC, an active subscription, and available capacity under the plan’s active-service limit.
- New bookings may be created only for publicly eligible active services.
- Subscription expiry or suspension blocks new service activation and new bookings, but a provider may complete or close bookings already in `accepted` or `in_progress` state.
- Account suspension blocks all protected business actions except the minimum support and session actions approved by security.

### 6.5 Categories and service listings

- **BE-SVC-001:** Administrators shall be able to create, update, and delete service categories.
- **BE-SVC-002:** Providers shall be able to create, read, update, delete, activate, and deactivate their own services.
- **BE-SVC-003:** A service shall include provider, category, title, description, duration, price in minor currency units, and active status.
- **BE-SVC-004:** Service writes shall validate category eligibility, positive duration, non-negative price, and ownership.
- **BE-SVC-005:** Service creation and activation shall enforce the provider eligibility gate above.
- **BE-SVC-006:** Active service count shall not exceed the provider’s subscription-plan limit.
- **BE-SVC-007:** Administrators shall be able to force-activate or deactivate a listing.
- **BE-SVC-008:** Administrative service moderation shall preserve the provider’s record and create an audit event.

### 6.6 Availability

- **BE-AVL-001:** Providers shall be able to create, list, update, and delete recurring weekly availability slots.
- **BE-AVL-002:** A slot shall include day of week, start time, and end time.
- **BE-AVL-003:** End time must be later than start time.
- **BE-AVL-004:** Overlapping slots for the same provider shall be rejected.
- **BE-AVL-005:** Public availability shall expose only slots for providers and services eligible for booking.
- **BE-AVL-006:** Booking creation shall revalidate the requested time against current availability and conflicting bookings.

### 6.7 Booking lifecycle

- **BE-BOOK-001:** An authenticated customer shall be able to create a booking for an active service and available time.
- **BE-BOOK-002:** The server shall derive provider, duration, quoted amount, and ownership from authoritative service data.
- **BE-BOOK-003:** Booking creation shall be idempotent when the client supplies an idempotency key.
- **BE-BOOK-004:** Concurrent requests for the same provider and time shall not create conflicting bookings.
- **BE-BOOK-005:** Customers shall be able to list and view only their bookings.
- **BE-BOOK-006:** Providers shall be able to list and view only bookings assigned to them.
- **BE-BOOK-007:** Customers may cancel or reschedule a `pending` or `accepted` booking.
- **BE-BOOK-008:** Providers may accept or reject a `pending` booking.
- **BE-BOOK-009:** Providers may start an `accepted` booking.
- **BE-BOOK-010:** Providers may complete an `in_progress` booking.
- **BE-BOOK-011:** Provider cancel, no-show, and reschedule actions shall follow the server-owned transition policy.
- **BE-BOOK-012:** Administrators may override status only through an explicit admin endpoint with an audit record.
- **BE-BOOK-013:** Every status transition shall append booking status history with actor, timestamp, previous status, new status, and reason when supplied.
- **BE-BOOK-014:** Customers shall be able to request a rebook preview and create one new booking from an eligible prior booking.
- **BE-BOOK-015:** Rebooking shall copy service context but revalidate service activity, provider eligibility, price, and availability.
- **BE-BOOK-016:** A booking conflict or stale transition shall return a stable conflict response suitable for client refetch.
- **BE-BOOK-017:** Role-scoped booking detail responses, or a dedicated history endpoint, shall expose ordered status history required by customer, provider, and admin timelines.

The first-release state model is:

`pending → accepted → in_progress → completed`

Terminal or alternate states are `rejected`, `cancelled`, and `no_show`. The server is the only authority for allowed transitions.

### 6.8 Reviews and reputation

- **BE-REV-001:** Only the customer who owns a completed booking may review it.
- **BE-REV-002:** A booking may have at most one customer review.
- **BE-REV-003:** A review shall include a validated rating and may include text and uploaded media references.
- **BE-REV-004:** Providers shall be able to list reviews received and submit one reply per review.
- **BE-REV-005:** Public provider responses shall include only visible reviews.
- **BE-REV-006:** Administrators shall be able to approve or hide reviews.
- **BE-REV-007:** Provider aggregate rating and review count shall be recalculated consistently after review or moderation changes.

### 6.9 Provider subscriptions and payments

- **BE-SUB-001:** The system shall expose active subscription plans publicly.
- **BE-SUB-002:** A plan shall define name, price in minor units, billing period, service limit, and active status.
- **BE-SUB-003:** Providers shall be able to create a Razorpay order for an eligible plan.
- **BE-SUB-004:** Payment verification shall be performed server-side using gateway signatures.
- **BE-SUB-005:** Razorpay webhooks shall be signature-verified, persisted, and processed idempotently.
- **BE-SUB-006:** A successful verified payment shall activate the corresponding provider subscription.
- **BE-SUB-007:** Providers shall be able to view the current subscription, payment history, change plan, cancel, and toggle auto-renew where supported.
- **BE-SUB-008:** Administrators shall be able to create and update plans and inspect all subscriptions.
- **BE-SUB-009:** Administrators shall be able to inspect payments and initiate supported refunds.
- **BE-SUB-010:** Refund state shall be reconciled with gateway responses and duplicate refund attempts shall be prevented.
- **BE-SUB-011:** Customer booking amounts shall remain separate from provider subscription payments and shall not be included in recognized payment revenue.

### 6.10 Notifications, realtime, and chat

- **BE-COMM-001:** The system shall create in-app notifications for important booking, KYC, subscription, payment, support, and moderation events.
- **BE-COMM-002:** Authenticated users shall be able to list notifications, mark one as read, and mark all as read.
- **BE-COMM-003:** Authenticated customer and provider devices shall be able to register FCM device tokens.
- **BE-COMM-004:** Push and email delivery shall degrade safely when providers are not configured; in-app notification creation must still succeed.
- **BE-COMM-005:** WebSocket connections shall require a valid JWT and shall deliver user-scoped events only.
- **BE-COMM-006:** A booking-linked conversation shall be available only to the booking customer and provider.
- **BE-COMM-007:** Conversation participants shall be able to list conversations, read messages, send messages, and mark messages as read.
- **BE-COMM-008:** Message and notification events shall contain identifiers sufficient for clients to refetch authoritative REST data.
- **BE-COMM-009:** Chat and notification writes shall prevent cross-account data access.
- **BE-COMM-010:** Authenticated users shall be able to revoke a device token on logout, role switch, or device-token rotation.

### 6.11 Support, reports, and moderation

- **BE-SAFE-001:** Customers shall be able to create and list support tickets.
- **BE-SAFE-002:** Administrators shall be able to list tickets, inspect threads, set priority/status, and reply.
- **BE-SAFE-003:** Authenticated customers and providers shall be able to report supported users or content.
- **BE-SAFE-004:** Administrators shall be able to list and resolve reports and export report data as CSV.
- **BE-SAFE-005:** Support and moderation actions shall preserve history and actor attribution.
- **BE-SAFE-006:** Internal notes, gateway errors, stack traces, and secrets shall never appear in user-facing responses.

### 6.12 Administration, analytics, and settings

- **BE-ADMIN-001:** `/admin/dashboard/summary` is the canonical current operational snapshot. It shall expose accurate customer, provider, booking, service, subscription, payment, and revenue KPIs. Payment counts shall use canonical payment statuses, and quoted booking value shall not be labeled as collected revenue.
- **BE-ADMIN-002:** Administrators shall be able to list and inspect users and activate or suspend non-admin accounts.
- **BE-ADMIN-003:** Analytics endpoints shall support bounded date ranges for revenue, bookings, category performance, and provider metrics.
- **BE-ADMIN-004:** Provider analytics shall include booking status, review, and summary metrics scoped to that provider.
- **BE-ADMIN-005:** Administrators shall be able to read and update approved platform settings.
- **BE-ADMIN-006:** Secret values shall be redacted on reads and never returned after writes.
- **BE-ADMIN-007:** Runtime environment variables remain the source of truth for infrastructure credentials. Database settings may control product behavior but shall not replace secrets loaded at process startup.
- **BE-ADMIN-008:** Every mutating admin route shall create an audit record with actor, action, target type, target identifier, timestamp, outcome, and safe request metadata.
- **BE-ADMIN-009:** An admin audit-log read endpoint is required before the admin panel’s audit viewer can be released.
- **BE-ADMIN-010:** The operational summary shall expose pending provider, pending KYC, open report, and open/urgent support queue counts, or document separate bounded endpoints used for those dashboard cards.
- **BE-ADMIN-011:** `/admin/analytics/*` endpoints are the canonical source for date-ranged trends and breakdowns; they shall not redefine the current snapshot.

### 6.13 File uploads

- **BE-FILE-001:** Authenticated users shall be able to request presigned upload details for approved content types and size limits.
- **BE-FILE-002:** Upload metadata shall record owner, purpose, media type, size, and storage key.
- **BE-FILE-003:** File deletion shall require ownership or admin authorization.
- **BE-FILE-004:** KYC file references shall be excluded from public APIs. Authorized admin reads shall use short-lived presigned download URLs or an authenticated proxy rather than permanent public URLs.
- **BE-FILE-005:** When object storage is disabled, the API shall return a clear capability error rather than accepting unusable file references.

## 7. API contract requirements

1. Responses shall use the platform envelope:

   `{ "success": true, "message": "...", "data": ... }`

   `{ "success": false, "message": "...", "error": "STABLE_CODE" }`

2. Error messages exposed to clients shall be safe and actionable. Internal database, network, gateway, and stack details must be logged only on the server.
3. Validation failures shall identify invalid fields where safe.
4. List endpoints shall use a consistent pagination shape and bounded limits.
5. Money shall be represented as integers in minor currency units.
6. Dates and timestamps shall have explicit formats and timezone semantics. Persisted event timestamps shall use UTC.
7. Breaking contract changes require a new API version or an approved compatibility period.
8. The OpenAPI document and route inventory shall be updated in the same change as an endpoint contract.
9. Mobile terminology may use Provider, but wire-level role and route compatibility shall remain `employee` until a versioned migration is approved.
10. Booking creation idempotency shall use a documented `Idempotency-Key` request header with defined scope, retention, replay response, and conflict behavior.
11. OpenAPI shall document request and response schemas, error codes, and pagination for every P0 mobile and admin journey.
12. Existing floating-point booking amount fields are a migration gap. New money contracts shall use integer minor units, with a versioned migration and compatibility plan for `total_amount` and rebook-price fields.

## 8. Data and consistency requirements

- PostgreSQL is the system of record.
- Foreign keys and check constraints shall protect core relationships and enum-like states.
- Service, booking, subscription, payment, refund, moderation, and status-history writes shall use transactions where partial success would corrupt state.
- Duplicate webhook event identifiers shall be rejected or treated as already processed.
- Destructive account actions shall not remove financial, booking, moderation, or audit evidence required for operations.
- Personally identifiable information shall be minimized in logs, analytics events, and exports.
- Production backups, restore tests, retention, and migration rollback procedures must be documented before launch.

## 9. Security and privacy requirements

- All production traffic shall use HTTPS.
- JWT, database, Razorpay, SMTP, FCM, and storage secrets shall come from secure environment configuration.
- Passwords shall use an approved adaptive password hash.
- Authorization checks shall be enforced in handlers/services regardless of client UI state.
- Login, password reset, verification resend, uploads, messaging, reports, and webhook endpoints shall be rate-limited appropriately.
- CORS shall allow only approved origins.
- Upload MIME type, extension, and size shall be validated.
- Admin and KYC access shall produce security-relevant logs.
- Logs shall not contain access tokens, refresh tokens, passwords, gateway signatures, KYC document contents, or raw infrastructure secrets.
- Public and mobile errors shall not expose database or internal implementation details.

## 10. Reliability and performance requirements

- Health checks shall report API and database readiness without exposing secrets.
- Core read endpoints should meet a p95 server response target of 500 ms under the agreed launch load, excluding external gateway latency.
- Core write endpoints should meet a p95 target of 800 ms under the agreed launch load, excluding external gateway latency.
- Webhook endpoints shall acknowledge valid events quickly and move non-critical follow-up work to background processing.
- Notification provider failure shall not roll back an already committed booking state transition.
- Database queries for paginated marketplace and admin lists shall have supporting indexes and avoid unbounded scans.
- The service shall support graceful shutdown and complete or cancel in-flight work safely.
- Background workers shall expose failure logs and retry only operations that are safe to repeat.

## 11. Observability and audit

The backend shall emit structured logs with request ID, actor ID where available, route, status, duration, and safe error code. It shall expose metrics for:

- Request rate, error rate, and latency by route group.
- Authentication failures and refresh failures.
- Booking creation, conflicts, and transition failures.
- Provider onboarding and KYC decision counts.
- Payment orders, verifications, webhook failures, and refunds.
- Notification delivery attempts and failures by channel.
- WebSocket connection count and reconnect/error rate.
- Worker queue or retry activity.

Alert thresholds and on-call ownership must be defined before production launch.

## 12. Testing and quality gates

- Service-layer tests shall cover success, validation, authorization, state conflict, and repository failure cases.
- Booking transition tests shall cover every allowed and denied transition for each actor.
- Repository integration tests shall cover constraints, transaction rollback, pagination, and concurrent slot conflicts.
- Auth tests shall cover refresh rotation, revoked tokens, suspended accounts, and cross-role access.
- Payment tests shall cover bad signatures, duplicate webhooks, delayed webhooks, failed payments, and duplicate refunds.
- Contract tests shall verify registered routes against OpenAPI.
- End-to-end tests shall cover provider onboarding, customer booking, provider completion, customer review, provider subscription purchase, and admin moderation.
- `gofmt`, static checks, and `go test ./...` must pass before release.

## 13. Success metrics

Backend product success shall be measured with:

- Booking API success rate excluding user validation errors.
- Booking conflict rate and duplicate booking count.
- Median time from provider registration to approval.
- Percentage of approved providers who publish at least one service.
- Subscription payment verification and webhook reconciliation success.
- Notification delivery success by channel.
- Support-ticket first-response and resolution times.
- API p95 latency and 5xx rate.
- Number of unauthorized cross-account access incidents, with a target of zero.

Final launch targets require product and operations approval after baseline load testing.

## 14. Release phases

### Current baseline blockers

The repository already implements much of this PRD, but the following gaps must be closed before the dependent UI is considered complete:

- Booking creation does not yet expose the required idempotency contract.
- Booking status history is persisted but not exposed by the documented API.
- Device tokens can be registered but not revoked through a documented endpoint.
- Legacy booking amount responses use floating-point values.
- The current service layer does not yet enforce the complete provider eligibility gate defined in this PRD.
- Admin audit middleware does not yet cover every mutating admin route, and audit logs have no read API.
- Dashboard payment/revenue queries and labels must be reconciled with canonical `success` payment status and quoted-booking-value semantics.
- Admin KYC responses need protected short-lived document access.
- OpenAPI lacks complete request/response schemas for some P0 journeys.

### Phase 1 — Contract alignment and core marketplace

- Freeze role terminology and mobile/backend endpoint mapping.
- Complete auth, discovery, profiles, provider onboarding, services, availability, and booking state machine.
- Add idempotency and conflict coverage for booking writes.
- Verify OpenAPI and mobile client models.
- Before releasing any admin mutation UI, apply audit coverage to provider, KYC, category, service, subscription-plan, review, report, support, booking, payment, user, and settings mutations in scope.

### Phase 2 — Engagement and provider monetization

- Release reviews, chat, notifications, WebSocket events, provider subscriptions, Razorpay verification, and payment history.
- Complete device-token lifecycle and email/push provider configuration.

### Phase 3 — Operations and hardening

- Complete admin operations, analytics, reports, support, settings, audit coverage, load tests, backup verification, monitoring, and incident runbooks.
- Add the audit-log read API required by the admin PRD.

### Future phases

- Customer checkout, provider payouts, invoices, advanced discovery, account linking, multi-admin RBAC, and compliance automation require separate approved specifications.

## 15. Launch acceptance criteria

The backend is release-ready when:

1. All P0 mobile and admin journeys have stable, documented API contracts.
2. Role and ownership tests prove customers, providers, and admins cannot cross protected boundaries.
3. The booking lifecycle passes state, concurrency, idempotency, and status-history tests.
4. Provider onboarding gates service publication as specified.
5. Razorpay signatures, webhooks, subscriptions, and refunds pass sandbox reconciliation tests.
6. In-app notifications work without optional providers; configured push/email delivery is verified.
7. OpenAPI matches registered routes and client expectations.
8. Database migrations run cleanly from an empty database and against the previous release.
9. Production secrets, backups, health checks, monitoring, rate limits, and runbooks are configured.
10. `go test ./...` and required static checks pass.
11. All mutating admin actions in release scope are audited, and the authorized audit-log read API is available before the admin Audit Log module launches.

## 16. Dependencies and open decisions

- Product must define provider approval and KYC service-level targets.
- Operations must define cancellation, no-show, refund, and moderation policies shown to users.
- Finance must define currency, tax display, refund eligibility, and revenue-recognition rules.
- Product and engineering must approve the mobile-to-backend route migration plan because the existing mobile client uses several non-canonical paths.
- Product must approve if and when one person may link customer and provider identities for seamless role switching.

## 17. Normative references

- `internal/app/spec/openapi.yaml` — wire contract.
- `internal/app/expected_routes.go` — registered route expectation.
- `docs/UI_INVENTORY.md` — API/route inventory and historical screen plan; its UI status and customer/employee web-portal plan are non-normative for this release.
- `docs/DEFERRED_SCOPE.md` — currently deferred backend capabilities.

If this PRD conflicts with an approved versioned API contract, the approved contract controls wire behavior until the conflict is resolved through product and engineering change control.
