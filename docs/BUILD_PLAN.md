# Go Connect — Web UI Build Plan

Execution plan for building all pending web UI screens across the 4 Next.js apps.

- **Source of truth for screens/APIs:** [`docs/UI_INVENTORY.md`](./UI_INVENTORY.md)
- **OpenAPI spec:** `internal/app/spec/openapi.yaml`
- **Deferred (no UI):** [`docs/DEFERRED_SCOPE.md`](./DEFERRED_SCOPE.md)
- **API base:** `http://localhost:8080/api/v1` (`NEXT_PUBLIC_API_URL`)
- **Execution style:** phased, with **parallel subagents** per app-section. Main thread coordinates + reviews.

---

## 0. Current repo reality (verified)

| App | Path | Port | State |
|-----|------|------|-------|
| Website | `web/website` | 3000 | Real app, partial (search W6 + cross-app links missing) |
| Admin | `web/admin` | 3001 | Real app, partial (A4, A7–A20 pending) |
| Customer | `web/customer` | 3002 | **Stray 0-byte file — must be deleted, then scaffolded** |
| Employee | `web/employee` | 3003 | **Does not exist — must be scaffolded** |

**Reference pattern lives in `web/admin`** and must be copied exactly:

- `lib/api-client.ts` — `apiRequest<T>(path, {token, params, ...})`; parses envelope `{ success, message, data, error }`; throws `ApiError`.
- `lib/auth.ts` — role-scoped cookie `<role>_access_token` + `localStorage` key `go_connect_<role>_auth`. **Tokens are never shared across apps.**
- `middleware.ts` — redirect to `/login` when no cookie; `PUBLIC_PATHS` whitelist.
- `services/*.ts` — pure API calls (no React), wrap with `authOptions()` for bearer.
- `hooks/use-*.ts` — TanStack Query (`useQuery`/`useMutation`, invalidate on success).
- `components/ui/*` — shadcn-style primitives; `components/admin/{data-table,confirm-dialog,pagination}`; `components/providers/query-provider.tsx`.
- Envelope types in `types/api.ts`; auth types in `types/auth.ts`.
- Stack: Next 15 + Turbopack, Tailwind, RHF + Zod, lucide-react, recharts, TanStack Query.

---

## 1. Scope

71 pending screens:

| App | Pending | IDs |
|-----|---------|-----|
| Customer | 27 | C1–C27 (all) |
| Employee | 27 | E1–E27 (all) |
| Admin | 16 | A4, A6 (partial), A7–A20 |
| Website | 3 | W6 (new), W11/W12 (move to customer) |

Respect deferred scope: **no UI** for customer checkout/wallet/payouts, invoices, OTP, social login, 2FA, CMS, campaigns, RBAC, GDPR, maps, favorites, guest booking, recurring series.

---

## 2. Delegation model

- **Main thread = coordinator.** Owns the shared foundation, inventory checklist, integration, and review. Never delegates Phase 0.
- **`generalPurpose` subagent per screen-group.** Each group is self-contained within one app (new apps have no cross-imports), so groups run in **parallel** safely. Each subagent receives: exact file list, the admin file(s) to copy from, the API contract (method/path/body) from `UI_INVENTORY.md`, and the cookie/role for that app.
- **`explore` subagents** for read-only codebase questions only.
- **Review gate:** main thread runs lint + pattern conformance check on each subagent's output before flipping the inventory status flag.

**Parallelism cap:** up to ~4 background subagents per phase (one per app-section). Subagents within a phase must not touch the same files.

---

## 3. Phases

### Phase 0 — Foundation (main thread, BLOCKING — no parallelism)

1. Delete stray `web/customer` file.
2. Scaffold `web/customer` (3002) and `web/employee` (3003) by copying from `web/admin`:
   - configs: `package.json` (port + name), `tsconfig.json`, `next.config.ts`, `postcss.config.mjs`, `tailwind.config.ts`, `eslint.config.mjs`, `.gitignore`, `next-env.d.ts`
   - `lib/` (api-client, utils) + `lib/auth.ts` (rename cookie/key/role helper per app)
   - `components/ui/*`, `components/providers/query-provider.tsx`, shared `components/admin/*` (rename to `components/shared/*` or keep)
   - `types/api.ts`, `types/auth.ts`
   - `middleware.ts` (swap cookie name + public paths)
   - `app/layout.tsx` (wrap QueryProvider), `app/login/page.tsx`, `app/register/page.tsx`, `app/(portal)/layout.tsx` + sidebar, `app/(portal)/dashboard/page.tsx` placeholder
   - `services/auth.ts` → correct endpoint (`/auth/login/customer` | `/auth/login/employee`, register too)
   - `.env.example` with `NEXT_PUBLIC_API_URL`
3. Run `npx shadcn@latest mcp init --client cursor` in **both** new apps.
4. `npm install` in both; smoke-test `npm run dev` boots on 3002/3003.
5. Update `web/README.md` to document all 4 apps + ports.

**Exit criteria:** both apps boot, login pages render, middleware redirects unauthenticated users.

### Phase 1 — Core marketplace loop (4 parallel subagents)

| Subagent | App | Screens |
|----------|-----|---------|
| P1-A | Customer | Booking wizard C14, my bookings C15–C16 |
| P1-B | Employee | Profile E6, services E9–E12, availability E13 |
| P1-C | Employee | Booking inbox + actions E15–E16, dashboard E14 |
| P1-D | Admin | Booking detail + status override A7, finish bookings list A6 |

> P1-B and P1-C are both in `web/employee`; assign non-overlapping files (B owns `services/`+`availability/`+`profile/`, C owns `bookings/`+`dashboard/`) or run sequentially if file overlap risk.

### Phase 2 — Accounts & onboarding (4 parallel subagents)

| Subagent | App | Screens |
|----------|-----|---------|
| P2-A | Customer | Profile C7, settings C8–C10, addresses C11–C13 |
| P2-B | Employee | KYC E7, onboarding wizard E5, security E8 |
| P2-C | Admin | KYC queue A8, employee detail A4, users A9–A10 |
| P2-D | Website | Search W6 + cross-app "Book"/"Sign in" links |

### Phase 3 — Engagement (3 parallel subagents)

| Subagent | App | Screens |
|----------|-----|---------|
| P3-A | Customer | Review C20, rebook C19, cancel/reschedule C17–C18, notifications C21 |
| P3-B | Employee | Reviews E17–E18, analytics E24, notifications E25 |
| P3-C | Both portals | Chat C22–C23, E26–E27 + WebSocket `/ws` client (shared hook pattern) |

### Phase 4 — Money & admin ops (3 parallel subagents)

| Subagent | App | Screens |
|----------|-----|---------|
| P4-A | Employee | Subscription E19–E22 + Razorpay checkout, payments E23 |
| P4-B | Admin | Payments A12, subscriptions A13, plans A14, analytics A15 |
| P4-C | Admin | Services A11, reviews A17, reports A18, support A19–A20, settings A16 |

### Phase 5 — Auth extras & polish (2 parallel subagents)

| Subagent | App | Screens |
|----------|-----|---------|
| P5-A | Customer + Employee | forgot/reset password, verify-email (C3–C5, E3–E4) |
| P5-B | Customer | Support tickets C24–C26, report modal C27 |

Then main thread: remove website `lib/mocks.ts` fallbacks, finalize cross-app header nav, full lint pass.

---

## 4. Per-screen build checklist (subagent contract)

1. Confirm screen ID + exact API (method/path/body) in `UI_INVENTORY.md`. **Do not invent endpoints.**
2. Add `types/*` for new response shapes.
3. Add `services/*.ts` function (pure, `authOptions()` for auth).
4. Add `hooks/use-*.ts` TanStack Query wrapper (invalidate on mutation success).
5. Add `app/.../page.tsx`; reuse `DataTable`/`ConfirmDialog`/`Pagination`; install missing shadcn via MCP.
6. Match existing naming, error display, loading states.
7. Add nav entry to the app sidebar when it's a top-level route.
8. Lint clean.

---

## 5. Guardrails

- Each app: own cookie (`customer_access_token` / `employee_access_token` / `admin_access_token`), own localStorage key. Never share.
- Login page must reject wrong-role accounts (mirror admin's `isAdminUser` check).
- No UI for deferred backend scope.
- Status enums for badges/filters come from `UI_INVENTORY.md §2`.
- Booking action availability follows the state machine in `UI_INVENTORY.md §2`.

---

## 6. Progress tracking

`UI_INVENTORY.md` status flags (✅/🔶/❌) are the live checklist. Flip a row only after the review gate passes.

| Phase | Status |
|-------|--------|
| 0 — Foundation | ✅ both apps scaffolded, build clean |
| 1 — Core loop | ✅ C14–C16, E6/E9–E16, A6–A7 — all 3 apps build clean |
| 2 — Accounts & onboarding | ✅ C7–C13, E5/E7/E8, A4/A8–A10, W6 — all 4 apps build clean |
| 3 — Engagement | ✅ C17–C23, E17–E18/E24–E27 — customer + employee build clean |
| 4 — Money & admin ops | ✅ E19–E23, A11–A20 — employee + admin build clean |
| 5 — Auth extras & polish | ☐ |
