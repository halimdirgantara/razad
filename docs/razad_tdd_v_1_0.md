# Razad
## Technical Design Document (TDD) v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Scope:** Self-hosted OSS core, Razad Cloud BYO VPS, and Razad Managed Infrastructure readiness  
**Primary Stack:** Go backend, SvelteKit frontend, systemd-native execution  
**License Assumption:** AGPLv3

---

## 1. Purpose

This document defines the technical architecture of Razad v1.0 and the design constraints required to support future growth into a cloud control plane and managed infrastructure platform.

The main objective is to keep the execution layer lightweight and local while enabling the cloud layer to scale independently.

---

## 2. Technical Strategy

Razad should follow a **control plane / execution plane** model.

- **Execution Plane:** runs on the customer server or managed node. Responsible for app execution, systemd, nginx, DB service management, logs, and local enforcement.
- **Control Plane:** runs in Razad Cloud. Responsible for dashboard, auth, organization, billing, provisioning orchestration, fleet visibility, and AI coordination.

This separation is the core scalability decision for the product.

---

## 3. Architectural Goals

1. Keep the local daemon small and reliable.
2. Avoid Docker dependency in the core product.
3. Use systemd as the process supervisor.
4. Embed the frontend into a single distributable binary for self-hosted mode.
5. Make the cloud layer stateless where possible.
6. Treat AI as a policy-restricted automation assistant, not a general shell proxy.
7. Preserve an upgrade path from self-hosted OSS to managed infrastructure without a rewrite.

---

## 4. System Overview

### 4.1 Major Components

- **razad-daemon**: Go service running on the target server.
- **razad-ui**: SvelteKit frontend compiled to static assets and embedded into the Go binary for self-hosted mode.
- **razad-api**: API layer exposed by the daemon for local UI and remote cloud control.
- **razad-agent**: local execution/observer component responsible for trusted operations.
- **razad-cloud**: optional SaaS control plane for BYO VPS and managed infrastructure.
- **razad-ai-worker**: isolated AI orchestration worker.
- **razad-event-bus**: event transport layer for cloud-scale notifications.

### 4.2 Communication Modes

- **Local mode:** UI talks to local daemon over local authenticated API.
- **BYO VPS mode:** cloud control plane talks to agent via secure outbound connection.
- **Managed infrastructure mode:** cloud control plane provisions and controls managed nodes through the same agent protocol.

---

## 5. Deployment Model

### 5.1 Self-Hosted Deployment

A single Go binary includes:

- backend API
- embedded frontend assets
- process/domain/database management logic
- WebSocket hub
- AI policy gate

This binary runs under systemd.

### 5.2 Cloud Deployment

The cloud product is separated into services:

- web app / dashboard
- API gateway
- identity service
- billing service
- provisioning service
- AI service
- agent registry
- audit/event storage

### 5.3 Managed Infrastructure Deployment

Razad provisions or claims nodes from a provider or internal pool. Each node runs the agent and reports health back to the control plane.

---

## 6. Repository Architecture

### 6.1 Monorepo Layout

```text
razad/
├── cmd/razad/
├── internal/
├── web/
├── scripts/
├── configs/
├── docs/
└── tests/
```

### 6.2 Internal Package Boundaries

- `internal/api` — REST handlers, auth guards, request validation.
- `internal/agent` — reactive loop, policy engine, observers, action executor.
- `internal/process` — systemd service generation and lifecycle control.
- `internal/proxy` — Nginx configuration generation and reload orchestration.
- `internal/database` — service provisioning for MySQL, PostgreSQL, Redis.
- `internal/ssl` — Certbot integration.
- `internal/runtime` — runtime detection and execution abstraction.
- `internal/config` — configuration loading, encryption at rest, secret management.
- `internal/ws` — streaming logs and live events.
- `internal/audit` — immutable privileged action audit trail.

---

## 7. Execution Plane Design

### 7.1 razad-daemon Responsibilities

The daemon is the authoritative process on each node.

It must:

- accept authenticated API requests,
- create/update/remove app service definitions,
- manage proxy and SSL configuration,
- provision supported local services,
- stream logs,
- report node health,
- enforce local policy for AI and user actions,
- persist audit records.

### 7.2 Local Privilege Model

The daemon should run with the minimum privileges required and use controlled elevation for system operations. The design should prefer explicit command wrappers over arbitrary shell execution.

### 7.3 Service Identity

Each application should map to:

- a dedicated Linux service identity,
- a dedicated working directory,
- a dedicated config record,
- a dedicated log stream.

This reduces blast radius and simplifies observability.

---

## 8. Control Plane Design

### 8.1 Responsibilities

The cloud control plane must provide:

- user authentication,
- organization and workspace management,
- server registration,
- node provisioning,
- billing,
- AI orchestration,
- fleet health aggregation,
- audit visibility,
- policy distribution.

### 8.2 Stateless Preference

The cloud API should remain stateless wherever possible. Node state belongs on the execution node; cloud state belongs to durable storage and event streams.

### 8.3 Node Registration Flow

1. User creates or connects a server.
2. Cloud issues a node enrollment token.
3. Agent enrolls outbound.
4. Agent receives identity and policy.
5. Agent begins heartbeat and event streaming.

---

## 9. Agent Protocol

### 9.1 Transport

Use an outbound-first secure transport such as HTTPS/WebSocket with mutual authentication or signed session tokens.

### 9.2 Message Types

- heartbeat
- health snapshot
- app status update
- log stream event
- action request
- action result
- policy update
- enrollment request

### 9.3 Delivery Guarantees

The system should favor at-least-once delivery for important events and idempotent command handling on the agent.

---

## 10. Application Lifecycle Architecture

### 10.1 App Model

An app is a managed unit with:

- source location or upload artifact,
- runtime,
- environment variables,
- service configuration,
- domain bindings,
- log configuration,
- resource statistics.

### 10.2 Lifecycle States

- `created`
- `provisioning`
- `starting`
- `running`
- `stopped`
- `crashed`
- `updating`
- `deleted`

### 10.3 Zero-Downtime Preference

When supported by the runtime and service model, Razad should prefer reload or rolling handoff semantics instead of hard restarts.

---

## 11. Runtime Detection Design

### 11.1 Detection Inputs

- repository files,
- package manifests,
- lockfiles,
- entrypoints,
- project conventions,
- explicit user override.

### 11.2 Detection Priority

1. Explicit user selection.
2. Strong manifest evidence.
3. Heuristic fallback.
4. Manual confirmation when ambiguous.

### 11.3 Supported Runtimes

- Node.js
- Bun
- PHP
- Go
- Python
- Ruby

### 11.4 Execution Strategy

Each runtime should map to a standardized start command template and environment contract.

---

## 12. Process Management Design

### 12.1 systemd Wrapper

systemd is the core process supervisor. Razad generates unit files instead of inventing a new supervisor.

### 12.2 Unit File Responsibilities

- working directory
- environment file path
- execution command
- restart policy
- user/group binding
- resource limits where appropriate
- log routing

### 12.3 Process Control API

Required operations:

- create service
- start service
- stop service
- restart service
- reload service
- inspect status
- remove service

---

## 13. Nginx / Proxy Design

### 13.1 Proxy Generation

Razad should generate Nginx server blocks for each domain binding.

### 13.2 Proxy Responsibilities

- terminate inbound HTTP/S traffic,
- route to app backends,
- support HTTP to HTTPS redirects,
- support websocket upgrade headers,
- apply sane default timeouts.

### 13.3 Reload Semantics

Proxy config changes should be validated before apply and reloaded safely.

---

## 14. Database Provisioning Design

### 14.1 Supported Services

- MySQL
- PostgreSQL
- Redis

### 14.2 Provisioning Model

The daemon should create:

- service configuration,
- data directory,
- credentials,
- user privileges,
- connection metadata.

### 14.3 Backup Model

v1 supports manual backup only. Scheduled backup belongs to later versions.

---

## 15. SSL Design

### 15.1 Certificate Flow

- validate domain binding,
- generate or update Nginx config,
- request certificate via Certbot,
- install certificate,
- reload proxy safely.

### 15.2 Failure Handling

The system must preserve prior working configuration if certificate issuance fails.

---

## 16. Log Streaming Design

### 16.1 Streaming Layer

Use WebSocket for live log streaming to the UI and optionally to cloud observers.

### 16.2 Log Sources

- application stdout/stderr,
- process manager events,
- database service logs,
- proxy logs,
- audit events.

### 16.3 Log Retention

Retention should be configurable and enforced locally.

---

## 17. AI Architecture

### 17.1 AI Layer Boundary

The AI layer may observe and recommend, but it never directly gains unrestricted system access.

### 17.2 AI Flow

1. Observe logs/metrics/events.
2. Detect anomaly.
3. Classify likely issue.
4. Map to approved action candidates.
5. Request approval or execute if policy permits.
6. Record audit event.

### 17.3 Provider Integration

Support BYOK providers:
- OpenAI
- Anthropic
- Google Gemini
- Ollama

### 17.4 AI Execution Rule

AI must only call explicitly whitelisted actions. The agent must reject all unregistered commands.

---

## 18. Action Registry Design

### 18.1 Allowed Actions

- `restart_app`
- `reload_nginx`
- `clear_app_cache`
- `restart_database_service`
- `scale_worker_count` (up only)
- `send_alert_notification`
- `run_predefined_healthcheck`

### 18.2 Forbidden Actions

- `delete_app`
- `delete_database`
- `drop_table`
- `modify_env_production`
- `uninstall_runtime`
- `modify_firewall_rules`
- `execute_arbitrary_command`

### 18.3 Policy Enforcement

The registry must be enforced at the execution layer, not only at the UI.

---

## 19. Data Model Overview

### 19.1 Core Entities

- User
- Organization
- Server
- Node
- App
- Domain
- Database
- Service
- EnvironmentVariable
- AuditEvent
- LogStreamCursor
- AIAction
- ProviderCredential
- ProvisioningJob

### 19.2 Entity Ownership

In cloud mode, entities are scoped by organization and node. In self-hosted mode, they are scoped by local installation.

---

## 20. Security Design

### 20.1 Authentication

- Self-hosted mode may start with local authentication plus session tokens.
- Cloud mode requires organization-aware auth.

### 20.2 Secret Handling

- encrypt secrets at rest,
- avoid logging sensitive values,
- isolate provider credentials,
- minimize secret exposure to the frontend.

### 20.3 Authorization

Permission checks must occur on the server side for all privileged actions.

### 20.4 Audit Trail

All privileged actions must create immutable audit records.

---

## 21. Scalability Design

### 21.1 Self-Hosted Scale

A single daemon should support many services on one server as long as host resource limits allow it.

### 21.2 Cloud Scale

Cloud scalability comes from separating the control plane from nodes and storing event history outside the runtime path.

### 21.3 Node Pool Model

Managed infrastructure can be implemented as a pool of nodes with per-tenant isolation and scheduling logic.

### 21.4 Suggested Infrastructure Choices

- **API/service RPC:** Go HTTP or gRPC where justified
- **Event bus:** NATS for fleet events
- **Primary DB:** PostgreSQL for control plane metadata
- **Cache/session layer:** Redis if needed
- **Object storage:** for logs, backups, and artifacts

---

## 22. Managed Infrastructure Provisioning Design

### 22.1 Provisioning Options

- provider API provisioning,
- internal node pool allocation,
- later: customer-selected region and size.

### 22.2 Tenant Isolation Choices

Preferred sequence:

1. separate Linux users for low-risk early releases,
2. rootless containers or equivalent isolation for stronger tenancy,
3. full cluster scheduler only if business demand justifies it.

### 22.3 Abstraction Requirement

The managed infrastructure layer must use the same agent contract as BYO VPS nodes whenever possible.

---

## 23. Failure Modes and Recovery

### 23.1 Local Failure Types

- app crash
- DB service down
- nginx config failure
- cert renewal failure
- disk pressure
- log stream interruption

### 23.2 Recovery Policy

- try safe non-destructive recovery first,
- fall back to user notification,
- never silently apply destructive recovery.

### 23.3 Rollback Policy

Config changes should be staged and validated before commit.

---

## 24. Observability and Metrics

### 24.1 Host Metrics

- CPU
- RAM
- disk usage
- load average
- process state

### 24.2 Product Metrics

- app uptime
- restart frequency
- deployment success rate
- SSL issuance success rate
- AI action success rate
- provisioning success rate

### 24.3 Event Model

Important system events should be emitted uniformly so they can be consumed by UI, AI, and cloud services.

---

## 25. Recommended Rollout Phases

### Phase 1 — OSS Self-Hosted Core

- daemon
- UI
- app management
- process management
- logs
- database provisioning
- SSL and domains

### Phase 2 — BYO VPS Cloud Control Plane

- auth
- node registration
- fleet visibility
- cloud AI coordination
- remote observability

### Phase 3 — Managed Infrastructure

- node provisioning
- tenant allocation
- provider API integration
- cloud billing and lifecycle automation

---

## 26. Risks

### Risk 1: Tight coupling between cloud and local agent
**Mitigation:** formalize the agent protocol and keep node state local.

### Risk 2: Unsafe AI action execution
**Mitigation:** whitelist registry, policy engine, audit logs.

### Risk 3: Managed infrastructure becomes too complex too early
**Mitigation:** launch with BYO VPS first and keep provisioning adapters thin.

### Risk 4: Self-hosted binary becomes too large
**Mitigation:** keep cloud services out of the self-hosted binary.

---

## 27. Open Questions

- Should the cloud control plane start as a separate repo or remain a monorepo package?
- Which provider should be first for managed VPS provisioning?
- Should managed nodes use rootless containers or separate Linux users in phase 1?
- What is the minimal credential model for BYO VPS enrollment?
- How should billing integrate with provisioning state?

---

## 28. Technical Definition of Done

The TDD v1.0 is considered implementable when the following are true:

1. The self-hosted daemon can manage apps, domains, SSL, databases, logs, and AI policy locally.
2. The cloud model is separable from the execution node.
3. The agent protocol is defined and secure enough for BYO VPS onboarding.
4. The action registry is enforced outside the UI.
5. The architecture can extend into managed infrastructure without a rewrite.

