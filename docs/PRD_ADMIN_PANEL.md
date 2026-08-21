# Go Connect Too — Admin Panel Product Requirements Document

**Document owner:** Product, Operations, and Web Engineering  
**Status:** Draft  
**Surface:** Responsive web application  
**Primary role:** `admin`  
**Suggested stack:** Next.js, TypeScript, TanStack Query, React Hook Form, Zod  
**Backend prefix:** `/api/v1`

## 1. Purpose

This document defines the internal admin panel used to operate Go Connect Too. The panel shall give authorized administrators one secure interface to monitor marketplace health, approve providers and KYC, manage users and bookings, moderate services and reviews, operate subscriptions and refunds, answer support tickets, and configure approved platform settings.

The admin panel is a separate web application. It is not part of the unified customer/provider mobile app.

## 2. Product context

Go Connect Too has three backend roles: customer, `employee` (product label Provider), and admin. Most required admin APIs already exist in the Go modular monolith, but the current workspace does not contain a verifiable `web/admin` implementation. The partial admin-screen statuses in `docs/UI_INVENTORY.md` are therefore treated as an aspirational or stale historical plan. This PRD treats the admin frontend as a product to build or restore against the existing API contract.

The first release supports one super-admin permission set. Team roles such as Support, Operations, Finance, and Moderator are future scope.

## 3. Goals

1. Give operations a complete view of marketplace health and work queues.
2. Make provider and KYC review fast, consistent, explainable, and auditable.
3. Allow safe intervention in users, services, bookings, subscriptions, payments, reviews, reports, and support.
4. Prevent accidental destructive or high-impact actions.
5. Keep admin actions attributable and reviewable.
6. Provide useful analytics without exposing secrets or unnecessary personal data.
7. Ensure every screen has explicit loading, empty, error, permission, and success states.

## 4. Non-goals for the first release

- Multi-admin RBAC and custom permissions.
- Admin account creation or invitation UI.
- CMS, blog, FAQ, or marketing campaign management.
- Bulk push/email campaigns.
- Customer booking payment reconciliation.
- Provider payouts, bank accounts, or settlement operations.
- Invoice generation.
- Feature-flag management.
- GDPR automation or legal case management.
- Reading or moderating private customer/provider chat.
- Editing infrastructure secrets through the browser.

## 5. Users and access model

### 5.1 Super administrator

The first-release admin can access every module in this PRD. Authorization is enforced by the backend `admin` role, not by hidden navigation alone.

### 5.2 Future operations roles

The information architecture should not prevent future separation into:

- Marketplace Operations.
- KYC Reviewer.
- Customer Support.
- Trust and Safety Moderator.
- Finance Operator.
- Read-only Analyst.

These roles are design considerations only and shall not be represented as working permissions before backend RBAC exists.

## 6. Information architecture

Primary navigation:

- Dashboard
- Providers
- KYC
- Bookings
- Users
- Categories
- Services
- Payments
- Subscriptions
- Subscription Plans
- Analytics
- Reviews
- Reports
- Support
- Audit Log
- Settings

The shell shall include:

- Current admin identity.
- Global navigation and breadcrumbs.
- Environment indicator outside production.
- Sign out.
- Session-expiry handling.
- Optional global search only after supported search endpoints are defined.

The Audit Log item shall remain unavailable until the backend read API exists.

## 7. Shared admin experience requirements

- **ADM-SHARED-001:** All protected routes shall require an authenticated admin session.
- **ADM-SHARED-002:** A non-admin or expired session shall be redirected to login without rendering protected data.
- **ADM-SHARED-003:** List screens shall support server-side pagination and the filters supported by the corresponding API.
- **ADM-SHARED-004:** Filter state should be represented in the URL so work queues can be refreshed and shared internally.
- **ADM-SHARED-005:** Mutations shall show pending state, disable duplicate submission, and invalidate affected queries after success.
- **ADM-SHARED-006:** High-impact actions shall require a confirmation dialog naming the target and consequence.
- **ADM-SHARED-007:** Reject, suspend, override, hide, resolve, and refund actions shall require a reason when the backend supports one.
- **ADM-SHARED-008:** API errors shall be converted to safe, actionable UI messages. Internal details shall not be rendered.
- **ADM-SHARED-009:** Dates shall be displayed in the admin’s configured timezone while preserving UTC semantics in requests.
- **ADM-SHARED-010:** Money shall be displayed from integer minor units with explicit currency.
- **ADM-SHARED-011:** Status badges and filters shall use canonical backend values.
- **ADM-SHARED-012:** Data exports shall clearly show applied filters and export time.
- **ADM-SHARED-013:** The UI shall not optimistically claim a financial or irreversible action succeeded before backend confirmation.

## 8. Functional requirements

### 8.1 Authentication and session

- **ADM-AUTH-001:** Admin login shall use `POST /auth/login/admin`.
- **ADM-AUTH-002:** The login form shall support email, password, validation, loading, and generic failure messages.
- **ADM-AUTH-003:** Admin tokens shall be stored using a secure web-session design approved by engineering. Production access tokens should not be exposed to arbitrary client scripts where an HTTP-only cookie/BFF design is available.
- **ADM-AUTH-004:** Refresh and session expiry shall be handled consistently across all routes.
- **ADM-AUTH-005:** Logout shall revoke the refresh token where available and clear local session data.
- **ADM-AUTH-006:** Authentication events shall be logged server-side without recording credentials.

### 8.2 Dashboard

- **ADM-DASH-001:** Dashboard shall show a time-stamped platform snapshot.
- **ADM-DASH-002:** Summary metrics shall include total/active customers, total/approved/pending providers, bookings by key status, active services, active subscriptions, successful payments, and recognized subscription revenue where available.
- **ADM-DASH-003:** Dashboard shall show actionable queue counts for pending provider approvals, pending KYC, open reports, and open/urgent support tickets where APIs provide them.
- **ADM-DASH-004:** Dashboard cards shall link to the relevant filtered work queue.
- **ADM-DASH-005:** The page shall distinguish booking quoted value from collected subscription payment revenue.
- **ADM-DASH-006:** `/admin/dashboard/summary` shall power the current operational snapshot. `/admin/analytics/*` shall power date-ranged trends and breakdowns; it shall not redefine snapshot metrics.
- **ADM-DASH-007:** Payment KPIs shall not be released until dashboard queries use the canonical `success` payment status and metric definitions have been reconciled.

### 8.3 Provider management

- **ADM-PROV-001:** Administrators shall be able to list providers with search, pagination, and verification-status filter.
- **ADM-PROV-002:** Provider detail shall show account state, profile, contact fields permitted for operations, verification state, KYC summary, services, subscription summary, ratings, and booking summary where available.
- **ADM-PROV-003:** Administrators shall be able to approve, reject, or suspend a provider.
- **ADM-PROV-004:** Rejection and suspension shall require confirmation and a reason suitable for provider communication.
- **ADM-PROV-005:** The UI shall not imply that provider approval also approves KYC unless the backend performs both actions atomically.
- **ADM-PROV-006:** Provider actions shall refresh all affected dashboard, provider, service, and queue data.

### 8.4 KYC review

- **ADM-KYC-001:** Administrators shall be able to list KYC submissions by status.
- **ADM-KYC-002:** KYC detail shall show provider identity context, submission timestamps, current status, and protected document links.
- **ADM-KYC-003:** Document access shall use short-lived authorized URLs and shall not expose storage credentials. KYC document viewing is blocked until backend requirement `BE-FILE-004` is implemented.
- **ADM-KYC-004:** Administrators shall be able to approve or reject a pending submission.
- **ADM-KYC-005:** Rejection shall require a clear reason displayed to the provider.
- **ADM-KYC-006:** The interface shall prevent duplicate decisions while a request is pending.
- **ADM-KYC-007:** KYC screens shall avoid browser caching and analytics capture of document content.

### 8.5 User management

- **ADM-USER-001:** Administrators shall be able to list users with role/status/search filters supported by the API.
- **ADM-USER-002:** User detail shall show identity, role, status, creation date, profile summary, and relevant marketplace activity returned by the backend.
- **ADM-USER-003:** Administrators shall be able to suspend or activate non-admin users.
- **ADM-USER-004:** Suspension shall require confirmation and a reason if supported.
- **ADM-USER-005:** The panel shall prevent attempts to suspend an admin account.
- **ADM-USER-006:** Account status changes shall not be represented as data deletion.

### 8.6 Categories

- **ADM-CAT-001:** Administrators shall be able to list, create, edit, and delete categories.
- **ADM-CAT-002:** Category forms shall validate required name/slug fields and active state supported by the API.
- **ADM-CAT-003:** Deletion shall require confirmation and shall explain or surface backend conflicts when services reference the category.
- **ADM-CAT-004:** Category mutations shall invalidate public discovery data indirectly through backend behavior or subsequent client refetch.

### 8.7 Service moderation

- **ADM-SVC-001:** Administrators shall be able to list all services with pagination and supported status/provider/category filters.
- **ADM-SVC-002:** Each row/detail shall show service, provider, category, price, duration, provider-controlled status, and moderation state available from the backend.
- **ADM-SVC-003:** Administrators shall be able to force-activate or deactivate a service.
- **ADM-SVC-004:** Deactivation shall require confirmation and a reason if supported.
- **ADM-SVC-005:** The UI shall distinguish provider deactivation from admin moderation where backend data permits.

### 8.8 Booking operations

- **ADM-BOOK-001:** Administrators shall be able to list bookings with status and pagination filters.
- **ADM-BOOK-002:** Booking detail shall show customer, provider, service, schedule, quoted amount, notes safe for admins, current status, and status history after backend requirement `BE-BOOK-017` is available. Before then, it shall show current status only.
- **ADM-BOOK-003:** Admin status override shall be available only on detail, not as an unconfirmed table action.
- **ADM-BOOK-004:** Override confirmation shall name current status, target status, affected booking, and operational consequence.
- **ADM-BOOK-005:** Override shall require a reason and shall create an audit record.
- **ADM-BOOK-006:** The UI shall warn that an override does not automatically charge, refund, or pay out funds.
- **ADM-BOOK-007:** Stale conflicts shall refetch the booking and require the admin to reconfirm.

### 8.9 Payments and refunds

- **ADM-PAY-001:** Administrators shall be able to list provider subscription payments with status, provider, plan/reference, amount, currency, and gateway identifiers.
- **ADM-PAY-002:** Payment detail shall distinguish order creation, verification, webhook reconciliation, refund state, and failure state where available.
- **ADM-PAY-003:** Refund shall be available only for backend-eligible successful payments.
- **ADM-PAY-004:** Refund confirmation shall include amount, currency, payment identifier, provider, and consequence.
- **ADM-PAY-005:** The panel shall wait for backend confirmation and then display pending or final refund status accurately.
- **ADM-PAY-006:** Duplicate refund submission shall be prevented.
- **ADM-PAY-007:** The panel shall not show customer booking payments because that capability is not in the first release.

### 8.10 Subscriptions and plans

- **ADM-SUB-001:** Administrators shall be able to list provider subscriptions with plan, status, dates, auto-renew, and provider.
- **ADM-SUB-002:** Administrators shall be able to filter subscriptions by supported statuses.
- **ADM-SUB-003:** Administrators shall be able to create and update subscription plans.
- **ADM-SUB-004:** Plan forms shall include name, integer minor-unit price, billing period, service limit, and active state supported by the backend.
- **ADM-SUB-005:** Plan updates shall warn about effects on existing subscriptions and shall not claim retroactive changes unless the backend explicitly applies them.
- **ADM-SUB-006:** The panel shall not invent plan deletion if only create/update endpoints exist.

### 8.11 Reviews and reports

- **ADM-MOD-001:** Administrators shall be able to list reviews in the moderation queue.
- **ADM-MOD-002:** Review context shall include rating, text, author, provider, booking reference, created time, and current visibility where available.
- **ADM-MOD-003:** Administrators shall be able to approve or hide a review.
- **ADM-MOD-004:** Hide shall require confirmation and a reason if supported.
- **ADM-MOD-005:** Administrators shall be able to list open/resolved reports and inspect target context.
- **ADM-MOD-006:** Administrators shall be able to resolve a report with a recorded outcome.
- **ADM-MOD-007:** Administrators shall be able to export report data as CSV.
- **ADM-MOD-008:** Moderation changes shall refresh public reputation data after backend processing.

### 8.12 Support

- **ADM-SUP-001:** Administrators shall be able to list support tickets with status and priority filters.
- **ADM-SUP-002:** Ticket detail shall show customer, subject, thread, timestamps, status, and priority.
- **ADM-SUP-003:** Administrators shall be able to reply and update status/priority.
- **ADM-SUP-004:** Reply submission shall prevent duplicate sends and preserve the draft after recoverable errors.
- **ADM-SUP-005:** Closing or resolving a ticket shall require confirmation when it prevents further customer interaction.
- **ADM-SUP-006:** The panel shall clearly distinguish customer-visible messages from any future internal notes.
- **ADM-SUP-007:** Admin replies shall not be described as customer-visible in the mobile product until the backend exposes a customer ticket-thread detail API.

### 8.13 Analytics

- **ADM-ANL-001:** Analytics shall support an explicit date range with sensible maximum bounds.
- **ADM-ANL-002:** The panel shall show platform summary, revenue over time, bookings over time, and category breakdown using existing APIs.
- **ADM-ANL-003:** Charts shall state metric definition, unit, date range, timezone, and data freshness.
- **ADM-ANL-004:** Subscription revenue and booking quoted value shall be presented as different metrics.
- **ADM-ANL-005:** Empty periods shall be shown as valid zero/empty datasets, not as request failures.
- **ADM-ANL-006:** CSV or image export is future scope unless a backend or approved client export is explicitly implemented.

### 8.14 Settings

- **ADM-SET-001:** Administrators shall be able to view and update approved `general` and provider-behavior settings.
- **ADM-SET-002:** Settings forms shall be schema-driven and validated before submission.
- **ADM-SET-003:** Infrastructure credentials shall not be editable through this screen.
- **ADM-SET-004:** Secret values returned as redacted shall never be overwritten with mask text.
- **ADM-SET-005:** Maintenance-mode activation shall require an explicit high-impact confirmation.
- **ADM-SET-006:** Settings changes shall create an audit record with safe before/after metadata.

### 8.15 Audit log

- **ADM-AUD-001:** The backend shall provide a paginated audit-log read endpoint before this module is enabled.
- **ADM-AUD-002:** Audit records shall include actor, action, target, timestamp, outcome, request ID, and safe metadata.
- **ADM-AUD-003:** The UI shall support date, actor, action, target type, and outcome filters supported by the API.
- **ADM-AUD-004:** Audit records shall be read-only.
- **ADM-AUD-005:** Sensitive request bodies, tokens, passwords, secrets, and KYC document contents shall never appear.

## 9. Backend dependencies and required changes

Existing APIs cover most first-release modules. The following backend work is required or must be resolved before UI completion:

1. Apply admin audit middleware consistently to provider, KYC, category, service, subscription-plan, and review moderation mutations.
2. Add an authorized audit-log list/detail API before enabling Audit Log navigation.
3. Align `/admin/dashboard/summary` with the canonical snapshot definitions in `BE-ADMIN-001`; reserve `/admin/analytics/*` for date-ranged trends and breakdowns.
4. Confirm whether provider-setting records are product behavior or infrastructure configuration; secrets must remain environment-controlled.
5. Add required reason fields to high-impact actions where operational policy requires them.
6. Confirm customer support ticket detail behavior; the current customer API may not expose a dedicated thread endpoint.
7. Keep the OpenAPI contract synchronized with all admin routes and response models.
8. Fix dashboard payment-status filters and define collected subscription revenue separately from quoted booking value.
9. Expose booking status history required by admin timelines.
10. Add authorized, short-lived KYC document-download access.

## 10. Data presentation requirements

- Tables shall use stable identifiers and server pagination.
- Columns shall prioritize operational decisions; secondary data belongs in detail views.
- Every list shall define default sort, supported filters, no-result state, and error recovery.
- Personal data shall be shown only where needed for the admin task.
- Long identifiers shall support copy without becoming the primary visual label.
- Timestamps shall include timezone context.
- Financial values shall include currency and never use floating-point formatting as the source value.
- Status history and audit data shall be immutable in the UI.

## 11. Security and privacy

- Production access shall require HTTPS.
- The backend shall enforce admin role on every admin API.
- The web app shall use a strict Content Security Policy and safe dependency practices.
- Tokens, secrets, KYC content, and personal data shall not appear in browser logs, analytics, or error reporting.
- KYC document links shall be short-lived and authorized.
- Login and mutation endpoints shall be protected by backend rate limiting.
- Session timeout behavior shall be documented and tested.
- Destructive and financial actions shall use re-authentication in a future security phase if supported.
- Admin screens shall be excluded from public indexing and caching.
- Audit data shall be retained according to an approved operations policy.

## 12. Accessibility and responsive behavior

- Core workflows shall meet WCAG 2.1 AA targets.
- All actions shall be keyboard accessible with visible focus.
- Dialogs shall trap focus and return focus on close.
- Tables shall have accessible names, headers, and usable small-screen alternatives.
- Status shall not rely on color alone.
- Forms shall associate labels, descriptions, and errors with inputs.
- The supported baseline is desktop and tablet. Mobile web shall remain functional for urgent review but is not the primary operational layout.

## 13. Performance and reliability

- Authenticated shell and primary lists should become usable within agreed p75/p95 targets on supported corporate connections.
- Navigation shall not wait on unrelated dashboard requests.
- List queries shall cancel or ignore stale responses when filters change quickly.
- Mutations shall be protected against duplicate submission.
- A failed optional chart shall not prevent other dashboard modules from rendering.
- Error boundaries shall provide a safe reload path without leaking details.
- Production releases shall include source maps only in protected error-reporting infrastructure.

## 14. Analytics and operational metrics

Privacy-safe admin-panel events may include:

- Admin login outcome.
- Module opened.
- Filter applied.
- Provider/KYC decision outcome.
- User status action outcome.
- Booking override outcome.
- Refund initiated/result.
- Moderation action outcome.
- Support reply/status outcome.
- Settings update outcome.

Event properties shall use IDs and categories, not passwords, message bodies, KYC content, or raw personal data.

Admin product success metrics:

- Median provider and KYC review time.
- Oldest pending provider/KYC item.
- Support first-response and resolution time.
- Report resolution time.
- Refund failure/duplicate rate.
- Admin task error rate.
- Percentage of mutating admin actions represented in audit logs.
- Admin-panel p95 page/API latency and frontend error rate.

## 15. Testing and quality gates

- Unit tests shall cover status labels, money/date formatting, permission guards, and action eligibility.
- Component tests shall cover filters, pagination, forms, confirmation dialogs, loading, empty, and error states.
- Contract tests shall validate admin services against OpenAPI.
- End-to-end tests shall cover:
  1. Admin login/session expiry.
  2. Provider approval/rejection/suspension.
  3. KYC document review and decision.
  4. User suspend/activate.
  5. Category CRUD and service moderation.
  6. Booking status override.
  7. Payment refund sandbox flow.
  8. Subscription plan create/update.
  9. Review/report moderation.
  10. Support reply/status update.
  11. Settings update and secret-redaction safety.
- Accessibility automation and keyboard-only checks shall run on critical screens.
- Typecheck, lint, unit tests, production build, and critical E2E tests shall pass before release.

## 16. Delivery phases

### Current baseline blockers

- The documented admin frontend is not present in the current workspace and must be created or restored.
- Dashboard and analytics endpoints have overlapping shapes and unresolved metric semantics.
- Payment-status filtering must be aligned with the canonical `success` status before finance KPIs ship.
- Booking history and protected KYC document-read contracts are not yet available to the UI.
- Audit coverage is incomplete and audit records have no authorized read API.

### Phase 0 — Foundation

- Create or restore the Next.js admin app.
- Implement secure auth/session, protected shell, API client, query provider, error handling, design primitives, data table, pagination, and confirmation dialog.
- Establish OpenAPI-derived or contract-checked types.

### Phase 1 — Marketplace operations

- Dashboard.
- Pending KYC, reports, and support queue counts required by `BE-ADMIN-010`.
- Providers and provider detail.
- KYC review.
- Bookings and override.
- Users.
- Categories and service moderation.

### Phase 2 — Finance and growth operations

- Payments and refunds.
- Subscriptions and plans.
- Analytics with canonical metric definitions.

### Phase 3 — Trust, safety, and support

- Review moderation.
- Reports and export.
- Support tickets and replies.
- Platform settings.

### Phase 4 — Audit and hardening

- Complete backend audit coverage and read API.
- Audit-log viewer.
- Security review, accessibility, performance, monitoring, and production rollout.

## 17. Launch acceptance criteria

The admin panel is ready for production when:

1. Every protected route rejects unauthenticated and non-admin access.
2. Required marketplace-operation modules work against documented backend endpoints.
3. Provider and KYC decisions clearly capture and communicate reasons.
4. Booking overrides and refunds require explicit confirmation and backend confirmation.
5. Subscription revenue is not confused with uncollected booking value.
6. Sensitive settings and KYC data are not leaked through UI, logs, analytics, or caching.
7. All mutating admin actions in scope produce an audit record.
8. Critical queues support pagination, filters, loading, empty, error, and stale-state recovery.
9. Accessibility, typecheck, lint, production build, and critical E2E checks pass.
10. Operations has approved policies and runbooks for suspension, KYC, moderation, refund, booking override, and support.

## 18. Dependencies and open decisions

- Operations must define provider/KYC review standards and rejection reason templates.
- Finance must define refund permissions, limits, reason codes, and reconciliation ownership.
- Product must define whether customer and provider users appear in one user list or role-specific views.
- Security must approve admin token storage, session duration, production access controls, and KYC handling.
- Backend must provide complete audit coverage and an audit read API.
- Multi-admin RBAC requires a separate product and backend specification before additional admin users are delegated limited access.

## 19. Normative references

- `docs/PRD_BACKEND.md` — backend business rules and system scope.
- `internal/app/spec/openapi.yaml` — API wire contract.
- `internal/app/expected_routes.go` — route registry.
- `docs/UI_INVENTORY.md` — endpoint and historical planned-screen inventory; its current admin UI status is non-normative.
- `docs/DEFERRED_SCOPE.md` — capabilities excluded from the first release.
