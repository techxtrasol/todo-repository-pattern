# Todo App

A small, layered REST API for managing todos, built in Go with PostgreSQL.

## Overview

The service exposes CRUD endpoints for todos (title, description, completed) plus a health check. It is structured with a clean, layered architecture (handler → service → repository) and uses sqlc for type-safe database access.

## Tech Stack

| Tool / Library | Why it's used |
| --- | --- |
| **Go 1.27** | Fast, statically typed, great standard library for HTTP services. |
| **chi (go-chi/chi/v5)** | Lightweight, idiomatic HTTP router with middleware support. |
| **pgx (jackc/pgx/v5)** | High-performance PostgreSQL driver and connection pool. |
| **sqlc** | Generates type-safe Go code from SQL queries; keeps SQL as the source of truth. |
| **PostgreSQL 16** | Reliable relational database with UUID and timestamp support. |
| **godotenv** | Loads local `.env` configuration during development. |
| **google/uuid** | UUID generation/parsing for todo IDs. |
| **Redoc / Swagger UI** | Interactive API documentation served from `/docs` and `/redoc`. |
| **Docker Compose** | Runs a local PostgreSQL instance. |
| **Make** | Common development tasks (`build`, `run`, `test`, etc.). |

## Architecture / Pattern

The codebase follows a **layered (clean) architecture**:

```
cmd/api            -> application entrypoint and wiring (DI)
internal/config    -> environment configuration
internal/handler   -> HTTP handlers, routing, request/response DTOs
internal/service   -> business logic
internal/repository-> data access (behind an interface)
internal/db        -> sqlc-generated, type-safe DB code
internal/domain    -> domain models
db/migrations      -> SQL schema migrations
db/queries         -> SQL queries consumed by sqlc
```

Key patterns:
- **Dependency injection**: `main.go` constructs repo → service → handler → router.
- **Repository interface**: `TodoRepository` abstracts data access, making the service testable.
- **sqlc code generation**: SQL files are the source of truth; generated Go code in `internal/db` should not be edited by hand.
- **Partial updates**: `PUT` uses nullable parameters with `COALESCE`, so omitted fields keep their current value.

## Repository Pattern

The data access layer is hidden behind an interface (`TodoRepository` in `internal/repository/todo_repository.go`). The service depends only on that interface, not on pgx/SQL:

```go
type TodoRepository interface {
    Create(ctx context.Context, title string, description *string) (db.Todo, error)
    GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error)
    List(ctx context.Context) ([]db.Todo, error)
    Update(ctx context.Context, id uuid.UUID, title, description *string, completed *bool) (db.Todo, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

A concrete `todoRepository` implements it using sqlc-generated code. This makes the service testable (inject a fake) and keeps SQL details out of business logic.

## Getting Started

### Prerequisites
- Go 1.27+
- Docker (for PostgreSQL) — or a running Postgres instance
- `sqlc` (optional, for regenerating DB code)

### 1. Start the database
```bash
make up
```
Postgres runs on `127.0.0.1:5432` with credentials from `docker-compose.yml`.

### 2. Configure environment
`.env` is already provided:
```env
DATABASE_URL=postgres://todo:todo123@localhost:5432/todo_db?sslmode=disable
APP_PORT=8080
```

### 3. Run migrations
Apply `db/migrations/001_create_todos.sql` to your database (e.g. with `psql` or goose).

### 4. Run the app
```bash
make run
```
Server starts on `http://localhost:8080`.

### Make targets
- `make build` — build all packages
- `make run` — run the API
- `make test` — run tests
- `make tidy` — tidy Go modules
- `make sqlc` — regenerate DB code
- `make up` / `make down` — start/stop Postgres

## API Endpoints

Base URL: `http://localhost:8080`

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Health check |
| GET | `/api/v1/todos` | List all todos |
| POST | `/api/v1/todos` | Create a todo |
| GET | `/api/v1/todos/{id}` | Get a todo by id |
| PUT | `/api/v1/todos/{id}` | Update a todo (partial) |
| DELETE | `/api/v1/todos/{id}` | Delete a todo |

### Create todo
`POST /api/v1/todos`
```json
{
  "title": "Buy milk",
  "description": "2% fat"
}
```
Response `201 Created`:
```json
{
  "id": "uuid",
  "title": "Buy milk",
  "description": "2% fat",
  "completed": false,
  "created_at": "...",
  "updated_at": "..."
}
```

### Update todo (partial)
`PUT /api/v1/todos/{id}`
```json
{
  "completed": true
}
```
Only provided fields are updated.

### Documentation
- Swagger UI (interactive): `http://localhost:8080/docs`
- Redoc: `http://localhost:8080/redoc`
- OpenAPI spec: `http://localhost:8080/openapi.yaml`

## Authentication

All `/api/v1/todos` endpoints require a JWT access token:

```
Authorization: Bearer <access_token>
```

### Register
`POST /api/v1/auth/register`
```json
{ "email": "user@example.com", "password": "secret" }
```

### Login
`POST /api/v1/auth/login` → returns `access_token` and `refresh_token`.

### Refresh
`POST /api/v1/auth/refresh` with `{ "refresh_token": "..." }`.

### Logout
`POST /api/v1/auth/logout` with `{ "refresh_token": "..." }`.

### Current user
`GET /api/v1/auth/me` (requires access token).

## Notes
- `.env` is gitignored; never commit real credentials.
- Postgres is bound to `127.0.0.1` only to avoid exposing it to the network.
EOF
echo done