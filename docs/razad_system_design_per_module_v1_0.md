# Razad
## System Design per Module v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Scope:** MVP self-hosted core with cloud-ready boundaries  
**Purpose:** Define module-level system design, responsibilities, dependencies, interfaces, and failure behavior.

---

## 1. Purpose

This document breaks Razad into implementation modules and defines the system design of each module.

The goal is to give engineering a clear map of what each module owns, how modules interact, and where the boundaries are so the product can be built without becoming a tangled monolith.

---

## 2. System-Level Partitioning

Razad should be treated as two major planes plus shared support modules:

- **Execution Plane** — local runtime, services, configs, logs, and node actions.
- **Control Plane** — auth, orchestration, tenant state, billing, AI coordination, fleet visibility.
- **Shared Support Modules** — config, secrets, audit, events, templates, validation.

For MVP, the self-hosted product can run as a single deployable binary, but the codebase must still respect these boundaries internally.

---

## 3. Module Map

### Core Modules

1. Bootstrap and Installer
2. Configuration and Secrets
3. Authentication and Sessions
4. Organization and Project Management
5. Server and Node Identity
6. App Management
7. Runtime Detection
8. Process and systemd Management
9. Proxy and Domain Management
10. SSL Management
11. Database Management
12. Log Streaming and Observability
13. AI Orchestration and Safety
14. Audit Logging
15. API Layer
16. UI Layer
17. Agent Protocol
18. Cloud Control Plane Readiness
19. Billing and Provisioning Hooks
20. Validation and Policy Engine
21. Notifications
22. Background Jobs and Scheduler

---

# 4. Module Specifications

---

## 4.1 Bootstrap and Installer Module

### Responsibility

Install Razad on a target Linux machine, configure its base directories, register the daemon with systemd, and bootstrap the first administrative access.

### Inputs

- OS environment
- install parameters
- domain or host metadata
- initial admin bootstrap token or setup flow

### Outputs

- installed binary
- systemd service file
- initial config file
- first admin account or bootstrap state
- initialized storage directories

### Dependencies

- configuration module
- secrets module
- systemd integration
- API bootstrap endpoints

### Design Notes

The installer should be idempotent where possible. A second run should not corrupt existing state and should either upgrade or no-op safely.

### Failure Modes

- missing dependencies
- unsupported OS version
- permission denied during installation
- service registration failure

### Recovery Strategy

- validate prerequisites before installation
- write clear error messages
- leave existing installation untouched on partial failure where possible

---

## 4.2 Configuration and Secrets Module

### Responsibility

Load, validate, persist, and protect application configuration and sensitive values.

### Inputs

- config files
- environment values
- installation identity
- secret payloads

### Outputs

- typed runtime configuration
- encrypted secret storage
- resolved runtime settings

### Dependencies

- installer
- auth
- app env vars
- database credentials
- AI provider credentials

### Design Notes

Configuration should be strongly typed at the boundary. Secrets must be encrypted at rest and masked in all outward-facing responses.

### Failure Modes

- malformed config
- missing required setting
- decryption failure
- key mismatch after rotation

### Recovery Strategy

- fail closed on invalid config
- support explicit config validation before apply
- keep a last-known-good config snapshot where practical

---

## 4.3 Authentication and Sessions Module

### Responsibility

Authenticate users, issue and validate sessions, and manage login/logout flows.

### Inputs

- login credentials
- session cookies
- API tokens
- bootstrap admin state

### Outputs

- authenticated user context
- session lifecycle events
- authorization principal

### Dependencies

- user schema
- sessions schema
- audit logging
- policy engine

### Design Notes

Self-hosted mode may use simple session authentication. Cloud mode will later extend the same base model with organization-aware identity and token flows.

### Failure Modes

- wrong password
- expired session
- revoked token
- CSRF failure

### Recovery Strategy

- deny access
- emit audit event for suspicious attempts where needed
- force re-authentication on token/session invalidation

---

## 4.4 Organization and Project Management Module

### Responsibility

Provide tenancy boundaries and logical grouping of app/server/database resources.

### Inputs

- user actions
- imported or created org data
- project metadata

### Outputs

- organization records
- membership relationships
- project groupings

### Dependencies

- identity module
- audit logging
- access control

### Design Notes

This module establishes the resource ownership structure that cloud mode will rely on later.

### Failure Modes

- duplicate slugs
- invalid tenant scope
- unauthorized access

### Recovery Strategy

- enforce unique constraints
- validate ownership before mutation
- record significant changes in audit logs

---

## 4.5 Server and Node Identity Module

### Responsibility

Represent managed servers and track node identity, enrollment, and liveness.

### Inputs

- enrollment token
- host facts
- agent handshake
- heartbeat payloads

### Outputs

- server records
- node agent records
- liveness state
- health visibility

### Dependencies

- organization/project ownership
- config and secrets
- agent protocol
- health snapshots

### Design Notes

A server is the host-level resource. A node agent is the trusted software identity running on that server.

### Failure Modes

- invalid enrollment
- missed heartbeats
- duplicate agent identity
- node offline state

### Recovery Strategy

- revoke invalid tokens
- mark node offline after heartbeat threshold
- prevent duplicate enrollment unless explicitly re-bound

---

## 4.6 App Management Module

### Responsibility

Create, update, deploy, start, stop, restart, and delete managed applications.

### Inputs

- app creation form
- Git repository URL
- uploaded artifact
- runtime override
- deployment request

### Outputs

- app records
- deployment records
- lifecycle state changes
- runtime execution plans

### Dependencies

- runtime detection
- process management
- env vars
- audit logging
- logs
- proxy integration when domain is attached

### Design Notes

App management is the core product module. It should be deterministic and should always surface clear state transitions.

### Failure Modes

- invalid source
- build or pull failure
- runtime mismatch
- service startup failure

### Recovery Strategy

- persist deployment attempt result
- expose failure reason in UI
- preserve last known good deployment data

---

## 4.7 Runtime Detection Module

### Responsibility

Infer application runtime from repository contents or uploaded project structure.

### Inputs

- repository file tree
- lockfiles
- manifest files
- explicit user override

### Outputs

- runtime classification
- suggested start command
- deployment hints

### Dependencies

- app management
- UI override controls

### Design Notes

Detection should be heuristic, not absolute. Explicit user selection always outranks inference.

### Failure Modes

- ambiguous project layout
- multiple runtime signals
- unsupported project conventions

### Recovery Strategy

- prompt user for explicit selection
- store runtime override in app metadata

---

## 4.8 Process and systemd Management Module

### Responsibility

Generate and manage systemd unit files, control app processes, and inspect service status.

### Inputs

- app start command
- working directory
- environment file path
- service lifecycle commands

### Outputs

- systemd unit file
- service status
- restart/start/stop results

### Dependencies

- app management
- configuration module
- validation module
- audit logging

### Design Notes

systemd is the authoritative process supervisor. Razad wraps it instead of replacing it.

### Failure Modes

- invalid unit file
- permission denied
- process crash
- service dependency failure

### Recovery Strategy

- validate unit file before install/reload
- preserve previous working unit where possible
- expose journal or service failure details in logs

---

## 4.9 Proxy and Domain Management Module

### Responsibility

Bind domains to applications and generate/load Nginx proxy configuration.

### Inputs

- domain assignment
- target app
- proxy mode
- path prefix

### Outputs

- Nginx config file
- active route mapping
- proxy status

### Dependencies

- domain records
- app records
- SSL module
- validation module

### Design Notes

This module must avoid unsafe templating. All values must be sanitized and config-validated before reload.

### Failure Modes

- invalid config template
- port mismatch
- domain conflict
- proxy reload failure

### Recovery Strategy

- syntax-check generated config before apply
- keep last-known-good config
- only reload after validation passes

---

## 4.10 SSL Management Module

### Responsibility

Issue, renew, and track SSL certificates for bound domains.

### Inputs

- domain record
- DNS/ownership readiness
- cert request trigger

### Outputs

- certificate status
- certificate metadata
- renewal schedule state

### Dependencies

- proxy module
- certbot integration
- domain bindings
- audit logging

### Design Notes

SSL issuance should be orchestrated only after domain routing is ready. Failed issuance must not break the existing route.

### Failure Modes

- DNS not pointing correctly
- ACME validation failure
- certificate renewal failure

### Recovery Strategy

- leave HTTP route intact if SSL issuance fails
- retry renewal later
- surface expiration warnings in UI

---

## 4.11 Database Management Module

### Responsibility

Provision and manage supported database services. MVP prioritizes PostgreSQL.

### Inputs

- database creation request
- service type
- credentials request
- backup or restore request

### Outputs

- database instance records
- credentials
- connection info
- backup artifacts

### Dependencies

- service management
- secrets module
- storage paths
- audit logging

### Design Notes

This module should be service-oriented. Database provisioning is not just schema creation; it includes service definition, user management, and connection metadata.

### Failure Modes

- port conflict
- service startup failure
- credential generation failure
- backup failure

### Recovery Strategy

- emit partial failure states clearly
- avoid exposing unusable credentials as active
- keep backup artifacts traceable

---

## 4.12 Log Streaming and Observability Module

### Responsibility

Stream logs in real time, persist useful operational history, and expose health snapshots.

### Inputs

- service logs
- app logs
- audit events
- node metrics
- health snapshots

### Outputs

- WebSocket log stream
- filtered log views
- health data
- status indicators

### Dependencies

- app management
- process management
- audit logging
- node agent heartbeat

### Design Notes

Observability must be built into the core user experience. If users cannot see why something is failing, the panel is incomplete.

### Failure Modes

- stream disconnect
- log backlog pressure
- metric capture gaps

### Recovery Strategy

- resume from cursor where possible
- limit memory use per stream
- fall back to recent buffered logs

---

## 4.13 AI Orchestration and Safety Module

### Responsibility

Coordinate AI prompts, provider integration, safety filtering, action selection, and execution gating.

### Inputs

- user chat messages
- logs
- metrics
- health snapshots
- policy state
- provider credentials

### Outputs

- AI responses
- recommendations
- action proposals
- approved safe actions
- audit records

### Dependencies

- logs module
- observability module
- action registry
- policy engine
- audit logging

### Design Notes

AI should never be able to act outside the approved registry. The module must separate inference from execution.

### Failure Modes

- provider timeout
- unsafe action suggestion
- policy denial
- malformed AI output

### Recovery Strategy

- fall back to recommendation-only behavior
- block any action outside the whitelist
- log the denial path for auditing

---

## 4.14 Audit Logging Module

### Responsibility

Record significant user, system, node, and AI actions in an immutable audit trail.

### Inputs

- privileged actions
- provisioning steps
- AI events
- authentication events
- policy decisions

### Outputs

- audit rows
- searchable history
- trace identifiers

### Dependencies

- all privileged modules

### Design Notes

Audit logging is not optional. It is a control-plane safety primitive.

### Failure Modes

- write failure
- payload redaction failure
- event duplication

### Recovery Strategy

- prioritize write durability
- keep payloads redacted
- ensure failure to audit blocks sensitive execution where appropriate

---

## 4.15 API Layer Module

### Responsibility

Expose a stable HTTP interface for the UI, node agent, and future cloud services.

### Inputs

- HTTP requests
- WebSocket upgrades
- auth context
- resource mutations

### Outputs

- JSON responses
- event streams
- error envelopes

### Dependencies

- every business module

### Design Notes

The API should remain resource-oriented and should not encode business logic directly in route handlers. Handlers should call domain services.

### Failure Modes

- validation failure
- auth failure
- route mismatch
- request timeout

### Recovery Strategy

- consistent error envelopes
- request IDs
- clear status codes

---

## 4.16 UI Layer Module

### Responsibility

Render the dashboard, forms, logs, and AI interactions.

### Inputs

- API responses
- WebSocket streams
- user interaction

### Outputs

- visual state
- user commands
- form submissions

### Dependencies

- API layer
- auth layer
- logs module

### Design Notes

The UI should optimize for operational clarity: what is running, what is broken, and what can be done next.

### Failure Modes

- stale state
- invalid form submission
- API desynchronization

### Recovery Strategy

- optimistic UI only where safe
- refresh authoritative state after mutation
- show explicit errors on failure

---

## 4.17 Agent Protocol Module

### Responsibility

Define the secure communication contract between cloud/control plane and node agents.

### Inputs

- enrollment token
- node handshake
- command requests
- heartbeat
- event payloads

### Outputs

- authenticated node sessions
- command acknowledgements
- node state updates

### Dependencies

- server identity
- security module
- control plane readiness

### Design Notes

The protocol must support bootstrap, trust establishment, regular operation, and revocation.

### Failure Modes

- invalid signature or token
- replayed command
- dropped heartbeat
- revoked node identity

### Recovery Strategy

- reject unauthenticated commands
- require re-enrollment if identity is revoked
- mark node offline after missed heartbeats

---

## 4.18 Cloud Control Plane Readiness Module

### Responsibility

Keep cloud-specific concerns separate from local execution while allowing the system to expand into BYO VPS and managed infrastructure.

### Inputs

- organizations
- servers
- node events
- billing state later

### Outputs

- fleet state
- node registry updates
- provisioning intent

### Dependencies

- agent protocol
- organization model
- provisioning hooks

### Design Notes

This module can be mostly dormant in MVP but should shape the architecture so the product can evolve cleanly.

### Failure Modes

- coupling control-plane logic into execution path
- state drift between cloud and node

### Recovery Strategy

- keep interfaces explicit
- prefer event-based sync
- never bypass node authority for local actions

---

## 4.19 Billing and Provisioning Hooks Module

### Responsibility

Prepare the product for future infrastructure monetization and provisioning flows.

### Inputs

- provisioning requests
- billing status
- quota state

### Outputs

- provisioning jobs
- billing events
- usage records later

### Dependencies

- control plane readiness
- organization scope
- node registry

### Design Notes

This should be thin in MVP. It exists as a foundation for cloud business expansion, not as a full engine now.

### Failure Modes

- quota mismatch
- provider failure
- provisioning timeout

### Recovery Strategy

- record job state clearly
- expose failure reason
- do not create orphaned resources silently

---

## 4.20 Validation and Policy Engine Module

### Responsibility

Centralize validation of inputs, actions, and security-sensitive decisions.

### Inputs

- commands
- configs
- AI action proposals
- provisioning requests

### Outputs

- allow/deny decisions
- validation errors
- policy results

### Dependencies

- security architecture
- action registry
- auth context

### Design Notes

This module should be consulted before execution, not after. It is the gatekeeper.

### Failure Modes

- invalid configuration
- disallowed action
- unsafe target
- malformed payload

### Recovery Strategy

- deny by default
- provide safe error detail
- log the decision path

---

## 4.21 Notifications Module

### Responsibility

Send alerts and operational notifications when important events occur.

### Inputs

- AI alerts
- node offline state
- backup failures
- SSL expiry warnings
- deployment failures

### Outputs

- in-app alerts
- email notifications later
- future webhook or chat notifications

### Dependencies

- observability module
- AI module
- provisioning and SSL modules

### Design Notes

MVP may keep notifications minimal, but the module boundary should exist early.

### Failure Modes

- notification delivery failure
- duplicated alerts
- noisy alerts

### Recovery Strategy

- debounce repeated events
- show notification status
- avoid alert storms

---

## 4.22 Background Jobs and Scheduler Module

### Responsibility

Run delayed or recurring tasks such as log retention, SSL renewal checks, health polling, and future backup schedules.

### Inputs

- scheduled tasks
- cron-like triggers
- job queue messages

### Outputs

- execution results
- retry state
- job audit events

### Dependencies

- audit logging
- SSL module
- observability module
- database module

### Design Notes

MVP may not need a rich scheduler, but recurring service tasks should still be centralized rather than scattered.

### Failure Modes

- missed job
- job overlap
- retry storm

### Recovery Strategy

- store job state
- use idempotent handlers
- limit retries and backoff aggressively

---

# 5. Cross-Module Interaction Rules

## 5.1 Execution Rule

No module should directly perform privileged system mutation without passing through the validation and policy engine.

## 5.2 Audit Rule

If a module can change system state, it must emit audit events.

## 5.3 Secret Rule

Modules may handle secrets only through the configuration and secrets module.

## 5.4 Ownership Rule

Each module should own one domain concept as much as possible. Shared logic should be explicit and minimal.

## 5.5 UI Rule

The UI may request and display state, but it should not contain business logic that belongs in the backend.

---

# 6. Recommended Call Flow Examples

## 6.1 Deploy App Flow

```text
UI -> API -> App Management -> Runtime Detection -> Process Management -> Audit
```

If domain is attached:

```text
UI -> API -> App Management -> Proxy Management -> SSL Management -> Audit
```

## 6.2 AI Restart Flow

```text
Logs/Health -> AI Orchestration -> Policy Engine -> Process Management -> Audit -> Notification
```

## 6.3 Database Provisioning Flow

```text
UI -> API -> Database Management -> Validation -> Secrets -> Audit -> Return Credentials
```

---

# 7. Module Dependency Matrix

| Module | Depends On |
|---|---|
| Installer | Config, Secrets, Auth bootstrap, systemd |
| Config & Secrets | None / base |
| Auth & Sessions | Users, Sessions, Audit |
| Org & Projects | Auth, Audit |
| Server & Node Identity | Org, Security, Agent Protocol |
| App Management | Runtime, Process, Audit, Logs |
| Runtime Detection | App Management |
| Process Management | Config, Validation, Audit |
| Proxy & Domain | App, Validation, SSL |
| SSL Management | Proxy, Certbot, Audit |
| Database Management | Process, Secrets, Audit |
| Logs & Observability | Apps, Process, Node |
| AI Orchestration | Logs, Health, Policy, Audit |
| Audit | All privileged modules |
| API | All domain modules |
| UI | API |
| Agent Protocol | Server identity, Security |
| Control Plane Readiness | Agent Protocol, Org model |
| Billing Hooks | Control plane readiness |
| Validation & Policy | Security, Action registry |
| Notifications | Observability, AI, SSL, Provisioning |
| Background Jobs | SSL, Logs, DB, Audit |

---

# 8. Implementation Order Recommendation

A practical implementation order for modules is:

1. Config and Secrets
2. Auth and Sessions
3. Organization and Project Management
4. Installer
5. API and UI shell
6. Server and Node Identity
7. App Management
8. Runtime Detection
9. Process Management
10. Proxy and Domain Management
11. SSL Management
12. Logs and Observability
13. Audit Logging
14. Validation and Policy Engine
15. Database Management
16. AI Orchestration
17. Background Jobs
18. Notifications
19. Control Plane Readiness hooks
20. Billing and Provisioning hooks

This order preserves the core value path first.

---

# 9. Module Definition of Done

A module is ready for integration when:

- its inputs and outputs are defined,
- its dependencies are explicit,
- its failure modes are known,
- its security constraints are enforced,
- its audit behavior is specified,
- it can be tested in isolation.

---

# 10. Closing Statement

Razad should be built as a product of clear modules, not as a feature pile.

The module boundaries in this document are what allow the system to remain lightweight now and scalable later.
