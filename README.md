# QuoteTrack

> Never forget a customer you already quoted.

QuoteTrack is a simple quote and follow-up tracker for small businesses. Create quotes for your customers, send them a shareable link, set a follow-up date, and get reminded to reach out before the deal goes cold.

## Features

- **Auth**: email + password signup and login (bcrypt + JWT), business profile settings.
- **Customers**: add, edit, and delete customers with contact details; see how much you've quoted them.
- **Quotes**: create quotes with line items (server-computed totals), sequential per-business quote numbers (`QT-0001`, `QT-0002`, …), statuses (`draft`, `sent`, `waiting`, `won`, `lost`, `expired`), quote/expiry dates, and notes.
- **Share links**: every quote gets a public link you can send to a customer (or via WhatsApp) — no login needed.
- **Follow-ups**: set a follow-up date on any quote. The dashboard shows what's due today; snooze reminders by a week, and the quote drops out when marked won/lost.
- **Dashboard**: headline stats (quotes, waiting, follow-ups due, won, lost) plus the follow-up list.
- **Analytics**: win rate, won/lost/waiting counts and values, pipeline-value breakdown.

## Tech stack

| Layer     | Technology |
|-----------|------------|
| Backend   | Go 1.22, net/http stdlib router, pgx (PostgreSQL driver), golang-jwt, bcrypt |
| Database  | PostgreSQL 16 (migrations embedded in the Go binary) |
| Frontend  | React 19, TypeScript, Vite, React Router |

Monorepo layout:

```
quotetrack/
├── backend/              # Go HTTP API
│   ├── cmd/server/       # entrypoint
│   └── internal/
│       ├── auth/         # tokens, password hashing, auth handlers
│       ├── config/       # env-driven config
│       ├── customers/    # customer CRUD + stats
│       ├── dashboard/    # stats, follow-ups, analytics
│       ├── database/     # pool + embedded migrations
│       ├── httpapi/      # JSON helpers
│       ├── middleware/   # CORS, logging, recovery
│       ├── models/       # shared types (money is integer cents)
│       ├── publicquote/  # unauth share-link view
│       ├── quotes/       # quote CRUD, status, follow-ups
│       └── server/       # route wiring
├── frontend/             # React SPA (Vite)
│   └── src/
│       ├── pages/        # login, dashboard, customers, quotes, analytics, settings, public quote
│       ├── components/   # modals, badges, line-item editor, follow-up cards
│       ├── services/     # typed API client (localStorage session)
│       └── ...
├── docker-compose.yml    # PostgreSQL only
└── LICENSE
```

## Getting started

### 1. Database

The only infrastructure needed is PostgreSQL 16:

```bash
docker compose up -d postgres
```

This creates a `quotetrack` database with user `quotetrack` / `quotetrack` on port `5432` (see `docker-compose.yml`). Tables are created automatically on first backend start via embedded migrations.

### 2. Backend

```bash
cd backend
go run ./cmd/server
```

Configuration via environment variables (see `backend/.env.example`):

| Variable            | Default                                              |
|---------------------|------------------------------------------------------|
| `PORT`              | `8080`                                               |
| `DATABASE_URL`      | `postgres://quotetrack:quotetrack@localhost:5432/quotetrack?sslmode=disable` |
| `JWT_SECRET`        | `dev-only-secret-change-me` (set a real one!)        |
| `JWT_EXPIRY`        | `168h`                                               |
| `FRONTEND_ORIGIN`   | `http://localhost:5173`                              |

> The backend and frontend use **JWT auth** by default. If you later want Supabase auth, the token manager in `internal/auth` is the only place that needs to change.

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173. Vite proxies `/api` → `http://localhost:8080` (configurable via `VITE_API_PROXY_TARGET`; see `frontend/.env.example`).

To build for production:

```bash
cd frontend && npm run build   # outputs static site in frontend/dist
```

## API overview

Authenticated routes require `Authorization: Bearer <jwt>`.

| Method | Path                          | Description                          |
|--------|-------------------------------|--------------------------------------|
| POST   | `/api/auth/signup`            | Create account, returns token        |
| POST   | `/api/auth/login`             | Log in, returns token                |
| GET    | `/api/me`                     | Current user                         |
| PATCH  | `/api/me`                     | Update business name / email         |
| GET    | `/api/customers`              | List customers (`?q=` searches)      |
| POST   | `/api/customers`              | Create customer                      |
| GET    | `/api/customers/{id}`         | Get customer + stats                 |
| PUT    | `/api/customers/{id}`         | Update customer                      |
| DELETE | `/api/customers/{id}`         | Delete customer                      |
| GET    | `/api/quotes`                 | List quotes (`?status=&search=&customer_id=`) |
| POST   | `/api/quotes`                 | Create quote (line items, dates)     |
| GET    | `/api/quotes/{id}`            | Get quote with items                 |
| PUT    | `/api/quotes/{id}`            | Update quote + items                 |
| DELETE | `/api/quotes/{id}`            | Delete quote                         |
| PATCH  | `/api/quotes/{id}/status`     | Change status (won/lost clears follow-up) |
| PATCH  | `/api/quotes/{id}/follow-up`  | Set/clear follow-up date             |
| GET    | `/api/dashboard`              | Stats + follow-ups due               |
| GET    | `/api/analytics`              | Win rate and won/lost/waiting values |
| GET    | `/api/public/quotes/{id}`     | Unauthenticated share view           |

Money is handled server-side as integer cents and serialized as dollars (e.g. `19.99`).

## Testing

Backend unit + integration tests:

```bash
cd backend
go test ./...
```

The `backend/internal/quotes` and `internal/auth` packages include unit tests for totals, quote numbering, password hashing, and JWT behavior. An end-to-end smoke test against a real PostgreSQL (`/tmp/opencode/integration_test.sh` in a dev environment) exercises signup, customer CRUD, quote lifecycle, follow-ups, analytics, and the public share view.

## Roadmap

- [x] Phase 1 — Backend scaffold (config, DB, migrations)
- [x] Phase 2 — Models + HTTP helpers + middleware
- [x] Phase 3 — Auth (JWT signup/login/me)
- [x] Phase 4 — Customers API
- [x] Phase 5 — Quotes API (create/update/status/follow-up/public)
- [x] Phase 6 — Dashboard + analytics
- [x] Phase 7 — Frontend app (auth, dashboard, customers, quotes, analytics, settings, public quote)
- [x] Phase 8 — Root wiring (docker-compose, env examples, README)
- [ ] Sign/restore quotes from a saved draft
- [ ] Quote PDF export
- [ ] Recurring follow-up templates
- [ ] Email/SMS reminders (planned hook; share links ready for WhatsApp)

## License

MIT — see [LICENSE](LICENSE).