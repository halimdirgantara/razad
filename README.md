# Razad

> *Your server, guided.* — Server kamu, jalur yang benar.

**Razad** is an open-source server management and deployment platform for Linux VPS and dedicated servers. The current repository focuses on the self-hosted/local control core: authentication, organizations, projects, app management, encrypted env vars, local process management, health checks, and real-time logs.

Some architecture sections in `docs/` describe the broader product vision. Those cloud, AI, proxy, SSL, and managed-infrastructure capabilities are roadmap items unless the code in this repository explicitly implements them.

No Docker required. No Kubernetes. Just a single Go binary, Linux services, and an embedded SvelteKit frontend.

---

## Features

- **Auth & Tenancy** — Login, sessions, organizations, and membership-based access control.
- **App Management** — Create apps from Git projects, manage runtime metadata, deploy, stop, restart, and delete.
- **Encrypted Environment Variables** — App env vars are encrypted at rest.
- **Process Management** — Local process runner for development/testing; systemd-backed production runner is being completed in code.
- **Health & Logs** — Node health snapshots and WebSocket log streaming.
- **Embedded UI** — SvelteKit frontend built into the Go binary.

---

## Architecture

### Current scope

Razad’s current codebase is a local-first control core:

- Go backend daemon
- SQLite/PostgreSQL-compatible persistence layer
- Auth and organization membership
- Project-scoped app management
- Local process supervision
- Health snapshot collection
- WebSocket log streaming

### Roadmap scope

The product docs also describe a larger cloud vision:

- control plane / execution plane split
- managed infrastructure
- AI orchestration
- proxy / SSL automation
- database provisioning
- fleet visibility / billing

Those are not all implemented in this repository yet.

---

## Quick Start (Development)

### Prerequisites

- Go 1.22+
- Node.js 20+
- CGO_ENABLED=1 (required for SQLite support)

### Setup

```bash
# Clone the repository
git clone https://github.com/razad/razad.git
cd razad

# Install Go dependencies
go mod tidy

# Install frontend dependencies
make web-setup

# Build frontend assets
make web-build

# Build the daemon binary
make build

# Run tests
make test
```

### Development Mode

```bash
# Terminal 1: Start the backend
make dev

# Terminal 2: Start the frontend with hot-reload
make web-dev
```

The Vite dev server proxies `/api` requests to the Go backend on `:8080`.

### Full Release Build

```bash
make release
```
