# Deferred Backend Scope

Items intentionally out of scope for the audit remediation plan (2026-06-04). Revisit when product priorities change.

## Payments and money movement

- Customer booking checkout (Razorpay at booking time)
- User wallet, saved payment methods, COD
- Provider payouts, bank account linking, earnings reports
- Invoice PDF/email generation (`internal/modules/invoices` remains reserved)

## Search and discovery

- Elasticsearch or external search service
- Autocomplete, facets, saved searches, search analytics
- Rating and availability filters beyond current Postgres ILIKE queries

## Booking complexity

- Recurring booking series (rebook is one-off only)
- Counter-offers and dispute workflows
- Calendar/directions integrations

## Admin platform surface

- CMS (FAQ, blog, help articles)
- Marketing campaigns and bulk notification campaigns
- Feature flags and team RBAC beyond single admin role
- Full GDPR export/delete and compliance module

## Location

- Geocoding, reverse geocoding, maps provider integration
- Service-area polygons and heatmaps

## Employee profile depth

- Certifications module with admin review queue
- Photo gallery, insurance/background-check badges

## Customer UX data

- Favorites, saved providers, recent searches
- Guest booking and social login autofill

## Auth extras

- Phone OTP verification
- Social OAuth (Google/Facebook)
- Two-factor authentication and device/session management

## Notifications extras

- SMS delivery
- Per-user notification preferences and quiet hours
- Template management and scheduled sends
