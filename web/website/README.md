# Go Connect — Public Website

Next.js public marketing and discovery site for the service marketplace.

## Stack

- Next.js 15, TypeScript, Tailwind CSS
- shadcn-style UI primitives
- TanStack Query, React Hook Form, Zod

## Setup

```bash
cp .env.example .env.local
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## API

Set `NEXT_PUBLIC_API_URL` to your Go API (default `http://localhost:8080/api/v1`).

Public endpoints used:

- `GET /public/home`
- `GET /public/categories`
- `GET /public/providers`
- `GET /public/providers/{id}`
- `GET /public/services`
- `POST /auth/register/customer`
- `POST /auth/login/customer`

If the API is unavailable, pages fall back to mock data in `lib/mocks.ts`.
