# Codeflow

A backend service for **secure, asynchronous remote code execution**. Codeflow accepts source code over a REST API, persists each submission, and dispatches it onto a queue for sandboxed execution by background workers — the architecture behind online judges, coding playgrounds, and interview platforms.

Built in Go with a clean, domain-driven layout (domain → repository → service → handler), PostgreSQL for durable state, and Redis as the execution queue.

> **Project status:** Active development. The API, authentication, persistence, and queue dispatch are working end to end. The execution worker and sandbox runtime are scaffolded but not yet implemented — see the [Roadmap](#roadmap). Submitted jobs are validated, stored, and enqueued, but remain in `pending` until a worker is built to consume them.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Layout](#project-layout)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Data Model](#data-model)
- [Roadmap](#roadmap)

---

## Features

- **JWT authentication** — register, login, token refresh, and logout, with bcrypt-hashed passwords and rotating refresh tokens.
- **Asynchronous execution pipeline** — code submissions are validated, persisted, and pushed onto a Redis queue for out-of-band processing.
- **Multi-language support (planned runtime)** — submissions accept Python, Go, and JavaScript, with a 10 KB source-size limit enforced at the domain layer.
- **Durable job tracking** — every execution moves through an explicit state machine (`pending → queued → running → completed | failed | timeout`) backed by PostgreSQL.
- **Production-grade plumbing** — structured `slog` logging, request tracing via `X-Trace-ID`, panic recovery middleware, connection pooling, and a built-in SQL migration runner.

## Architecture

```
                 ┌──────────────┐
   HTTP client → │   API server │  (cmd/api)
                 │              │
                 │  ┌────────┐  │   validate + persist
                 │  │  auth  │  │ ───────────────► PostgreSQL
                 │  │  exec  │  │ ◄───────────────
                 │  └────────┘  │
                 └──────┬───────┘
                        │ LPUSH executions:queue
                        ▼
                 ┌──────────────┐
                 │    Redis     │  (execution queue)
                 └──────┬───────┘
                        │ BRPOP  (planned)
                        ▼
                 ┌──────────────┐
                 │    Worker    │  (cmd/worker — not yet implemented)
                 │  + sandbox   │
                 └──────────────┘
```

Each domain (`auth`, `execution`) follows the same layered pattern:

- **`domain.go`** — entities, value objects, and business rules (no I/O).
- **`repository.go`** — storage interface defined where it is consumed.
- **`postgres/`** — concrete `pgx`-backed repository implementation.
- **`service.go`** — use-case orchestration and business logic.
- **`handler.go`** — HTTP transport: decode, call service, encode response.

Dependencies are wired explicitly via constructors in `cmd/api/main.go` — no global state, no DI framework.

## Tech Stack

| Concern         | Choice                                      |
| --------------- | ------------------------------------------- |
| Language        | Go 1.25                                     |
| HTTP routing    | `net/http` (standard library `ServeMux`)    |
| Database        | PostgreSQL 16 via `pgx/v5` (pooled)         |
| Queue           | Redis 7 via `go-redis/v9`                    |
| Auth            | `golang-jwt/v5` + `golang.org/x/crypto` bcrypt |
| Config          | Environment variables via `godotenv`        |
| Logging         | `log/slog` (JSON in prod, text in dev)      |

## Project Layout

```
codeflow/
├── cmd/
│   ├── api/            # HTTP API entrypoint (implemented)
│   ├── worker/         # Queue consumer (planned)
│   └── gateway/        # API gateway (planned)
├── internal/
│   ├── auth/           # Authentication domain
│   ├── execution/      # Code-execution domain
│   └── platform/       # Cross-cutting infrastructure
│       ├── config/     #   env-driven configuration
│       ├── logger/     #   structured slog setup
│       ├── middleware/ #   trace ID, request logging, recovery
│       └── migrator/   #   SQL migration runner
├── migrations/         # Versioned .sql schema migrations
├── docker-compose.yml  # Local PostgreSQL + Redis
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (for PostgreSQL and Redis)

### 1. Start the dependencies

```bash
docker compose up -d
```

This brings up PostgreSQL 16 (`localhost:5432`) and Redis 7 (`localhost:6379`) with health checks and a persistent volume for Postgres data.

### 2. Configure the environment

Create a `.env` file in the project root (see [Configuration](#configuration)):

```env
PORT=8080
DATABASE_URL=postgres://codeflow:codeflow@localhost:5432/codeflow
REDIS_URL=localhost:6379
JWT_SECRET=replace-with-a-long-random-secret
ENV=development
```

### 3. Run the API

```bash
go run ./cmd/api
```

Migrations run automatically on startup. The server listens on `:8080` (or `PORT`).

### 4. Verify

```bash
curl http://localhost:8080/health
```

## Configuration

All configuration is read from environment variables (loaded from `.env` in development).

| Variable       | Required | Default                  | Description                                |
| -------------- | -------- | ------------------------ | ------------------------------------------ |
| `DATABASE_URL` | Yes      | —                        | PostgreSQL connection string               |
| `JWT_SECRET`   | Yes      | —                        | Secret used to sign access tokens          |
| `PORT`         | No       | `8080`                   | HTTP listen port                           |
| `REDIS_URL`    | No       | `localhost:6379`         | Redis address for the execution queue      |
| `ENV`          | No       | `development`            | `production` switches logs to JSON output  |

> **Security note:** never commit a real `JWT_SECRET` or database credentials. The `.env` in this repo is for local development only — rotate any secret that has been shared.

## API Reference

Base URL: `http://localhost:8080`

### Health

| Method | Path       | Description       |
| ------ | ---------- | ----------------- |
| `GET`  | `/health`  | Liveness check    |

### Authentication

| Method | Path                     | Body                  | Description                                 |
| ------ | ------------------------ | --------------------- | ------------------------------------------- |
| `POST` | `/api/v1/auth/register`  | `{ email, password }` | Create a new user                           |
| `POST` | `/api/v1/auth/login`     | `{ email, password }` | Return an access token (15 min) + refresh token (7 days) |
| `POST` | `/api/v1/auth/refresh`   | `{ token }`           | Exchange a refresh token for a new pair     |
| `POST` | `/api/v1/auth/logout`    | `{ token }`           | Invalidate a refresh token                  |

**Example — register:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"s3cret-pass"}'
```

### Executions

| Method | Path                       | Body                | Description                              |
| ------ | -------------------------- | ------------------- | ---------------------------------------- |
| `POST` | `/api/v1/executions`       | `{ language, code }`| Submit code; persists and enqueues it    |
| `GET`  | `/api/v1/executions/{id}`  | —                   | Fetch a single execution (owner-scoped)  |
| `GET`  | `/api/v1/executions`       | —                   | List the caller's executions (newest first) |

**Example — submit code:**

```bash
curl -X POST http://localhost:8080/api/v1/executions \
  -H 'Content-Type: application/json' \
  -d '{"language":"python","code":"print(\"hello\")"}'
```

Supported languages: `python`, `go`, `javascript`. Source is capped at 10 KB.

> **Note:** request-level JWT enforcement on execution endpoints is not yet wired in — the user is currently resolved from a placeholder. See the Roadmap.

## Data Model

| Table             | Purpose                | Key columns                                                                 |
| ----------------- | ---------------------- | -------------------------------------------------------------------------- |
| `users`           | Accounts               | `id`, `email` (unique), `password_hash`, timestamps                        |
| `refresh_tokens`  | Refresh-token store    | `id`, `user_id` (FK), `token` (unique), `expires_at`                       |
| `executions`      | Submitted jobs         | `id`, `user_id` (FK), `language`, `code`, `status`, `output`, `error`, `exit_code`, `duration_ms`, timestamps |

Foreign keys cascade on user deletion. Indexes exist on `email`, `token`, `user_id`, `status`, and `created_at`.

Schema is managed by versioned files in `migrations/`, applied once each by the built-in migration runner (tracked in a `migration_files` table inside a transaction).

## Roadmap

The HTTP API, authentication, persistence, and queue dispatch are functional. The execution backend is the next major milestone:

- [ ] **Execution worker** (`cmd/worker`) — consume `executions:queue` and drive the status state machine.
- [ ] **Sandbox runtime** (`internal/sandbox`) — isolated, resource-limited execution (e.g. containers) for each language.
- [ ] **Result streaming** (`internal/streaming`) — real-time output back to clients.
- [ ] **JWT middleware** — enforce authentication and resolve the user on execution endpoints.
- [ ] **Rate limiting** (`internal/ratelimit`) — per-user submission throttling.
- [ ] **Test suite** — unit, integration, and end-to-end coverage.
- [ ] **App Dockerfile** — containerize the API and worker for deployment.

---

*Codeflow is a personal project exploring production-grade backend architecture in Go.*
