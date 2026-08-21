# Go Connect Too — Unified Mobile App Product Requirements Document

**Document owner:** Product and Mobile Engineering  
**Status:** Draft  
**Platforms:** iOS and Android  
**Technology direction:** React Native CLI  
**Product model:** One app binary for customers and providers  
**Backend roles:** `customer` and `employee`  
**Product terminology:** Customer and Provider

## 1. Purpose

This document defines one Go Connect Too mobile application that serves both marketplace user types:

- **Customer mode:** discover and book services.
- **Provider mode:** offer services and manage work.

The app shall ship under one store listing, one bundle/application identity per platform, one release pipeline, and one shared design system. Customer and provider journeys remain role-specific after authentication.

This requirement supersedes the current two-app direction in `mobile/ARCHITECTURE.md`. Migration shall reuse the existing shared core, API, UI, and screen code where practical.

The customer and employee web portals proposed in `docs/UI_INVENTORY.md` are deferred for this product release. End-user customer and provider functionality is delivered by the unified mobile app; the public website and admin web panel remain separate surfaces.

## 2. Product vision

Go Connect Too gives a person one trusted destination to hire a service provider or operate as a provider. The app should make the user’s current role obvious, avoid mixing incompatible actions, and preserve backend authorization boundaries.

The first release supports one active role session at a time. Customer and provider accounts with the same email remain separate backend identities. Seamless linked accounts and simultaneous dual-role sessions are future work.

## 3. Goals

1. Deliver the complete customer and provider marketplace loop in one app.
2. Make role selection and role context clear before and after authentication.
3. Allow customers to discover, book, manage, communicate, and review.
4. Allow providers to onboard, publish services, manage availability, process jobs, and manage subscriptions.
5. Keep booking, payment, approval, and subscription status server-authoritative.
6. Provide reliable realtime and push updates without treating event payloads as the data source.
7. Provide a secure, accessible, resilient mobile experience suitable for production.

## 4. Non-goals for the first release

- Admin-panel access from the mobile app.
- A combined customer/provider backend identity.
- Silent switching between two role accounts.
- Customer booking checkout or stored payment methods.
- Provider payouts or bank account management.
- Live map tracking, navigation, or geocoded service areas.
- Phone OTP, social login, biometric-only login, or two-factor authentication.
- Favorites, saved searches, recurring bookings, or guest booking.
- Offline booking, offline messaging, or queued financial mutations.

Provider subscription checkout is in scope. Customer booking payment is not in scope until supported by the backend.

## 5. Target users

### 5.1 Customer

A person who wants to find a trusted provider, choose a service and time, create and manage a booking, communicate with the provider, and review completed work.

### 5.2 Provider

A person or business that wants to complete onboarding and KYC, publish services, define availability, respond to incoming bookings, complete jobs, communicate with customers, and maintain a platform subscription.

### 5.3 Returning dual-role person

A person may separately hold customer and provider accounts. The app may remember the last selected role, but switching role ends the current app session and opens the selected role’s authentication flow. Role switch always requires authentication in the first release. Account linking is not implied.

## 6. Product principles

- **One app, clear modes:** Shared binary does not mean a shared dashboard. Each role receives a focused navigation tree.
- **Server authority:** The app displays backend state and available actions; it does not invent booking, payment, verification, or subscription outcomes.
- **Progressive onboarding:** Ask for only the information required at each stage.
- **No false money states:** A gateway callback is not proof of payment. The backend-confirmed status controls UI.
- **Safe retries:** Booking and payment intents reuse an idempotency key.
- **Graceful degradation:** Cached reads may remain visible offline, but state-changing actions are blocked.

## 7. Information architecture

### 7.1 Application root

The root navigator shall render one of these mutually exclusive states:

1. **Boot:** Secure session hydration and configuration checks.
2. **Role selection:** “I need a service” or “I provide services.”
3. **Unauthenticated customer:** Customer login, registration, password recovery, and verification.
4. **Unauthenticated provider:** Provider login, registration, password recovery, and verification.
5. **Provider onboarding gate:** Authenticated provider who is incomplete, pending, or rejected.
6. **Authenticated customer shell.**
7. **Authenticated provider shell.**

The navigator shall derive role from the authenticated server session. A client preference may select which login form to show, but it shall never grant role access.

### 7.2 Customer navigation

Primary tabs:

- **Explore:** Home, search, category, service detail, provider profile.
- **Bookings:** Upcoming, active, past, cancelled, booking detail.
- **Inbox:** Conversations and notifications.
- **Account:** Profile, addresses, support, security, legal, role switch.

Modal or full-screen flows:

- Booking wizard.
- Cancel, reschedule, and rebook.
- Review composer.
- Report content or user.

### 7.3 Provider navigation

Primary tabs:

- **Jobs:** Pending, accepted, active, completed, job detail.
- **Services:** Service list, create/edit service.
- **Schedule:** Weekly availability.
- **Inbox:** Conversations and notifications.
- **Account:** Provider profile, KYC, reviews, subscription, payments, analytics, security, role switch.

Provider onboarding may temporarily replace the normal provider tabs until minimum eligibility conditions are met.

## 8. Role and session requirements

- **APP-ROLE-001:** First launch shall present a concise role-intent choice unless a valid session is restored.
- **APP-ROLE-002:** The selected intent shall route to the correct role-scoped login or registration endpoint.
- **APP-ROLE-003:** The app shall reject a token whose server-issued role does not match the selected role flow.
- **APP-ROLE-004:** Only one access/refresh token pair shall be active in the app session.
- **APP-ROLE-005:** Role switch shall warn that the active session will end, call logout, clear role-scoped cache, and return to role selection.
- **APP-ROLE-006:** The app may remember the last selected role as a non-sensitive preference.
- **APP-ROLE-007:** Admin credentials shall not be accepted in mobile role flows.
- **APP-ROLE-008:** Customer and provider query caches shall never be reused across role changes.

## 9. Shared functional requirements

### 9.1 Authentication

- **APP-AUTH-001:** Customers and providers shall be able to register with the fields required by their backend role.
- **APP-AUTH-002:** Users shall be able to sign in, sign out, request password reset, reset password, and change password.
- **APP-AUTH-003:** The app shall support email verification and resend verification.
- **APP-AUTH-004:** Access and refresh tokens shall be stored only in iOS Keychain or Android Keystore-backed storage.
- **APP-AUTH-005:** App boot shall remain in a non-interactive splash state until secure session hydration finishes.
- **APP-AUTH-006:** Concurrent `401` responses shall trigger one refresh request; queued calls may replay after success.
- **APP-AUTH-007:** Refresh failure shall clear credentials and return to the matching role login.
- **APP-AUTH-008:** Authentication forms shall show field-level validation and safe server messages.

### 9.2 Notifications and inbox

- **APP-COMM-001:** Both roles shall have a notification inbox with unread state, mark-one-read, and mark-all-read.
- **APP-COMM-002:** The app shall register an FCM/APNs device token after login and refresh registration when the platform token changes.
- **APP-COMM-003:** Logout and role switch shall disassociate the device token. This requirement is blocked until the backend provides the revoke operation defined by `BE-COMM-010`.
- **APP-COMM-004:** Push notifications shall deep-link only to screens authorized for the active account and role.
- **APP-COMM-005:** Both roles shall have booking-linked conversation lists and message threads.
- **APP-COMM-006:** New message and booking events shall invalidate query data and refetch authoritative REST responses.
- **APP-COMM-007:** The app shall show connection/offline state without representing a disconnected socket as an account failure.

### 9.3 Account and safety

- **APP-ACCT-001:** Users shall be able to view session identity, role, email-verification status, and account status.
- **APP-ACCT-002:** Customers shall be able to edit customer profile fields; providers shall be able to edit provider profile fields.
- **APP-ACCT-003:** Users shall be able to report supported content or users.
- **APP-ACCT-004:** Terms, privacy policy, support contact, and app version shall be available without exposing internal configuration.
- **APP-ACCT-005:** Suspended or inactive users shall see a blocking state with status-appropriate support or reactivation guidance and shall not access protected tabs.

## 10. Customer requirements

### 10.1 Discovery

- **APP-CUST-001:** The customer home shall show featured categories, providers, services, and relevant marketplace content.
- **APP-CUST-002:** Customers shall be able to browse services by category.
- **APP-CUST-003:** Customers shall be able to search providers and services by text.
- **APP-CUST-004:** Service detail shall show title, description, duration, price/quoted amount, provider summary, reviews, and booking CTA.
- **APP-CUST-005:** Provider detail shall show profile, verification indicators, offered services, rating, reviews, and public availability.
- **APP-CUST-006:** Empty, loading, retry, and pagination states shall be defined for all discovery lists.

### 10.2 Addresses

- **APP-CUST-007:** Customers shall be able to list, add, edit, delete, and select saved addresses.
- **APP-CUST-008:** Booking UI shall request an address only if the backend booking contract supports and requires it. The app shall not send unsupported placeholder address data.

### 10.3 Booking wizard

- **APP-CUST-009:** The booking flow shall start from an active service.
- **APP-CUST-010:** The wizard shall present service summary, date, available slot, optional notes, and final confirmation.
- **APP-CUST-011:** Available times shall come from the provider availability API and shall be refreshed before submission.
- **APP-CUST-012:** The app shall create one idempotency key when the user confirms and reuse it across retries for that intent. Automatic retry is blocked until the backend documents and implements `BE-BOOK-003`.
- **APP-CUST-013:** A slot conflict shall refresh availability and ask the customer to choose again.
- **APP-CUST-014:** Confirmation shall state that a booking request is created and that payment is not collected in the first release.
- **APP-CUST-015:** A successful booking shall open booking detail and invalidate relevant booking lists.

### 10.4 Booking management

- **APP-CUST-016:** Customers shall be able to filter bookings into useful groups such as upcoming, active, completed, and cancelled.
- **APP-CUST-017:** Booking detail shall show service, provider, schedule, quoted amount, notes, current status, and status timeline after `BE-BOOK-017` is available. Before then, the app shall show current status only and shall not invent history.
- **APP-CUST-018:** Cancel and reschedule actions shall appear only for server-eligible statuses.
- **APP-CUST-019:** Reschedule shall revalidate current availability.
- **APP-CUST-020:** Eligible past bookings shall support rebook preview and creation.
- **APP-CUST-021:** Stale or rejected transitions shall refetch booking detail and explain the current state.

### 10.5 Reviews and support

- **APP-CUST-022:** A completed, unreviewed booking shall offer a review action.
- **APP-CUST-023:** Review submission shall support validated rating, optional text, and optional media when uploads are configured.
- **APP-CUST-024:** Customers shall be able to create and list support tickets.
- **APP-CUST-025:** If the backend lacks customer ticket-thread detail, the first release shall show ticket summary/status and shall not simulate replies.

Reviews are separate entities linked to completed bookings; `reviewed` is not a booking status.

## 11. Provider requirements

### 11.1 Registration and onboarding

- **APP-PROV-001:** Provider registration shall clearly explain that approval and KYC are required before full marketplace operation.
- **APP-PROV-002:** Onboarding shall guide the provider through account verification, business profile, KYC, service setup, availability, and subscription requirements.
- **APP-PROV-003:** The provider shall be able to upload profile and KYC files through presigned upload flows.
- **APP-PROV-004:** The app shall show separate profile-verification and KYC statuses.
- **APP-PROV-005:** Pending providers shall see status, expected next step, and support access.
- **APP-PROV-006:** Rejected providers shall see the safe rejection reason and be able to correct and resubmit where the backend permits.
- **APP-PROV-007:** The app shall mirror the provider eligibility gate in `docs/PRD_BACKEND.md`: profile completion permits draft setup; approved verification, approved KYC, and active subscription permit service activation and new bookings. Existing accepted or in-progress jobs remain actionable after subscription expiry unless the account is suspended.

### 11.2 Provider profile and services

- **APP-PROV-008:** Providers shall be able to view and update business profile fields supported by the backend.
- **APP-PROV-009:** Providers shall be able to list, create, edit, delete, activate, and deactivate their own services.
- **APP-PROV-010:** Service forms shall validate category, title, description, duration, and integer minor-unit price.
- **APP-PROV-011:** The UI shall show subscription service limits and explain service-creation or activation gates returned by the backend.
- **APP-PROV-012:** Admin-deactivated services shall display a distinct state and shall not be represented as provider-controlled activation.

### 11.3 Availability

- **APP-PROV-013:** Providers shall be able to view and edit recurring weekly availability.
- **APP-PROV-014:** The editor shall prevent obvious invalid and overlapping time ranges before submission.
- **APP-PROV-015:** Server validation remains authoritative and conflicts shall be presented inline.

### 11.4 Job management

- **APP-PROV-016:** The Jobs area shall group bookings by actionable states.
- **APP-PROV-017:** Job detail shall show customer-safe details, service, schedule, notes, quoted amount, status, and status history after `BE-BOOK-017` is available. Before then, the app shall show current status only.
- **APP-PROV-018:** A pending job shall allow Accept or Reject.
- **APP-PROV-019:** An accepted job shall allow Start, Cancel, Reschedule, or No-show where the server permits.
- **APP-PROV-020:** An in-progress job shall allow Complete or No-show where the server permits.
- **APP-PROV-021:** Destructive actions shall require confirmation and an optional or required reason according to backend policy.
- **APP-PROV-022:** A conflict response shall refetch job state before offering another action.

### 11.5 Reviews, analytics, and subscriptions

- **APP-PROV-023:** Providers shall be able to list received reviews and reply where eligible.
- **APP-PROV-024:** Providers shall be able to view backend-supported summary, booking, and review analytics.
- **APP-PROV-025:** Providers shall be able to compare subscription plans and view service limits and billing period.
- **APP-PROV-026:** Subscription checkout shall create an order through the backend and launch the approved Razorpay mobile flow.
- **APP-PROV-027:** Gateway callback shall lead to a verifying state until backend verification or webhook reconciliation confirms the result.
- **APP-PROV-028:** Providers shall be able to view current subscription status, payment history, plan changes, cancellation, and auto-renew controls supported by the backend.
- **APP-PROV-029:** The app shall not show payout balances or payout requests in the first release.

## 12. API integration requirements

- The canonical base path is `/api/v1`.
- Customer auth shall use `/auth/register/customer` and `/auth/login/customer`.
- Provider auth shall use `/auth/register/employee` and `/auth/login/employee`.
- Product UI shall say Provider while API models may retain `employee`.
- Endpoint functions shall match the backend OpenAPI contract; existing mobile scaffold paths that use `/providers`, generic `/auth/login`, or non-canonical booking actions must be migrated.
- Provider search shall call `/search/employees`; public provider browsing shall use `/public/providers`; provider-owned operations shall use `/employee/*`.
- Generic `/auth/login` and `/auth/register` calls shall be removed. Client-supplied role fields shall not replace role-scoped auth endpoints.
- `/service-categories` shall migrate to `/public/categories` or `/categories` according to the screen contract.
- `/services/mine` shall migrate to `/employee/services`; provider booking transitions shall migrate to `/employee/bookings/{id}/*`.
- Notification device registration shall use `/device-tokens`.
- Customer `/payments/create-order` calls and provider `/payments/earnings` or `/payments/payouts` calls shall be removed. Provider subscription checkout shall use `/employee/subscriptions/create-order` and `/employee/subscriptions/verify-payment`.
- The HTTP layer shall unwrap the standard backend response envelope and throw a typed client error.
- Query keys shall be centralized. WebSocket and mutation success handlers shall invalidate keys instead of manually merging authoritative entities.
- Money shall remain integer minor units until formatted for display.
- No screen shall read or write tokens directly.

## 13. State, realtime, and offline behavior

### 13.1 State ownership

- TanStack Query owns server state.
- A small session store owns tokens, hydrated user identity, role, and authentication status.
- Local component state owns filters, modal state, and short-lived forms.
- Sensitive data shall not be copied into general-purpose persistent storage.

### 13.2 Realtime

- WebSocket authentication shall use the current access token.
- The socket shall reconnect with bounded exponential backoff.
- App backgrounding may disconnect the socket; foregrounding shall reconnect and refetch active data.
- Event payloads shall be treated as invalidation signals.

### 13.3 Offline

- Previously cached read data may be shown with a visible stale/offline indicator.
- Booking, job transition, chat send, KYC, service mutation, and subscription-payment actions shall be disabled offline.
- Mutations shall not be silently queued.
- Network retry shall not repeat non-idempotent user intents with a new idempotency key.

## 14. Design and accessibility

- The active role shall be visible in account settings and role-specific visual language, without relying on color alone.
- Shared components shall provide consistent spacing, typography, form feedback, buttons, dialogs, and status badges.
- Screens shall support Dynamic Type/font scaling, screen readers, logical focus order, adequate touch targets, and accessible labels.
- Text and controls shall meet WCAG AA contrast targets where applicable.
- Loading skeletons shall preserve layout; empty states shall provide a meaningful next action.
- Destructive, financial, booking, and role-switch actions shall require clear confirmation.
- Customer and provider terminology shall remain consistent across screens, notifications, support, and store copy.

## 15. Security and privacy

- Tokens shall use platform secure storage.
- No access token, refresh token, password, gateway signature, KYC document, or full sensitive payload shall be logged.
- Screens shall rely on backend authorization and handle `403` with a safe blocked state.
- Uploads shall enforce accepted type and size before transfer.
- KYC documents shall not be cached in the public image cache.
- Analytics and crash reporting shall exclude tokens, passwords, chat bodies, precise addresses, and KYC data.
- Payment success shall never be decided by local state.
- Screenshots may be restricted on KYC and payment screens where platform policy and usability permit.

## 16. Analytics events and product metrics

The app shall emit privacy-safe events for:

- Role selected.
- Registration started/completed by role.
- Login success/failure category.
- Provider onboarding step completed and KYC submitted.
- Search performed and result selected.
- Service viewed.
- Booking started, slot selected, submitted, conflicted, and created.
- Booking or job transition attempted and confirmed.
- Conversation opened and message-send outcome, without message content.
- Review submitted.
- Subscription plan viewed, checkout started, verification result confirmed.
- Push opened and deep-link destination.

Primary product metrics:

- Customer registration-to-first-booking conversion.
- Discovery-to-booking conversion.
- Booking acceptance and completion rates.
- Median provider response time.
- Provider registration-to-KYC and approval conversion.
- Approved-provider-to-first-service conversion.
- Provider subscription conversion and renewal.
- Crash-free sessions, cold-start time, API failure rate, and push-open success.

Numeric launch targets shall be approved after internal beta establishes baselines.

## 17. Performance and reliability targets

- Cold start shall reach an interactive role/auth/home state within the agreed platform target on supported mid-range devices.
- Scrolling lists shall remain responsive with pagination and virtualization.
- The app shall not block launch indefinitely on an unavailable WebSocket or optional analytics provider.
- A process restart shall restore a valid session or return safely to authentication.
- Unhandled errors shall be captured by a navigator-level error boundary with safe recovery.
- Network and server errors shall offer retry where repetition is safe.

## 18. Testing and quality gates

- Unit tests shall cover role/session transitions, booking action visibility, money formatting, API error mapping, and query invalidation.
- Component tests shall cover authentication, booking wizard, provider onboarding, service form, availability editor, job actions, and subscription verification states.
- Contract tests shall validate endpoint paths and payload types against OpenAPI.
- End-to-end tests shall cover:
  1. Customer registration, discovery, booking, cancellation/reschedule, completion view, and review.
  2. Provider registration, profile, KYC, approval test fixture, service creation, availability, accept/start/complete.
  3. Provider subscription checkout with sandbox verification.
  4. Role switch with credential and cache isolation.
  5. Push/deep-link authorization for each role.
- iOS and Android builds, lint, typecheck, unit tests, and critical E2E tests shall pass before store release.

## 19. Delivery phases

### Current baseline blockers

- The repository currently builds separate customer and provider apps.
- Several mobile endpoint paths and the `provider` wire role do not match the backend’s canonical routes and `employee` role.
- Booking idempotency, readable status history, and device-token revocation are not yet available as documented backend contracts.
- Legacy booking amount fields require a compatibility adapter until the backend minor-unit migration is complete.
- Existing mobile payment code references unsupported customer checkout, earnings, and payout endpoints and must not ship.

### Phase 0 — Unification and contract alignment

- Create one React Native application target and one root navigator.
- Define one production iOS bundle identifier, one Android application ID, and one store listing.
- Reuse shared `core`, `api`, and `ui` packages.
- Move reusable customer/provider screens into role feature areas.
- Implement role selection, one active session, role-aware theme, and cache isolation.
- Align every endpoint and model with backend OpenAPI.
- Retire the separate `apps/customer` and `apps/provider` production targets after migration.
- Replace customer/provider-specific build scripts with unified iOS, Android, typecheck, test, and release scripts.
- Remove the two-app role guard assumptions and verify that one binary can complete both role login flows.

### Phase 1 — Core marketplace

- Shared auth and account basics.
- Customer discovery, provider/service detail, booking wizard, and booking management.
- Provider profile, onboarding gate, services, availability, and job lifecycle.

### Phase 2 — Engagement

- Chat, WebSocket invalidation, push registration, notifications, reviews, reports, and customer support.

### Phase 3 — Provider monetization and analytics

- Subscription plans, Razorpay checkout and verification, payment history, provider analytics, cancellation/change/auto-renew. This phase depends on Backend Phase 2 provider-monetization contracts.

### Phase 4 — Hardening and store launch

- Accessibility, offline degradation, crash reporting, performance, security review, E2E coverage, store assets, privacy disclosures, and staged rollout.

## 20. Launch acceptance criteria

The unified app is ready for release when:

1. One iOS app and one Android app contain both customer and provider flows.
2. Role selection always uses the correct backend auth endpoints.
3. One active role session is securely restored, cleared on logout, and isolated on role switch.
4. Customer discovery-to-booking and provider onboarding-to-job-completion journeys pass E2E testing.
5. Booking actions reflect and reconcile with the server-owned state machine.
6. Mobile endpoint paths and payloads match OpenAPI; no placeholder booking data is sent.
7. Provider subscription payment remains pending until backend confirmation.
8. Chat, notifications, WebSocket invalidation, and push deep links work for both roles.
9. Offline mode blocks unsafe mutations and clearly marks stale reads.
10. Accessibility, crash-free beta, performance, privacy, and store review checks pass.

## 21. Dependencies and open decisions

- Backend and mobile teams must approve the endpoint and terminology migration before feature development continues.
- Any future linked-account token exchange requires a separate security and backend specification; first-release role switch always re-authenticates.
- Design must provide one brand system with distinct but related customer/provider mode cues.
- Operations must provide customer cancellation, provider no-show, review moderation, and support policies.
- Customer booking checkout, provider payouts, maps, and account linking require separate PRDs and backend capability.

## 22. Normative references

- `docs/PRD_BACKEND.md` — backend behavior and scope.
- `internal/app/spec/openapi.yaml` — API wire contract.
- `docs/UI_INVENTORY.md` — API inventory and historical screen plan; its customer/employee web-portal plan and current UI-status claims are non-normative for this mobile release.
- `mobile/ARCHITECTURE.md` — reusable technical patterns; its two-app product decision is superseded by this PRD.
- `mobile/README.md` — current two-app setup only; it must be replaced during Phase 0.
- `mobile/packages/core`, `mobile/packages/api`, and `mobile/packages/ui` — current shared mobile foundation.
