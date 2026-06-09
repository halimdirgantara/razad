# Razad
## System Architecture Document (SAD) v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Scope:** Self-hosted OSS, Razad Cloud BYO VPS, Razad Managed Infrastructure  
**Primary Stack:** Go, SvelteKit, systemd, Nginx, PostgreSQL, Redis, NATS  
**License Assumption:** AGPLv3

---

## 1. Purpose

This document defines the system-level architecture of Razad. It explains how the major subsystems interact, how responsibilities are separated, how requests move through the platform, and how the design supports both self-hosted and cloud-managed operating modes without collapsing into a single fragile monolith.

The goal is not only to describe components, but to define boundaries, contracts, and control flow in a way that can support implementation, scaling, and future expansion.

---

## 2. System Overview

Razad is a server management and deployment platform with three operational modes:

1. **Self-hosted** — the user installs Razad on a server they control.
2. **Razad Cloud BYO VPS** — the user connects their own VPS, and Razad manages it remotely.
3. **Razad Managed Infrastructure** — Razad provisions and operates infrastructure for the customer.

The architecture is intentionally split into two planes:

- **Execution Plane**: local server/node execution, service control, config application, app runtime management, logging.
- **Control Plane**: identity, dashboard, orchestration, billing, fleet visibility, AI coordination, provisioning workflows.

This separation is the primary scalability mechanism of the product.

---

## 3. Architectural Goals

### 3.1 Product Goals translated into architecture

- Keep the local node lightweight.
- Avoid Docker dependency in the core execution path.
- Use systemd as the process supervisor.
- Keep app, database, SSL, and proxy operations deterministic.
- Make AI policy-restricted and auditable.
- Support cloud growth without rewriting the self-hosted core.

### 3.2 Design Priorities

1. Safety
2. Determinism
3. Simplicity
4. Observability
5. Scalability
6. Extensibility

### 3.3 Non-Goals

- Kubernetes orchestration.
- General-purpose shell access from AI.
- Full public-cloud feature parity.
- Arbitrary plugin execution in the first version.
- Distributed consensus across nodes inside the self-hosted binary.

---

## 4. Architecture Principles

### 4.1 Native-first

Razad should manage native Linux services rather than abstracting everything through containers. systemd and Nginx are first-class primitives.

### 4.2 Boundary-driven design

Every major responsibility must have a clear owner. The UI does not own business logic. The control plane does not directly mutate node state without the agent. AI does not own execution privileges.

### 4.3 Safe automation

Automation is allowed only through explicit action registration. Destructive operations must remain blocked.

### 4.4 Local authority

The execution node must be the authority for local system changes. Cloud services may request actions, but the node enforces them.

### 4.5 Event visibility

Important state changes should become events that can be consumed by the UI, control plane, and AI layer.

---

## 5. High-Level Component Map

```text
                           ┌──────────────────────────┐
                           │      Razad Cloud         │
                           │  (optional control plane)│
                           │                          │
                           │  - Auth                  │
                           │  - Dashboard             │
                           │  - Billing               │
                           │  - Fleet orchestration   │
                           │  - AI coordination       │
                           │  - Provisioning          │
                           └───────────┬──────────────┘
                                       │
                                       │ secure outbound channel
                                       │
                    ┌──────────────────▼──────────────────┐
                    │            Razad Agent              │
                    │   (execution authority on node)     │
                    │                                      │
                    │  - app lifecycle                    │
                    │  - systemd wrapper                  │
                    │  - Nginx generation/reload          │
                    │  - DB provisioning                  │
                    │  - SSL issuance                     │
                    │  - log streaming                    │
                    │  - policy enforcement               │
                    └───────────┬───────────┬────────────┘
                                │           │
                                │           │
                       ┌────────▼───┐   ┌──▼─────────┐
                       │  systemd   │   │   Nginx    │
                       └─────┬──────┘   └────┬───────┘
                             │               │
                    ┌────────▼────────┐  ┌──▼──────────┐
                    │ App Processes    │  │ Public HTTP │
                    └──────────────────┘  └─────────────┘
```

---

## 6. Deployment Modes

### 6.1 Self-hosted Mode

In self-hosted mode, the entire product runs on the customer’s machine as a single installable package.

**Characteristics**
- One binary for backend logic.
- Embedded frontend assets.
- Local SQLite or PostgreSQL for metadata, depending on deployment choice.
- Local auth and session handling.
- No dependence on Razad Cloud.

### 6.2 BYO VPS Cloud Mode

In BYO VPS mode, the customer connects their server to Razad Cloud.

**Characteristics**
- Cloud stores account, org, billing, and fleet state.
- Node agent runs on customer VPS.
- Execution remains local to the node.
- Cloud issues requests through a secure agent protocol.

### 6.3 Managed Infrastructure Mode

In managed infrastructure mode, Razad provisions infrastructure on behalf of the customer.

**Characteristics**
- Customer interacts with cloud dashboard.
- Cloud provisions server or server pool.
- Agent is installed on managed node automatically.
- Same execution model as BYO VPS.

---

## 7. Core Runtime Architecture

### 7.1 Self-hosted Binary

The self-hosted product should compile into a single Go binary that includes:

- HTTP API server
- authentication/session management
- app management handlers
- systemd orchestration layer
- Nginx config generator
- database service manager
- SSL orchestration
- WebSocket log streaming
- policy gate for AI actions
- embedded static frontend assets

### 7.2 Cloud Service Set

The cloud product should be decomposed into services:

- Web frontend / dashboard
- API gateway
- Identity service
- Organization service
- Billing service
- Node registry service
- Provisioning service
- AI service
- Audit/event service
- Notification service

### 7.3 Agent Process

The agent is a small node-local process responsible for trusted execution.

**Responsibilities**
- heartbeat
- health snapshot
- local action execution
- log streaming
- policy enforcement
- secure command handling
- local audit event emission

---

## 8. Service Boundary Definitions

## 8.1 UI Layer

**Owns**
- screen rendering
- form state
- user interaction
- visual feedback

**Does not own**
- business rules
- policy logic
- privileged operations

## 8.2 API Layer

**Owns**
- request validation
- authentication checks
- command routing
- response shaping

**Does not own**
- execution of privileged system actions directly unless local daemon context authorizes it

## 8.3 Agent Layer

**Owns**
- systemd execution
- Nginx reloads
- local DB provisioning
- runtime start/stop/restart
- local log access
- local policy enforcement

**Does not own**
- user interface
- billing
- global organization state

## 8.4 Control Plane Layer

**Owns**
- user/org/workspace data
- billing state
- managed infrastructure orchestration
- fleet health overview
- AI orchestration coordination

**Does not own**
- direct mutation of local machine state without the agent

## 8.5 AI Layer

**Owns**
- observation analysis
- anomaly interpretation
- recommendation generation
- candidate action selection

**Does not own**
- arbitrary execution
- destructive operations
- policy definition

---

## 9. Main Data Flow

### 9.1 App Deployment Flow

```text
User
  ↓
UI
  ↓
API
  ↓
Deployment Orchestrator
  ↓
Runtime Detector
  ↓
Service Config Generator
  ↓
Systemd Manager
  ↓
Nginx Manager (if domain attached)
  ↓
Health Check
  ↓
Running App
```

### 9.2 Database Provisioning Flow

```text
User
  ↓
UI
  ↓
API
  ↓
Database Orchestrator
  ↓
Service Provisioner
  ↓
Credential Generator
  ↓
Metadata Store
  ↓
Connection Info Returned
```

### 9.3 Domain and SSL Flow

```text
User
  ↓
UI
  ↓
API
  ↓
Domain Service
  ↓
Nginx Config Generator
  ↓
Certbot Integration
  ↓
Certificate Install
  ↓
Nginx Reload
  ↓
HTTPS Live
```

### 9.4 AI Assistance Flow

```text
Logs / Metrics / Events
  ↓
Observer
  ↓
Anomaly Detector
  ↓
AI Policy Layer
  ↓
Allowed Action Matcher
  ↓
Human Approval or Auto-Approved Safe Action
  ↓
Execution Layer
  ↓
Audit Log
```

---

## 10. Control Plane and Execution Plane Contract

### 10.1 Contract Goal

The contract must allow cloud services to control nodes without giving cloud services direct privileged system access.

### 10.2 Contract Requirements

- authentication of node identity,
- encrypted transport,
- signed or tokenized action requests,
- idempotent commands,
- action status feedback,
- replay protection,
- heartbeat and liveness tracking.

### 10.3 Recommended Interaction Pattern

The node should initiate or maintain the outbound connection whenever possible. This reduces firewall friction and simplifies deployment behind NAT.

---

## 11. Node Registration Flow

### 11.1 Self-hosted

Node registration is local only. No external enrollment required.

### 11.2 BYO VPS

```text
User creates server in cloud
  ↓
Cloud generates enrollment token
  ↓
Agent installed on customer VPS
  ↓
Agent connects outbound
  ↓
Cloud verifies identity
  ↓
Policy and config delivered
```

### 11.3 Managed Infrastructure

```text
User purchases managed environment
  ↓
Cloud provisions server
  ↓
Bootstrap installs agent
  ↓
Agent enrolls automatically
  ↓
Cloud attaches server to tenant
```

---

## 12. App Model

### 12.1 Core App Attributes

- id
- name
- owner scope
- source type
- runtime
- start command
- environment variables
- service name
- domain bindings
- health status
- logs cursor
- resource metrics

### 12.2 App Lifecycle States

- `created`
- `provisioning`
- `starting`
- `running`
- `stopped`
- `crashed`
- `updating`
- `deleted`

### 12.3 App Ownership Rule

Every app must belong to exactly one tenant scope: local installation, organization, or managed tenant.

---

## 13. Runtime Detection Architecture

### 13.1 Detection Strategy

Use layered detection:

1. explicit user override
2. strong project signals
3. heuristics
4. confirmation fallback

### 13.2 Detection Signals

- `package.json`
- `composer.json`
- `go.mod`
- `requirements.txt`
- `pyproject.toml`
- Ruby project markers
- entrypoint files
- lockfiles

### 13.3 Fallback Behavior

If runtime cannot be reliably detected, Razad must ask the user to confirm or choose manually.

---

## 14. systemd Integration Architecture

### 14.1 Rationale

systemd provides the native supervision and boot integration needed for a Linux-first product. Introducing a separate process manager would increase complexity without clear benefit.

### 14.2 Generated Unit File Content

A generated unit should define:

- service description
- execution command
- working directory
- environment file
- restart policy
- user/group identity
- start limit policy
- optional resource constraints

### 14.3 Reload and Restart Semantics

- restart for code or process changes
- reload when supported by runtime or proxy
- validate before apply
- preserve previous known-good config when possible

---

## 15. Proxy and SSL Architecture

### 15.1 Proxy Ownership

Nginx owns public routing. Razad generates and updates its configuration.

### 15.2 SSL Ownership

Certbot is the certificate issuer; Razad orchestrates the request, validation, installation, and reload sequence.

### 15.3 Safety Strategy

Any config mutation must be validated before Nginx reload. Invalid config should not replace the existing working config.

---

## 16. Database Service Architecture

### 16.1 Supported Services

- MySQL
- PostgreSQL
- Redis

### 16.2 Provisioning Model

For each database service, Razad should manage:

- configuration
- service identity
- credentials
- data directory
- startup and restart behavior
- connection metadata

### 16.3 Backup Roadmap

Manual backup is available in v1. Scheduled backup enters later versions.

---

## 17. Logging Architecture

### 17.1 Log Sources

- application stdout/stderr
- systemd service logs
- Nginx logs
- database logs
- audit logs
- AI action logs

### 17.2 Transport

Use WebSocket for live streaming to the UI, and optionally event transport to the control plane.

### 17.3 Retention

Retention should be configurable per application or per node, with sane defaults.

---

## 18. AI Safety Architecture

### 18.1 AI Role

AI is an observer and adviser with tightly constrained execution privileges.

### 18.2 Policy Flow

```text
Observation
  ↓
Classification
  ↓
Action Candidate Selection
  ↓
Policy Check
  ↓
Approval / Block / Safe Execute
  ↓
Audit
```

### 18.3 Hard Safety Constraints

The AI layer must never be able to:

- execute arbitrary shell commands,
- delete applications,
- delete databases,
- drop tables,
- modify firewall rules,
- uninstall runtimes,
- silently change production environment variables.

### 18.4 Allowed Action Registry

Allowed actions are limited to explicit registry entries such as restart, reload, health check, cache clear, notification, and safe scaling within policy.

---

## 19. Event Architecture

### 19.1 Event Types

- app.created
- app.started
- app.crashed
- domain.bound
- ssl.issued
- db.provisioned
- ai.alert.generated
- ai.action.executed
- node.heartbeat.failed
- node.enrolled

### 19.2 Event Bus Recommendation

Use an event bus such as NATS when cloud-scale control plane expansion becomes active.

### 19.3 Event Consumers

- UI
- AI service
- audit service
- notification service
- control plane analytics

---

## 20. Security Boundaries

### 20.1 Trust Zones

- Browser trust zone
- Cloud control plane trust zone
- Node trust zone
- Provider trust zone
- AI provider trust zone

### 20.2 Boundary Rules

- Browser never talks directly to privileged node internals.
- Cloud never bypasses the agent for node mutation.
- AI never bypasses policy enforcement.
- Provider credentials are isolated from normal UI data.

### 20.3 Secret Handling

- encrypt at rest,
- minimize exposure in responses,
- exclude secrets from logs,
- keep provider credentials separate from application secrets.

---

## 21. Scalability Strategy

### 21.1 Self-hosted Scale

Scale is constrained by the host machine. A single node can manage many apps as long as host resources allow it.

### 21.2 Cloud Scale

Cloud scale comes from:

- stateless services,
- horizontally scalable APIs,
- event-driven node communication,
- durable metadata storage,
- isolated agent execution.

### 21.3 Managed Infrastructure Scale

Managed infrastructure scale should initially use provider-backed VPS provisioning and a node pool model, not a custom hardware fleet.

---

## 22. Recommended Technology Placement

### 22.1 Runtime Placement

- **Go** for the daemon, agent, and orchestration logic.
- **SvelteKit** for the UI.
- **systemd** for process supervision.
- **Nginx** for ingress.
- **PostgreSQL** for control plane metadata.
- **Redis** for cache, session, or short-lived coordination if needed.
- **NATS** for event transport when cloud distribution requires it.

### 22.2 Why this placement

This split keeps the local node simple while allowing the cloud side to scale independently.

---

## 23. Failure and Recovery Model

### 23.1 Failure Classes

- app crash
- deployment failure
- DB provisioning failure
- SSL issuance failure
- proxy misconfiguration
- node disconnect
- AI action rejection

### 23.2 Recovery Rules

1. Prefer safe non-destructive repair.
2. If repair cannot be guaranteed, notify.
3. Never replace working config with invalid config.
4. Never silently escalate to destructive recovery.

### 23.3 Rollback Principle

All config changes should be applied in a rollback-friendly manner where possible.

---

## 24. Observability Strategy

### 24.1 Node Metrics

- CPU
- RAM
- disk usage
- load average
- process count
- service status

### 24.2 Product Metrics

- deployment success rate
- app startup latency
- SSL issuance success
- AI action success rate
- node enrollment success
- log stream uptime

### 24.3 Operational Principle

If the platform cannot explain its own state clearly, it is not production-ready.

---

## 25. Future Expansion Points

The following architecture extensions are already anticipated:

- multi-node scheduling
- managed tenants
- provider plugin adapters
- billing automation
- richer AI policy engine
- backup orchestration
- global audit timelines

These should be introduced without breaking the execution plane contract.

---

## 26. Open Questions

- Should control plane and agent share a monorepo or split into separate repos later?
- Which provider API should be the first managed infrastructure integration?
- What is the minimal node enrollment protocol for secure onboarding?
- What secret storage backend should be default for self-hosted mode?
- How strict should cross-tenant isolation be in the first managed infrastructure release?

---

## 27. Architecture Readiness Criteria

This architecture is ready for implementation when the following are true:

1. The execution plane responsibilities are fixed.
2. The control plane responsibilities are fixed.
3. The agent protocol is defined.
4. The AI safety boundary is explicit.
5. The deployment modes are separable.
6. The app, database, domain, SSL, and logs flows are all represented as deterministic sequences.
7. The product can grow into managed infrastructure without rewriting the core node daemon.

---

## 28. Closing Statement

Razad should remain a Linux-native, boundary-driven, safety-first platform.

The architecture succeeds if the product can scale in two directions at once:

- vertically on a single node through efficient local execution,
- horizontally in the cloud through a clean control plane and secure agent model.

