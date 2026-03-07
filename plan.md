# Auth Service Execution Plan

This document is a detailed, step-by-step execution plan for building a production-grade auth service in Go, with clear milestones and learning focus. It is designed to be worked through sequentially and verified at each phase.

## 0) Current State Summary

- Entry point: `cmd/api/main.go`.
- Config: `pkg/config`.
- Logger: `pkg/logger`.
- DB connection: `pkg/database`.
- Router: `internal/server`.
- Health endpoint: `internal/handler/health_handler.go`.
- Data model: `migrations/migrations.sql` includes users, roles, permissions, sessions, audit_logs.
- Repositories: `internal/repository/*`.
- DB is local Postgres via docker compose.

## 1) Project Goals

- Provide a real-world auth service with clean Go architecture.
- Cover modern auth flows: register, login, refresh, logout, RBAC.
- Practice idiomatic Go: clear packages, error handling, interfaces, tests.
- Keep minimal dependencies and visible execution paths.

## 2) Architecture & Packages

Keep a small, explicit layering:

- `cmd/api`: app wiring, config, server start/stop.
- `internal/server`: router + middleware wiring, http server.
- `internal/handler`: HTTP handlers (thin; only request parsing and response writing).
- `internal/service`: business logic (auth flows, token issuing, user/session management).
- `internal/repository`: DB access layer (already present).
- `internal/domain`: data types and invariants.
- `pkg/*`: reusable infrastructure (config, logger, database).

Add these packages:

- `internal/service/auth`: register, login, refresh, logout.
- `internal/service/user`: user profile and admin actions.
- `internal/token`: token hashing/verification, JWT signing.
- `internal/password`: password hashing and verification (bcrypt/argon2).
- `internal/httpx`: request/response helpers (decode/encode, error payloads).

## 3) API Design (v1)

Base path: `/v1`

### Auth (public)

- `POST /v1/auth/register`
  - Request: `{ "email": "", "password": "" }`
  - Response: `user + access_token + refresh_token`

- `POST /v1/auth/login`
  - Request: `{ "email": "", "password": "" }`
  - Response: `user + access_token + refresh_token`

- `POST /v1/auth/refresh`
  - Request: `{ "refresh_token": "" }`
  - Response: `access_token + refresh_token`

- `POST /v1/auth/logout`
  - Request: `{ "refresh_token": "" }` or `session_id`
  - Response: `{ "revoked": true }`

### User (protected)

- `GET /v1/me`
  - Response: user info

### RBAC (admin)

- `POST /v1/roles`
- `POST /v1/permissions`
- `POST /v1/roles/:id/permissions/:permId`
- `POST /v1/users/:id/roles/:roleId`
- `GET /v1/users/:id/permissions`

### Audit Logs (admin)

- `GET /v1/audit?user_id=...`
- `GET /v1/audit?action=...`

## 4) Core Flows

### Register

1) Validate email/password.
2) Hash password.
3) Create user.
4) Create session with refresh token hash.
5) Issue access token (JWT).
6) Return user + tokens.

### Login

1) Fetch user by email.
2) Verify password hash.
3) Create session.
4) Issue access/refresh tokens.
5) Return user + tokens.

### Refresh

1) Validate refresh token.
2) Find session by token hash.
3) Check revoked/expiry.
4) Rotate session (new token hash).
5) Issue new access token.
6) Return tokens.

### Logout

1) Validate refresh token.
2) Revoke session.
3) Return success.

## 5) Data & Security Decisions

- Password hashing: `bcrypt` or `argon2id` (prefer argon2id).
- Refresh token: random, store hash in DB.
- Access token: JWT with short TTL (e.g., 10-15m).
- Refresh token TTL: 7-30 days.
- Session revocation: `revoked` flag and `expires_at`.
- Token signing key: env `JWT_SECRET` or `JWT_PRIVATE_KEY`.
- Session binding: store `ip_address` and `user_agent`.

## 6) Config & Env

Add required env vars:

- `JWT_SECRET` (or `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` if using asymmetric).
- `ACCESS_TOKEN_TTL` (e.g., `15m`).
- `REFRESH_TOKEN_TTL` (e.g., `30d`).
- `PASSWORD_HASH_COST` (bcrypt) or `ARGON2_*` (argon2id params).

Extend `pkg/config` to load and validate these.

## 7) Error Handling & Responses

Adopt consistent error responses:

- `400` for validation errors.
- `401` for auth failures.
- `403` for permission failures.
- `404` for not found.
- `409` for conflicts (duplicate email).
- `500` for unexpected errors.

Define a common response shape:

- `{"error": {"code": "", "message": ""}}`

## 8) Repository Usage

Use repository errors consistently:

- `ErrNotFound` -> `404`.
- `ErrInvalidArgument` -> `400`.
- raw DB errors -> `500` unless mapped (e.g., unique violation -> `409`).

## 9) Testing Strategy

### Unit Tests

- Password hashing/verification.
- Token generation/validation.
- Config validation.

### Integration Tests

- Register -> login -> refresh -> logout.
- RBAC role assignment and permission checks.
- Audit log inserts.

### Test DB

- Use docker compose to run Postgres for integration tests.
- Clear tables between tests.

## 10) Observability

- Structured logs for each request.
- Include `request_id` middleware.
- Audit log entries for auth flows.
- Health endpoint already present.

## 11) Execution Steps (Detailed)

### Step 1: Add new packages

- `internal/password`: hash + verify.
- `internal/token`: JWT signing + refresh token generation + hash.
- `internal/service/auth`: register/login/refresh/logout.
- `internal/httpx`: JSON helpers and error rendering.

### Step 2: Add config for auth

- Extend `pkg/config` with token TTLs and secrets.
- Validate secrets at startup.

### Step 3: Implement Auth service

- Define service interfaces: `UserRepo`, `SessionRepo`, `AuditRepo`.
- Implement methods using repositories.
- Handle unique email violation.

### Step 4: Implement handlers

- `internal/handler/auth_handler.go`: register/login/refresh/logout.
- Parse JSON, call service, return JSON.
- No DB logic in handlers.

### Step 5: Add middleware

- JWT auth middleware (validate access token, attach user context).
- Optional request-id middleware.

### Step 6: Wire routes

- Update router to mount `/v1/auth/*`.
- Add `/v1/me` protected route.

### Step 7: RBAC endpoints

- Add role/permission handlers in `internal/handler/rbac_handler.go`.
- Use repository methods.
- Add permission check middleware for admin routes.

### Step 8: Audit logging

- Log register/login/logout/refresh.
- Log RBAC changes.

### Step 9: Tests

- Add unit tests in `internal/password`, `internal/token`, `internal/service`.
- Add integration tests under `internal/handler` or `tests/`.

### Step 10: Tooling

- Add `Makefile` with targets: `test`, `lint`, `run`, `migrate`.
- Add `golangci-lint` config if desired.

## 12) Acceptance Checklist

- Register/login/refresh/logout work.
- Access token protects `/v1/me`.
- Sessions can be revoked.
- RBAC endpoints work.
- Audit logs recorded.
- Tests pass.
- Logs are structured.

## 13) Learning Focus Areas

- Proper Go package design.
- Error handling idioms.
- Context propagation.
- Table-driven tests.
- Database transaction boundaries.
- Security practices in auth.

## 14) Suggested Timeline

- Day 1: Auth service + handlers + tests.
- Day 2: Middleware + protected routes.
- Day 3: RBAC + audit logs.
- Day 4: Hardening, rate limits, docs.

## 15) Optional Enhancements

- Email verification flow.
- Password reset flow.
- Login rate limiting with IP-based buckets.
- Account lockout on repeated failures.
- Session management UI API.
