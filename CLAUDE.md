# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Razad is an open-source (AGPLv3) server management and deployment platform for solo developers and small teams deploying on Linux VPS. It combines native process management, automated app deployment, database provisioning, domain/SSL management, real-time logs, and a safety-gated AI layer.

**Tagline:** *Your server, guided.*

The project is in active development following the architecture defined in `docs/`.

## Commands

**Build daemon binary:**
```bash
make build
```

**Run locally (dev mode — frontend proxy expected on :5173):**
```bash
make dev
```

**Run all tests (use module prefix to avoid node_modules confusion):**
```bash
go test github.com/razad/razad/... -count=1 -short
```
Or via Makefile:
```bash
make test
```

**Frontend setup and build:**
```bash
make web-setup   # npm install in web/
make web-build   # npm run build (populates web/build/ for Go embed)
make web-dev     # start Vite dev server on :5173
```

**Full release build (frontend + daemon):**
```bash
make release
```

**Lint and format:**
```bash
make lint        # golangci-lint
make fmt         # go fmt + go vet
```

## Go Module

`github.com/razad/razad` — Go 1.22, CGO_ENABLED=1 required for SQLite support.

## Tech Stack

- **Backend:** Go (single-binary daemon, systemd-managed)
- **Frontend:** SvelteKit (embedded static assets in the Go binary)
- **Process Supervisor:** systemd (native, not Docker)
- **Reverse Proxy:** Nginx (config generated and validated by Razad)
- **Primary DB:** PostgreSQL (SQLite optional for lightweight self-hosted)
- **Messaging/Cache:** Redis, NATS
- **Deployment target:** Linux (single binary, no Docker dependency)

## Architecture: Two-Plane Design

Razad separates into two planes — this boundary is the primary scalability mechanism:

- **Execution Plane** — local server: process management, config application, app runtime, logging. The node is the authority for local system changes.
- **Control Plane** — cloud (optional): identity, dashboard, orchestration, billing, fleet visibility, AI coordination. Cloud services request actions; the node enforces them.

Three operating modes share this architecture:
1. **Self-hosted** — single binary on customer-owned server, no cloud dependency
2. **BYO VPS Cloud** — customer owns VPS, Razad Cloud manages via node agent
3. **Managed Infrastructure** — Razad provisions and operates infrastructure

## Internal Module Map (`internal/`)

Each directory under `internal/` corresponds to a domain module. Modules follow a strict service-repository-handler pattern:

| Module | Responsibility |
|---|---|
| `config` | Typed runtime config, env loading, validation |
| `crypto` | Encryption at rest, secrets protection |
| `auth` | Login/logout, sessions, API tokens |
| `org` | Tenancy boundaries, projects, membership |
| `server` | Node identity, enrollment, liveness/heartbeat |
| `app` | App CRUD, deployment lifecycle, state transitions |
| `runtime` | Detect app runtime from repo structure |
| `process` | systemd unit generation, service control |
| `proxy` | Domain binding, Nginx config generation |
| `ssl` | Certificate issuance/renewal via certbot |
| `database` | PostgreSQL provisioning, credentials, backups |
| `observability` | Log streaming (WebSocket), health snapshots |
| `agent` | AI orchestration, safety filtering, action gating |
| `audit` | Immutable audit trail for all privileged actions |
| `eventbus` | Internal event pub/sub |
| `policy` | Validation and authorization gate (allow/deny) |
| `notification` | Alerts for node offline, SSL expiry, failures |
| `job` | Background scheduler, recurring tasks |
| `websocket` | Real-time log and event streams |
| `api` | HTTP handlers (thin — delegate to services) |
| `install` | Bootstrap, OS checks, idempotent installer |
| `deployment` | Deployment workflow orchestration |
| `storage` | File/artifact storage paths |
| `proxy` | Reverse proxy config management |

## Critical Architecture Rules

These constraints are enforced by design, not convention:

1. **No module mutates privileged system state without passing through `policy` (validation engine).**
2. **Every state-changing action must emit an audit event via `audit`.**
3. **Secrets are handled only through `crypto`/`config` — never ad-hoc.**
4. **AI may only execute explicitly whitelisted, non-destructive actions.** The action registry gates all AI execution. Destructive operations (delete app, drop DB, modify firewall) are blocked by design.
5. **Nginx/systemd config must be validated before apply** — keep last-known-good config as rollback.
6. **Handlers are thin** — business logic lives in services, persistence in repositories.

## Directory Structure

```
cmd/            # Go main entry points (daemon, CLI tools)
internal/       # All domain modules (see map above)
web/            # SvelteKit frontend
  src/lib/      # Shared components, API client, stores, WebSocket client
  src/routes/   # SvelteKit page routes
  static/       # Static assets
configs/        # Reference configs (nginx templates, systemd unit templates)
deployments/    # Deployment manifests (nginx, systemd, templates)
migrations/     # Database migration files (versioned)
scripts/        # Build, install, and operational scripts
tests/          # Test suites (unit, integration, e2e)
docs/           # All design and architecture documents
```

## Git Workflow

- **Branch model:** Trunk-based with short-lived feature branches off `main`
- **Branch naming:** `feature/<name>`, `fix/<name>`, `chore/<name>`, `hotfix/<name>`, `refactor/<name>`
- **Commit convention:** `type(scope): summary` (types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `ci`, `perf`, `sec`)
- **No direct commits to main** — everything via reviewed PR
- **Semantic versioning:** `v0.x` during development, `v1.0.0` at MVP

## Go-Specific Conventions

- Idiomatic Go formatting; short, meaningful package names
- No cyclic imports between modules
- Return typed errors with context (wrap, don't swallow)
- Internal errors converted to stable public error responses
- Avoid premature optimization — profile first

## Frontend Conventions

- Keep API calls isolated in a client layer (`web/src/lib/api/`)
- Components for repeated UI structures; screens operationally focused
- WebSocket streams for real-time logs
- Optimistic UI only where safe; refresh authoritative state after mutations
- Surface errors clearly; no raw internal errors in UI

## API Conventions

- Versioned routes (`/api/v1/...`)
- JSON request/response; stable error envelopes
- Destructive actions must be explicit (never implied by a different operation)
- Authentication required for all privileged endpoints
- Tenant-scoped: every sensitive request scoped to an org or installation

## Testing Strategy

- **Unit tests:** service logic, validation, policy rules, config parsing, error conversion
- **Integration tests:** DB interaction, API endpoints, migration correctness, config generation, AI action gating
- **E2E tests:** install → login → deploy → domain → SSL → logs → audit flow
- **Security tests (mandatory):** secrets never returned plaintext, logs redacted, AI allowlist enforced, cross-tenant access denied, auth required on privileged endpoints
- **Negative tests required** for: unauthorized access, invalid input, malformed config, forbidden AI actions, token replay, broken templates
- **Regression tests required** for every production bug fix

## Implementation Order (from design docs)

The recommended build order prioritizes the core value path:
1. Config and Secrets → 2. Auth and Sessions → 3. Org/Projects → 4. Installer → 5. API/UI shell → 6. Server/Node Identity → 7. App Management → 8. Runtime Detection → 9. Process Management → 10. Proxy/Domain → 11. SSL → 12. Logs/Observability → 13. Audit → 14. Validation/Policy → 15. Database Management → 16. AI Orchestration → 17. Background Jobs → 18. Notifications → 19-20. Cloud/Billing hooks

## Key Design Documents

All in `docs/`:
- `razad_prd_v_1_0.md` — Product requirements and scope
- `razad_sad_v_1_0.md` — System architecture
- `razad_system_design_per_module_v1_0.md` — Per-module specs, dependencies, failure modes
- `razad_api_specification_v_1_0.md` — Full API contract
- `razad_security_architecture_v_1_0.md` — Threat model, trust zones, AI safety
- `razad_erd_v_1_0.md` — Entity-relationship diagram
- `01_architecture_decisions_adr_v1_0.md` — Recorded architecture decisions with rationale
- `04_development_standards_v1_0.md` — Coding and review standards
- `06_testing_strategy_v1_0.md` — Test pyramid, critical scenarios, release criteria
