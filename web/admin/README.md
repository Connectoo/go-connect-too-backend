# Go Connect — Admin Dashboard

Next.js admin dashboard connected to Go backend admin APIs.

## Setup

```bash
cp .env.example .env.local
npm install
npm run dev
```

Runs on [http://localhost:3000](http://localhost:3000) by default. Use port 3001 if the public website is also running:

```bash
npm run dev -- -p 3001
```

## Auth

- Sign in at `/login` with an **admin** user (role `admin` in the database).
- Uses `POST /api/v1/auth/login/admin` and stores the access token in `localStorage` + cookie for middleware.
- Non-admin users are rejected on the client after login.

Seed an admin user in PostgreSQL or via your internal tooling before first login.

## Protected routes

Middleware requires `admin_access_token` cookie for all routes except `/login`.
