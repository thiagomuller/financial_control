# My Finances APP — Documentation

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture Decisions](#2-architecture-decisions)
3. [Directory Structure](#3-directory-structure)
4. [Infrastructure & Deployment](#4-infrastructure--deployment)
5. [Backend — Go REST API](#5-backend--go-rest-api)
6. [Frontend — Angular UI](#6-frontend--angular-ui)
7. [Database Schema](#7-database-schema)
8. [Security Design](#8-security-design)
9. [Recurring Transactions Scheduler](#9-recurring-transactions-scheduler)
10. [Testing Strategy](#10-testing-strategy)
11. [API Reference](#11-api-reference)
12. [Milestone 2 Changes](#12-milestone-2-changes)
13. [Milestone 3 Changes](#13-milestone-3-changes)

---

## 1. Project Overview

My Finances APP is a self-hosted personal finance manager. It tracks bank accounts, categorised transactions, transfers between accounts, savings goals, and repeatable transactions/transfers. A background scheduler automatically materialises repeatable templates as real records each month. The stack is:

- **Backend**: Go 1.25.8 — standard library only (+ `github.com/lib/pq` for PostgreSQL)
- **Frontend**: Angular 21 — standalone components, reactive forms, lazy-loaded routes
- **Database**: PostgreSQL (latest) — single persistent volume mounted in the project directory
- **Container runtime**: Podman (rootless) + podman-compose via `distrobox-host-exec`

---

## 2. Architecture Decisions

### 2.1 Go Standard Library Only

The goal was to keep the dependency tree small and auditable. Go 1.22 introduced method-qualified patterns in `net/http.ServeMux` (e.g., `"GET /api/users/{id}"`), which eliminates the primary reason developers reach for third-party routers. The only external packages are:

- `github.com/lib/pq` — the canonical PostgreSQL driver; there is no stdlib alternative for `database/sql` + Postgres.
- `github.com/testcontainers/testcontainers-go` — integration testing only; spins up a real PostgreSQL container so tests never mock the database.

Everything else — JWT, password hashing, CORS middleware, pagination, JSON encoding — is implemented using stdlib packages.

### 2.2 Manual JWT Implementation

Rather than pulling in a JWT library (adds transitive dependencies and version drift), JWT was implemented using three stdlib packages:

```
crypto/hmac  →  HMAC construction
crypto/sha256  →  hash function (HS256)
encoding/base64  →  base64url-encoding of header, payload, signature
```

The implementation in `my_finances_api/internal/auth/jwt.go` follows the JWT spec exactly:
1. Encode `{"alg":"HS256","typ":"JWT"}` header as base64url.
2. Encode claims (`sub`, `exp`, `iat`) as base64url.
3. HMAC-SHA256 sign `header.payload` with the secret, encode as base64url.
4. Join with `.` → `header.payload.signature`.

Validation uses `hmac.Equal` for constant-time comparison, which prevents timing-based signature leaks.

Tokens expire after 24 hours. The `JWT_SECRET` is read from an environment variable (defaulting to a dev placeholder that warns against production use in the compose file).

### 2.3 Password Hashing with SHA-256 + Random Salt

bcrypt is the industry standard but lives in `golang.org/x/crypto`, which is an external dependency. The chosen approach uses:

1. 16 bytes of `crypto/rand` as salt.
2. SHA-256 of `saltHex + plaintext`.
3. Stored as `saltHex:hashHex`.

Verification re-derives the hash from the stored salt and compares with `hmac.Equal` (constant-time). The random salt ensures identical passwords produce different hashes, neutralising rainbow table attacks. The trade-off versus bcrypt is that SHA-256 is much faster, making offline brute-force cheaper — but for a self-hosted personal app with no public login surface this is an acceptable cost.

### 2.4 PostgreSQL 18+ Volume Path

PostgreSQL changed its data directory layout in version 18. The data directory is now `/var/lib/postgresql` rather than `/var/lib/postgresql/data`. The volume mount in `docker-compose.yml` therefore uses:

```yaml
volumes:
  - ./db_data:/var/lib/postgresql:Z
```

The `:Z` suffix is a SELinux relabelling label required for rootless Podman on SELinux-enforcing systems. Without it, the container's PostgreSQL process (running as uid 999 inside the container, mapped through the user namespace) cannot write to the host directory. Rootless Podman also requires the host directory to be world-writable or owned by the mapped uid; `podman unshare chmod 777 ./db_data` sets this up on first run.

### 2.5 Database Transactions for All Balance Operations

Any operation that changes a bank account balance — creating/updating/deleting a transaction, creating/updating/deleting a transfer — runs inside a `BeginTx` … `Commit` block. This prevents split-brain states where, for example, a transfer deducts from the source account but a crash before the UPDATE of the target account leaves funds in limbo. The pattern throughout the handlers is:

```go
tx, err := h.db.BeginTx(ctx, nil)
// ... mutate rows
// ... update balance
if err := tx.Commit(); err != nil { tx.Rollback() }
```

Transfer creation additionally acquires a `SELECT … FOR UPDATE` row lock on the source account to prevent a race condition where two concurrent transfers both read the same balance and both succeed, overdrawing the account.

### 2.6 Go 1.22 ServeMux Pattern Matching

Routes are registered with the new method+path syntax:

```go
mux.HandleFunc("GET /api/bank-accounts/{id}/statement", protect(h.getBankAccountStatement))
mux.HandleFunc("GET /api/bank-accounts/{id}", protect(h.getBankAccount))
```

The Go 1.22 mux resolves conflicts by specificity: a longer path beats a shorter one, so `/api/bank-accounts/{id}/statement` is matched before `/api/bank-accounts/{id}`. Path values are extracted with `r.PathValue("id")`.

### 2.7 Angular Standalone Components and Lazy Loading

Each feature is a standalone Angular component declared with `standalone: true`. There are no `NgModule` declarations. Routes are lazy-loaded via `loadComponent`:

```typescript
{ path: 'dashboard', loadComponent: () => import('./features/dashboard/...').then(m => m.DashboardComponent), canActivate: [authGuard] }
```

This keeps the initial bundle small — only the login/register components are eagerly loaded; the rest are fetched on demand. Authentication uses a functional `authGuard` and a functional `authInterceptor` (Angular 17+ pattern, no class-based guards or interceptors needed).

### 2.8 SVG Donut Chart — No External Library

The tag statistics chart on the dashboard is rendered as an SVG `<circle>` using the `stroke-dasharray` technique, implemented entirely in `tag-chart.component.ts`. The approach:

1. Compute each tag's percentage of total spending.
2. Each segment is a full-radius `<circle>` with `stroke-dasharray: segmentLength gapLength` and `stroke-dashoffset` set to position the visible arc after all previous segments.
3. A `FALLBACK_COLORS` array provides colours for tags that have no custom colour set.

This avoids Chart.js, D3, or any other charting library, keeping the bundle size small.

---

## 3. Directory Structure

```
My Finances APP/
├── docker-compose.yml          # Three-service compose: db, api, ui
├── db_data/                    # Postgres persistent volume (host-mounted)
├── documentation.md            # This file
│
├── my_finances_api/            # Go REST API
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                 # Entry point
│   └── internal/
│       ├── auth/
│       │   └── jwt.go          # JWT + password hashing
│       ├── config/
│       │   └── config.go       # Env-var configuration
│       ├── database/
│       │   ├── database.go     # Connection pool setup
│       │   ├── migrations.go   # CREATE TABLE IF NOT EXISTS schema
│       │   └── seed.go         # Seeds system tags (Income, Expense) at startup
│       ├── handlers/
│       │   ├── router.go       # Route registration
│       │   ├── helpers.go      # writeJSON, writeError, pagination
│       │   ├── user_handler.go
│       │   ├── bank_account_handler.go
│       │   ├── tag_handler.go
│       │   ├── transaction_handler.go
│       │   ├── transfer_handler.go
│       │   ├── goal_handler.go
│       │   └── dashboard_handler.go
│       ├── middleware/
│       │   ├── auth.go         # JWT bearer token middleware
│       │   └── cors.go         # CORS headers
│       ├── models/
│       │   └── models.go       # All domain structs
│       └── scheduler/
│           └── scheduler.go    # Materialises repeatable transaction/transfer templates
│   └── tests/
│       ├── unit/
│       │   ├── jwt_test.go
│       │   ├── balance_test.go
│       │   └── repeatable_test.go
│       └── integration/
│           └── api_test.go
│
└── my_finances_ui/             # Angular 21 SPA
    ├── Dockerfile
    ├── nginx.conf              # SPA routing + API proxy
    ├── package.json
    ├── angular.json
    └── src/
        ├── main.ts
        ├── index.html
        ├── styles.css          # Global design system
        └── app/
            ├── app.ts
            ├── app.routes.ts
            ├── app.config.ts
            ├── core/
            │   ├── guards/auth.guard.ts
            │   ├── interceptors/auth.interceptor.ts
            │   ├── models/models.ts
            │   └── services/   # One service per domain entity
            ├── features/       # One folder per feature/route
            │   ├── auth/
            │   ├── dashboard/
            │   ├── bank-accounts/
            │   ├── bank-account-detail/
            │   ├── tags/
            │   ├── transactions/
            │   ├── transfers/
            │   └── goals/
            └── shared/
                └── components/
                    ├── nav/
                    └── tag-autocomplete/
```

---

## 4. Infrastructure & Deployment

### 4.1 docker-compose.yml

```yaml
services:
  my-finances-db:      # PostgreSQL with persistent volume and healthcheck
  my-finances-api:     # Go API, waits for db healthcheck before starting
  my-finances-ui:      # Angular SPA behind nginx, proxies /api/ to the API
```

The API container's `depends_on` uses `condition: service_healthy`, which means Compose waits until `pg_isready` returns success before starting the API. This avoids startup-order race conditions without any sleep-based retry logic in application code.

### 4.2 Go API Dockerfile

Multi-stage build to keep the final image minimal:

```dockerfile
FROM golang:1.25.8-alpine AS builder
# CGO_ENABLED=0 produces a fully static binary
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:latest
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
```

`CGO_ENABLED=0` is required because `lib/pq` uses only Go's `database/sql` interface with no C bindings — the binary is fully static and runs in a scratch-compatible image.

### 4.3 Angular UI Dockerfile

```dockerfile
FROM node:22-alpine AS builder
RUN npm ci && npm run build -- --configuration production

FROM nginx:alpine
COPY --from=builder /app/dist/my_finances_ui/browser /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
```

### 4.4 Nginx Configuration

```nginx
location / {
    try_files $uri $uri/ /index.html;   # SPA fallback for client-side routing
}

location /api/ {
    proxy_pass http://my-finances-api:8080/api/;   # Reverse proxy to Go API
}
```

The `/api/` proxy uses the Docker/Podman internal DNS name `my-finances-api`, which resolves because both containers are on the same compose network. This means the Angular app never needs to know the API's public hostname — all HTTP calls go to the same origin (the nginx container), and nginx forwards them.

### 4.5 Rootless Podman Notes

Running with rootless Podman requires two one-time setup steps on first launch:

```bash
# Fix SELinux label on volume directory
distrobox-host-exec podman unshare chmod 777 ./db_data

# Start all services
distrobox-host-exec podman-compose up -d
```

The `:Z` volume label in `docker-compose.yml` tells Podman to relabel the host directory's SELinux context to one that the container process can access. Without it, PostgreSQL fails with a permission denied error even after the chmod.

---

## 5. Backend — Go REST API

### 5.1 Entry Point (`main.go`)

The startup sequence is strictly ordered:

1. `config.Load()` — reads environment variables and applies defaults.
2. `database.Connect(cfg)` — opens the connection pool with `lib/pq`.
3. `database.RunMigrations(db)` — runs `CREATE TABLE IF NOT EXISTS` for all tables. Idempotent; safe to run on every startup.
4. `database.SeedSystemTags(db)` — inserts the Income and Expense system tags if they do not yet exist (`ON CONFLICT DO NOTHING`).
5. `scheduler.Run(db)` — starts the background goroutine for repeatable items.
6. `handlers.NewRouter(db, cfg)` — registers all routes.
7. `http.ListenAndServe(addr, router)` — blocks serving requests.

### 5.2 Configuration (`internal/config/config.go`)

All configuration is read from environment variables with sensible defaults for local development:

| Variable      | Default                            | Description                    |
|---------------|------------------------------------|--------------------------------|
| `DB_HOST`     | `localhost`                        | PostgreSQL hostname            |
| `DB_PORT`     | `5432`                             | PostgreSQL port                |
| `DB_NAME`     | `my_finances`                      | Database name                  |
| `DB_USER`     | `finances_user`                    | Database user                  |
| `DB_PASSWORD` | `finances_pass`                    | Database password              |
| `JWT_SECRET`  | `change-me-in-production-...`      | Signing key for JWT tokens     |
| `PORT`        | `8080`                             | HTTP listen port               |

### 5.3 Database Connection (`internal/database/database.go`)

The connection pool is tuned for a single-user personal app:

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

`lib/pq` constructs the DSN as `postgres://user:pass@host:port/dbname?sslmode=disable`. SSL is disabled because all communication is within a private container network.

### 5.4 Migrations (`internal/database/migrations.go`)

All `CREATE TABLE IF NOT EXISTS` statements are executed in a single `db.Exec(schema)` call. There is no migration versioning system because:

- All tables are created from scratch on every startup.
- `IF NOT EXISTS` makes the operation idempotent — re-running does nothing if the table exists.
- For a personal app that owns its own data, this is simpler than a migration runner.

The `pgcrypto` extension is enabled to provide `gen_random_uuid()` for UUID primary keys.

### 5.5 Auth Middleware (`internal/middleware/auth.go`)

The `Auth(secret)` function returns a middleware that wraps any `http.HandlerFunc`:

```go
func Auth(secret string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // Parse "Authorization: Bearer <token>"
            // Validate token, extract userID
            // Store userID in context
            // Call next(w, r)
        }
    }
}
```

The user ID is stored in the request context under a private `contextKey("userID")` type (not a bare string, to prevent collisions with other context values). Handlers retrieve it with `middleware.UserID(r)`.

### 5.6 CORS Middleware (`internal/middleware/cors.go`)

Added as the outermost handler wrapper in `NewRouter`. Sets:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

OPTIONS preflight requests are short-circuited with a 200 response before reaching any handler.

### 5.7 Handler Helpers (`internal/handlers/helpers.go`)

Three shared utilities used by every handler:

- `writeJSON(w, status, v)` — sets `Content-Type: application/json`, writes status, marshals `v`.
- `writeError(w, status, msg)` — writes `{"error": "msg"}`.
- `decodeJSON(r, v)` — decodes request body, returns error on malformed JSON.
- `parsePagination(r)` — reads `?page=` and `?limit=` query parameters with defaults (page 1, limit 20).
- `pages(total, limit)` — computes total page count for paginated responses.

### 5.8 Bank Account Handler (`internal/handlers/bank_account_handler.go`)

The most complex handler. Beyond standard CRUD, it provides:

**Statement endpoint** (`GET /api/bank-accounts/{id}/statement`):
Returns the account details, a paginated list of transactions for that account, and a list of upcoming items for the next 90 days.

**`buildUpcoming`**: Calculates future repeatable transaction and goal transfer events:
- Queries `transactions` where `is_repeatable=true` for this account, calls `nextOccurrences(day, now, horizon)` to get each monthly recurrence.
- Queries `goals` where this account is the source, distributes `target_value` over the goal's interval, and projects each future transfer date.

**`nextOccurrences(day, from, until)`**: Iterates month by month. For each month it clamps the requested day to the actual last day of the month (e.g., day 31 becomes day 28/29/30 as appropriate). Returns only dates strictly after `from` and not after `until`.

### 5.9 Transaction Handler (`internal/handlers/transaction_handler.go`)

`CREATE`:
1. Opens a DB transaction.
2. Inserts the transaction row.
3. If `is_repeatable=false`, updates `bank_accounts.balance` by `+value` (`add`) or `-value` (`subtract`). Repeatable templates are stored without touching the real balance.
4. Auto-tags the transaction with the "Income" system tag for `add` operations and the "Expense" system tag for `subtract` operations (deduped against any user-provided `tag_ids`).
5. Inserts tag associations into `transaction_tags`.
6. If `is_repeatable=true`, computes a projected monthly balance (sum of all repeatable adds/subtracts/transfers for the account) and returns an advisory `projected_balance_warning` string if the net is negative. This is informational — it does not block the request.
7. Commits.

`UPDATE`:
1. Opens a DB transaction.
2. Reverses the old transaction's balance effect.
3. Applies the new transaction's balance effect.
4. Replaces tag associations.
5. Commits.

`DELETE`:
1. Opens a DB transaction.
2. Reverses the transaction's balance effect.
3. Deletes the transaction row (cascades to `transaction_tags`).
4. Commits.

### 5.10 Transfer Handler (`internal/handlers/transfer_handler.go`)

`CREATE`:
1. Opens a DB transaction.
2. Locks the source account row with `SELECT … FOR UPDATE` — prevents concurrent overdraft.
3. If `is_repeatable=false`, checks `source.balance >= value`; returns 400 if insufficient. Repeatable transfer templates skip this check.
4. Inserts the transfer row with any provided `tag_ids`.
5. If `is_repeatable=false`, updates both account balances (`source - value`, `target + value`). Repeatable templates leave balances untouched.
6. If `is_repeatable=true`, computes a projected balance warning and returns it as `projected_balance_warning` in the response.
7. Commits.

`DELETE`: Reverses the balance changes atomically in the same way.

### 5.11 Dashboard Handler (`internal/handlers/dashboard_handler.go`)

Aggregates data across all the user's accounts:

- Total balance across all bank accounts.
- Latest 3 transactions per account (using a `ROW_NUMBER() OVER (PARTITION BY bank_account_id ORDER BY date DESC)` window function approach or a per-account subquery).
- Tag spending statistics: `SELECT tag_id, SUM(value)` grouped by tag for `subtract` transactions in the current month.
- Upcoming items in the next 30 days (from repeatable transactions and goals).

### 5.12 Goal Handler (`internal/handlers/goal_handler.go`)

Goals store a `source_account_id`, `target_account_id`, `start_date`, `end_date`, `interval_days`, and `target_value`. The API does not automatically execute goal transfers — it only records the goal definition. The `buildUpcoming` function in the bank account statement calculates when goal transfers are due and shows them as future items, making the goal visible on the account statement without any scheduled side effects.

### 5.13 Tag Handler (`internal/handlers/tag_handler.go`)

Tags support system-level and user-level entries. Key behaviours:

- **List**: returns both system tags (`is_system=true`, `user_id=NULL`) and the current user's own tags.
- **Create / Update**: rejects names that match reserved system tag names (`income`, `expense`, case-insensitive) with `400`.
- **Delete**: returns `403` when a system tag is targeted; users cannot delete the Income or Expense tags.

---

## 6. Frontend — Angular UI

### 6.1 Application Bootstrap

```typescript
// main.ts
bootstrapApplication(App, appConfig);
```

`appConfig` in `app.config.ts` provides:
- `provideRouter(routes)` — the lazy-loaded route tree.
- `provideHttpClient(withInterceptors([authInterceptor]))` — HTTP client with the auth interceptor registered.

### 6.2 Routing (`app.routes.ts`)

All protected routes use `canActivate: [authGuard]`. The guard reads the JWT from localStorage and redirects to `/login` if missing or expired. Routes:

| Path                       | Component                    | Protected |
|----------------------------|------------------------------|-----------|
| `/login`                   | LoginComponent               | No        |
| `/register`                | RegisterComponent            | No        |
| `/dashboard`               | DashboardComponent           | Yes       |
| `/bank-accounts`           | BankAccountsComponent        | Yes       |
| `/bank-accounts/:id`       | BankAccountDetailComponent   | Yes       |
| `/tags`                    | TagsComponent                | Yes       |
| `/transactions`            | TransactionsComponent        | Yes       |
| `/transfers`               | TransfersComponent           | Yes       |
| `/goals`                   | GoalsComponent               | Yes       |

### 6.3 Auth Interceptor (`core/interceptors/auth.interceptor.ts`)

A functional interceptor that reads `auth_token` from `localStorage` and injects it as a `Bearer` header on every outgoing request:

```typescript
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    req = req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return next(req);
};
```

### 6.4 Auth Service (`core/services/auth.service.ts`)

Manages the token lifecycle:
- `login(username, password)` — calls `POST /api/auth/login`, stores the returned JWT in `localStorage`.
- `register(...)` — calls `POST /api/auth/register`, stores the JWT.
- `logout()` — removes the token from `localStorage`, navigates to `/login`.
- `isLoggedIn()` — checks for token presence.

### 6.5 Domain Services

Each domain entity has a corresponding service in `core/services/`. All services inject `HttpClient` and construct URLs relative to `/api/`:

- `BankAccountService` — CRUD + `statement(id, page)` for the detail view.
- `TransactionService` — list with optional `bank_account_id` filter, create, delete.
- `TransferService` — paginated list, create, delete.
- `TagService` — list, create, update, delete.
- `GoalService` — CRUD.
- `DashboardService` — single `get()` call.

### 6.6 Transactions Component TypeScript Notes

Tag multi-select required careful typing. Angular's `FormBuilder` infers the array control type from its initial value:

```typescript
// Wrong: Angular infers type as `null`, not `string[]`
tag_ids: [[]]

// Correct: explicit cast tells Angular the element type
tag_ids: [[] as string[]]
```

Accessing the value also required going through `form.controls` rather than `form.value`:

```typescript
// Wrong: TypeScript refuses to cast `null` to `string[]`
(this.form.value.tag_ids as string[])

// Correct: controls-level access preserves type information
this.form.controls.tag_ids.value ?? []
```

### 6.7 Tag Chart Component (`features/dashboard/tag-chart/`)

The SVG donut chart uses a single `<circle>` element per tag segment. The key properties:

```
r = 40  →  circumference = 2π × 40 ≈ 251.2
stroke-dasharray = "segLen, (251.2 - segLen)"
stroke-dashoffset = -(sum of all previous segment lengths)
```

Each segment length is `(tag.percentage / 100) * 251.2`. The offset positions the start of each arc after all preceding arcs. The `viewBox="0 0 100 100"` coordinate system lets the radius of 40 sit inside a 10-unit margin for the stroke width.

### 6.8 Global Styles (`styles.css`)

The design system uses CSS custom properties defined on `:root`:

```css
--primary: #6366f1;       /* indigo */
--surface: #1e1e2e;       /* dark card background */
--background: #13131f;    /* page background */
--text: #e2e8f0;          /* primary text */
--text-muted: #94a3b8;    /* secondary text */
```

Utility classes: `.card`, `.btn`, `.btn-primary`, `.btn-danger`, `.form-group`, `.form-control`, `.table`, `.badge`, `.badge-add`, `.badge-subtract`, `.spinner`, `.pagination`.

### 6.9 Environment and API Base URL

`environment.ts` sets `apiUrl: ''` (empty string). All service calls use relative URLs like `/api/bank-accounts`. In development, Angular's dev server is not used (the app is served from nginx in Docker), so all requests go to the nginx container which proxies `/api/` to the Go API. In production the same nginx config applies.

---

## 7. Database Schema

### Entity Relationship Summary

```
users
  └── bank_accounts  (user_id FK)
  └── tags           (user_id FK, nullable — system tags have user_id=NULL)
  └── transactions   (user_id FK, bank_account_id FK)
        └── transaction_tags  (transaction_id, tag_id composite PK)
  └── transfers      (user_id FK, source_account_id FK, target_account_id FK)
        └── transfer_tags     (transfer_id, tag_id composite PK)
  └── goals          (user_id FK, source_account_id FK, target_account_id FK)
```

### Key Design Decisions

- **UUID primary keys** via `gen_random_uuid()` from `pgcrypto`. No sequential IDs exposed in URLs — prevents enumeration attacks.
- **`bank_accounts.icon_url` is `TEXT`** (changed from `VARCHAR(500)` in Milestone 2) — accommodates data URIs and long image URLs without truncation.
- **`NUMERIC(18,2)` for monetary values** — exact decimal arithmetic, no floating point rounding.
- **`TIMESTAMPTZ` for all timestamps** — stores in UTC, handles timezone-aware comparisons correctly.
- **`ON DELETE CASCADE`** — deleting a user removes all their data; deleting a bank account removes its transactions.
- **`operation CHECK ('add', 'subtract')`** — database-level enforcement of the two valid transaction types.
- **`tags.user_id` is nullable** — system tags (`is_system=true`) have `user_id=NULL`; user tags have a FK to `users`. A partial unique index `idx_tags_system_name ON tags(name) WHERE is_system=true` ensures system tag names stay unique across restarts.
- **`transactions.is_repeatable BOOLEAN`** — when `true`, the row is a template; the scheduler creates real (`is_repeatable=false`) copies from it. Statement and dashboard queries always filter `AND is_repeatable=false` to exclude templates from history views.
- **`transactions.repeatable_day INTEGER CHECK (1-31)`** — nullable; the day of month on which the scheduler materialises a template.
- **`last_executed_at TIMESTAMPTZ`** — nullable; the scheduler uses this to determine whether the repeatable template has been executed this month.
- **Junction tables** (`transaction_tags`, `transfer_tags`) use composite primary keys, which automatically prevent duplicate tag associations and create an index on both columns.

---

## 8. Security Design

### 8.1 User Isolation

Every query filters by `user_id = $N` derived from the JWT. A user cannot read, modify, or delete another user's data — even if they know the UUID of another resource. For example:

```sql
DELETE FROM bank_accounts WHERE id=$1 AND user_id=$2
```

If the account belongs to a different user, zero rows are affected and the handler returns 404 (not 403 — this avoids leaking the existence of the resource).

### 8.2 JWT Security

- Signature uses HMAC-SHA256 (constant-time comparison via `hmac.Equal`).
- Token expiry is checked on every request.
- The secret is configured via environment variable, never hardcoded.

### 8.3 SQL Injection Prevention

All database queries use parameterised statements (`$1`, `$2`, etc.) — never string interpolation. `database/sql` handles parameterisation via the prepared statement protocol.

### 8.4 CORS

The API allows all origins (`Access-Control-Allow-Origin: *`) because it is intended to be used with a single known frontend. In a multi-tenant deployment this should be restricted to the frontend origin.

### 8.5 Dark Theme (UI)

The UI supports a toggleable dark/light theme. Theme state is controlled via a `data-theme` attribute on the `:root` element. All color values are CSS custom properties scoped to `:root[data-theme="dark"]` and `:root[data-theme="light"]`, so a single attribute change cascades the full color swap with no JavaScript DOM traversal.

`ThemeService` (`core/services/theme.service.ts`) persists the selected theme to `localStorage` under the key `theme` so the preference survives page reloads. A color palette picker in the navigation bar lets the user switch between themes at runtime. On startup, the service reads `localStorage` and applies the stored theme before the first render to avoid a flash of unstyled content.

---

## 9. Repeatable Item Scheduler

The scheduler (`internal/scheduler/scheduler.go`) runs as a background goroutine started at application startup. It ticks every 24 hours and also fires once immediately on startup to catch any missed executions from a restart.

### Execution Logic

For each repeatable **transaction** template (`is_repeatable=true`):

1. Query rows where `repeatable_day <= today's day-of-month`.
2. Check if `last_executed_at` is in the current year and month. If yes, skip — already executed this month.
3. Open a DB transaction:
   - Insert a new real transaction row (`is_repeatable=false`) with the same fields.
   - Copy tag associations from the template via `copyTemplateTags()`.
   - Update `bank_accounts.balance` according to `operation`.
   - Set template's `last_executed_at = NOW()`.
4. Commit.

The same logic runs for repeatable **transfer** templates, adjusting both source and target account balances.

### Why `repeatable_day <= today` Instead of `= today`

If the application was offline on the scheduled day (e.g., the server was down on the 1st), it catches up the next time it runs. The `last_executed_at` check prevents double-execution within the same month.

### Month-End Handling

Day 31 templates will execute on the last day of shorter months (28/29/30) because `28 <= 31` is true before February ends.

---

## 10. Testing Strategy

### 10.1 Unit Tests

Located in `my_finances_api/tests/unit/`.

**`jwt_test.go`** — Table-driven tests covering:
- Token generation and validation round-trip.
- Expired token rejection.
- Wrong secret rejection.
- Malformed token (missing parts) rejection.
- Password hashing uniqueness (two calls to `HashPassword` with the same input produce different hashes due to random salt).
- Password verification correctness.

**`balance_test.go`** — Pure logic tests with no DB dependency:
- `transferAllowed(balance, amount)` — true when balance >= amount, false otherwise.
- `applyOperation(balance, value, op)` — correct addition and subtraction.

### 10.2 Integration Tests

Located in `my_finances_api/tests/integration/api_test.go`.

Uses `testcontainers-go` to spin up a real PostgreSQL container for each test run. The container is started with `postgres:latest`, migrations are run, and the full HTTP handler stack is tested via `httptest.NewServer`.

Tests cover:
- `POST /api/auth/register` — user creation returns a valid JWT.
- `POST /api/auth/login` — correct credentials return a token; wrong credentials return 401.
- `GET /api/bank-accounts` — requires auth; returns empty list for new user.
- `POST /api/bank-accounts` — creates an account, verifies the response.
- Transfer with insufficient balance — returns 400.

Helper functions:
- `registerAndGetToken(t, server)` — registers a test user, returns the JWT.
- `createBankAccount(t, server, token, name, balance)` — creates a bank account, returns the ID.

### 10.3 Running Tests

```bash
# Unit tests (no external dependencies)
cd my_finances_api
go test ./tests/unit/...

# Integration tests (requires Docker/Podman socket)
go test ./tests/integration/...

# All tests
go test ./...
```

---

## 11. API Reference

All endpoints (except `/api/auth/*`) require `Authorization: Bearer <token>`.

Paginated list responses return:

```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "limit": 20,
  "pages": 3
}
```

### Authentication

| Method | Path                   | Body                                     | Response            |
|--------|------------------------|------------------------------------------|---------------------|
| POST   | `/api/auth/register`   | `{name, username, email, password}`      | `{token, user}`     |
| POST   | `/api/auth/login`      | `{username, password}`                   | `{token, user}`     |

### Users

| Method | Path            | Body                              | Response  |
|--------|-----------------|-----------------------------------|-----------|
| GET    | `/api/users/me` | —                                 | User      |
| PUT    | `/api/users/me` | `{name, email}`                   | User      |

### Bank Accounts

| Method | Path                               | Body / Query         | Response                   |
|--------|------------------------------------|----------------------|----------------------------|
| GET    | `/api/bank-accounts`               | —                    | `BankAccount[]`            |
| POST   | `/api/bank-accounts`               | `{name, initial_balance, icon_url}` | BankAccount   |
| GET    | `/api/bank-accounts/{id}`          | —                    | BankAccount                |
| PUT    | `/api/bank-accounts/{id}`          | `{name, icon_url}`   | BankAccount                |
| DELETE | `/api/bank-accounts/{id}`          | —                    | 204                        |
| GET    | `/api/bank-accounts/{id}/statement`| `?page=&limit=`      | StatementResponse          |

`StatementResponse` contains `{account, transactions: PaginatedResponse<Transaction>, upcoming: UpcomingItem[]}`.

### Tags

| Method | Path            | Body              | Response  |
|--------|-----------------|-------------------|-----------|
| GET    | `/api/tags`     | —                 | `Tag[]` (system + user tags) |
| POST   | `/api/tags`     | `{name, color}`   | Tag — 400 if name is reserved (`income`/`expense`) |
| PUT    | `/api/tags/{id}`| `{name, color}`   | Tag — 400 if name is reserved |
| DELETE | `/api/tags/{id}`| —                 | 204 — 403 for system tags |

### Transactions

| Method | Path                      | Body / Query                                     | Response                        |
|--------|---------------------------|--------------------------------------------------|---------------------------------|
| GET    | `/api/transactions`       | `?bank_account_id=&page=&limit=`                 | `PaginatedResponse<Transaction>` (templates excluded) |
| POST   | `/api/transactions`       | `{name, value, operation, bank_account_id, date, tag_ids, is_repeatable?, repeatable_day?}` | `CreateTransactionResponse` |
| PUT    | `/api/transactions/{id}`  | same as POST                                     | Transaction                     |
| DELETE | `/api/transactions/{id}`  | —                                                | 204                             |

`operation` must be `"add"` or `"subtract"`. When `is_repeatable=true`, the response includes an optional `projected_balance_warning` string (advisory only).

### Transfers

| Method | Path                  | Body / Query                                             | Response                      |
|--------|-----------------------|----------------------------------------------------------|-------------------------------|
| GET    | `/api/transfers`      | `?page=&limit=`                                          | `PaginatedResponse<Transfer>` (templates excluded) |
| POST   | `/api/transfers`      | `{name, value, source_account_id, target_account_id, date, tag_ids, is_repeatable?, repeatable_day?}` | `CreateTransferResponse` |
| PUT    | `/api/transfers/{id}` | same as POST                                             | Transfer                      |
| DELETE | `/api/transfers/{id}` | —                                                        | 204                           |

Returns `400 {"error":"insufficient balance"}` if the source account balance is too low and `is_repeatable=false`. Repeatable transfer templates skip the balance check. When `is_repeatable=true`, the response includes an optional `projected_balance_warning` string.

### Goals

| Method | Path             | Body                                                           | Response   |
|--------|------------------|----------------------------------------------------------------|------------|
| GET    | `/api/goals`     | —                                                              | `Goal[]`   |
| POST   | `/api/goals`     | `{name, source_account_id, target_account_id, start_date, end_date, interval_days, target_value}` | Goal |
| PUT    | `/api/goals/{id}`| same as POST                                                   | Goal       |
| DELETE | `/api/goals/{id}`| —                                                              | 204        |

### Dashboard

| Method | Path             | Response                                              |
|--------|------------------|-------------------------------------------------------|
| GET    | `/api/dashboard` | `{accounts, total_balance, tag_stats, upcoming_expenses, recent_transactions}` |

---

## 12. Milestone 2 Changes

This section summarises the additions and fixes shipped in Milestone 2.

### 12.1 Bug Fix: Bank Account Image Upload

The `bank_accounts.icon_url` column was changed from `VARCHAR(500)` to `TEXT`. The previous limit caused failures when uploading base64-encoded image data URIs, which routinely exceed 500 characters. The update handler (`PUT /api/bank-accounts/{id}`) was also extended to support changing or removing the icon on an existing account — the `icon_url` field is now patched during updates, and sending an empty string clears the icon.

### 12.2 Bug Fix: Bank Account Initial Balance Validation

The create handler (`POST /api/bank-accounts`) now validates that `initial_balance >= 0`. Submitting a negative value returns `400 {"error": "initial_balance must be >= 0"}`. This prevents accounts from starting with an artificially negative balance, which would corrupt subsequent balance calculations.

### 12.3 Dark Theme and Color Palette Picker

A toggleable dark/light theme was added to the Angular UI. See Section 8.5 for the technical design. Key points:

- All colors are CSS variables on `:root` switched via a `data-theme` attribute.
- `ThemeService` reads and writes `localStorage` key `theme` for persistence.
- A color palette picker control in the navigation bar lets users toggle themes without navigating away from the current page.

### 12.4 Transfers in Bank Account Statement and Dashboard Summary

Transfers were previously excluded from the bank account statement and dashboard recent-transactions views. In Milestone 2, the statement and dashboard queries use a `UNION` approach that merges rows from both the `transactions` and `transfers` tables, ordered by date, so outgoing and incoming transfers appear inline alongside regular transactions. Each row carries a `type` discriminator (`"transaction"` or `"transfer"`) so the frontend can render the correct icon and label.

### 12.5 Makefiles

Both the API and UI directories now have a `Makefile` with the following targets:

| Target        | Description                                      |
|---------------|--------------------------------------------------|
| `lint`        | Run the language linter (golangci-lint / ESLint) |
| `trivy`       | Trivy filesystem/source scan for vulnerabilities |
| `trivy-image` | Trivy scan of the built container image          |
| `test`        | Run the test suite                               |
| `build`       | Build the binary / production bundle             |
| `all`         | Run lint, trivy, test, and build in sequence     |

Running `make all` from either project directory gives a full quality gate before deployment.

### 12.6 End-to-End Tests (Playwright)

A Playwright E2E test suite was added under `e2e/`. The tests run against the full compose stack (API + UI + database) and cover the critical user journeys:

- User registration and login.
- Creating, editing, and deleting a bank account.
- Creating transactions and verifying balance updates.
- Creating a transfer between two accounts.

Tests are executed with:

```bash
cd e2e
npx playwright test
```

The suite runs in headed or headless mode and targets the local nginx URL (`http://localhost`). CI configuration can point `BASE_URL` to any environment.

---

## 13. Milestone 3 Changes

### 13.1 System Tags (D16)

Two global system tags — **Income** (`#16a34a`) and **Expense** (`#dc2626`) — are seeded at startup via `database.SeedSystemTags`. They are stored in the `tags` table with `user_id=NULL` and `is_system=true`. Users:
- **Can see** system tags in the tags list, displayed in a separate read-only section.
- **Cannot** create, edit, or delete them.
- **Cannot** name their own tags `income` or `expense` (case-insensitive); the API returns `400` and the UI validates client-side.

Every new transaction is automatically associated with the Income or Expense system tag based on its `operation` field.

### 13.2 Repeatable Transactions (D17)

The `transactions` table gained three columns: `is_repeatable BOOLEAN`, `repeatable_day INTEGER`, and `last_executed_at TIMESTAMPTZ`. A repeatable transaction (`is_repeatable=true`) is a **template** — it does not affect the real account balance and does not appear in statement or dashboard history views. The scheduler materialises it into a real (`is_repeatable=false`) transaction on the chosen day each month.

The create endpoint returns a `projected_balance_warning` string when creating a repeatable template whose net monthly effect on the account would be negative. This is advisory only; the request still succeeds.

The UI form adds a "Repeatable" checkbox. When checked, a day-of-month picker and an income/expense type selector replace the operation dropdown.

### 13.3 Repeatable Transfers (D18)

The same `is_repeatable` / `repeatable_day` / `last_executed_at` columns were added to the `transfers` table. Repeatable transfer templates skip the source balance check and leave account balances untouched. The scheduler applies the balance changes when materialising the real record each month.

### 13.4 Tag Autocomplete (D19)

The tag multi-select was redesigned as a reusable `TagAutocompleteComponent` (`shared/components/tag-autocomplete/`). It accepts `@Input() tags`, `@Input() selectedIds`, and emits `@Output() selectionChange`. The user types to filter tags, clicks a suggestion from the dropdown to select it, and removes selections via chip ×-buttons. The dropdown closes 150 ms after blur to allow click events to register first.

### 13.5 Tags on Transfers (D20)

The transfer form and list were updated to support tags through `<app-tag-autocomplete>`. The `POST /api/transfers` and `PUT /api/transfers/{id}` bodies accept `tag_ids[]`, stored in `transfer_tags`. The transfers list table shows a Tags column with colour-coded chips.

### 13.6 Removal of Incomes and Expenses (D21)

The separate Income and Expense features were removed entirely:
- **API**: `income_handler.go` and `expense_handler.go` deleted; 8 routes removed from the router; `incomes` and `expenses` tables dropped in the alterations migration.
- **UI**: `IncomesComponent`, `ExpensesComponent`, `IncomeService`, `ExpenseService` deleted; routes removed from `app.routes.ts`; nav links removed.

Repeatable transactions replace the functionality these features provided.
