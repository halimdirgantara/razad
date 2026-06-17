# Razad
## Product Requirements Document (PRD) v1.1

**Product Name:** Razad  
**Tagline:** *Your server, guided.* / *Server kamu, jalur yang benar.*  
**Document Version:** 1.1  
**Status:** Draft Final  
**License:** AGPLv3  
**Primary Language:** English for technical terms; Indonesian explanations where helpful

> **Current implementation note:** the repository currently implements the local/self-hosted core (auth, orgs, projects, apps, encrypted env vars, local process management, health, logs, and embedded UI). The cloud, AI, proxy/SSL, billing, and managed-infrastructure sections in this document describe roadmap scope unless explicitly implemented in code.

---

## 1. Product Summary

Razad is an open-source server management and deployment platform built for solo developers and small teams that deploy applications on Linux VPS, dedicated servers, or managed infrastructure provisioned by Razad.

The platform combines lightweight native process management, automated app deployment, database provisioning, domain and SSL management, real-time logs, and a proactive AI layer that can observe server state and safely recommend or execute only approved non-destructive actions.

Razad is designed to be the opposite of heavy Docker-first panels and overly complex infrastructure tooling. The core design principle is simple: keep the server on the right path, with as little operational friction as possible.

---

## 2. Vision

Razad’s vision is to become the most practical open-source control plane for small production workloads: fast to install, easy to understand, safe to operate, and intelligent enough to reduce routine operational burden.

The product must support both self-hosted users and customers who want Razad to provide the infrastructure itself.

---

## 3. Problem Statement

Freelancers and small teams often deploy multiple applications on VPS instances but face several recurring problems:

1. Existing panels are too complex for lean teams.
2. Docker-heavy platforms introduce overhead and mental load that some teams do not want.
3. Commercial tools solve the problem but create recurring cost and vendor dependency.
4. Server operations still require manual monitoring, manual restarts, and reactive troubleshooting.
5. Most tools do not combine native process management and proactive AI assistance in a single coherent system.
6. Some teams do not own or want to manage a VPS but still need a platform that behaves like one.

As a result, teams waste time on deployment mechanics instead of shipping product.

---

## 4. Product Goals

### 4.1 Primary Goals

- Enable production app deployment to a Linux server in under 3 minutes for a standard Node.js app.
- Provide a single interface for application, process, database, domain, SSL, and log management.
- Reduce operational overhead through safe automation and intelligent monitoring.
- Offer an AI layer that can detect anomalies, explain issues, and execute only approved actions.
- Keep the deployment model lightweight: single binary, no Docker dependency, systemd-native.
- Support three operating modes: self-hosted, Razad Cloud BYO VPS, and Razad Managed Infrastructure.

### 4.2 Secondary Goals

- Provide a clear path from open-source product to sustainable hosted revenue.
- Support multi-tenant cloud operation without forcing container orchestration into the core product.
- Preserve a clean boundary between control plane and execution plane.

---

## 5. Non-Goals

Razad v1.1 explicitly does not aim to solve the following:

- Full container orchestration.
- Kubernetes management.
- Multi-cloud provisioning.
- Advanced CI/CD pipeline authoring.
- Destructive automation such as deleting applications, dropping databases, or modifying firewall rules.
- Enterprise identity management, SSO federation, or complex RBAC hierarchies beyond what is needed for v1.
- Full backup orchestration with point-in-time recovery; scheduled backups are reserved for later versions.
- Operating as a generalized public cloud competitor in the first release.

---

## 6. Target Users

### 6.1 Primary Persona

**Solo Developer**
- Deploys personal or client applications.
- Knows how to code but does not want to manage Linux server internals all day.
- Wants visibility, stability, and speed.

### 6.2 Secondary Persona

**Small Team Lead / Technical Founder**
- Manages 2–5 team members.
- Runs several production apps on one or more VPS servers.
- Wants a practical control panel without Docker complexity.

### 6.3 Tertiary Persona

**Freelance DevOps / Agency Operator**
- Maintains servers for clients.
- Needs repeatable deployment and quick recovery workflows.
- Values auditability and safe automation.

### 6.4 Managed Infrastructure Customer

**Non-Infrastructure Customer**
- Does not own a VPS.
- Wants Razad to provision infrastructure on their behalf.
- Needs a product experience similar to a managed platform rather than a self-hosted panel.

---

## 7. Product Principles

Razad v1.1 follows these principles:

1. **Native first** — systemd, Nginx, and local services are first-class citizens.
2. **Lightweight by default** — minimal runtime overhead and minimal operational dependencies.
3. **Safe by design** — AI may recommend, but destructive actions are blocked.
4. **Transparent automation** — every meaningful automated action must be auditable.
5. **Production-oriented** — features are judged by production usefulness, not novelty.
6. **Open-source aligned** — the architecture and license should support sustainable open development.
7. **Control plane separation** — cloud services should manage infrastructure, not collapse into the agent runtime.

---

## 8. Operating Modes

Razad supports three operating modes:

### 8.1 Self-Hosted

The user installs Razad on a server they already own and manage.

### 8.2 Razad Cloud BYO VPS

The customer owns the VPS, but Razad manages deployment, monitoring, domains, SSL, logs, and AI-assisted operations.

### 8.3 Razad Managed Infrastructure

The customer does not own infrastructure. Razad provisions and operates the server or server pool on their behalf.

The platform must keep the control plane separate from the execution plane so that cloud features can scale independently from server agents.

---

## 9. Scope of Version 1.1

### 9.1 Included in v1.1

- App deployment from Git repository.
- Manual app upload support.
- Runtime detection and runtime-specific execution support.
- Application start, stop, restart, delete.
- Environment variable management with encryption at rest.
- Systemd service generation and process management.
- Resource usage visibility.
- MySQL, PostgreSQL, and Redis provisioning.
- Basic database connection info.
- Domain binding.
- Nginx config generation.
- SSL issuance via Certbot.
- Real-time log streaming.
- Log filtering and retention settings.
- AI assistant with BYOK support.
- Proactive anomaly detection.
- Approved non-destructive action execution.
- Audit logging for AI actions.
- Cloud onboarding for BYO VPS.
- Managed infrastructure provisioning as a future-facing product mode.

### 9.2 Deferred to v2+

- Scheduled backups.
- Advanced rollout strategies.
- Full multi-server fleet management.
- Container support.
- Firewall orchestration.
- Arbitrary plugin marketplace.
- Advanced role-based access control.
- Arbitrary AI command execution.
- Fully autonomous remediation for high-risk events.
- Billing automation details beyond basic subscription support.

---

## 10. Functional Requirements

### 10.1 Application Management

Razad must allow users to create and manage applications from a repository or manual upload.

**Requirements**

- User can add an app from a Git repository.
- User can upload an app package manually.
- System must detect common runtimes automatically.
- Supported runtimes in v1:
  - Node.js
  - Bun
  - PHP
  - Go
  - Python
  - Ruby
- User can start, stop, restart, and delete apps.
- User can view app status in the dashboard.
- User can manage environment variables.
- Environment variables must be encrypted at rest.
- App restarts should be zero-downtime where the runtime and process model allow it.

**Acceptance Criteria**

- A newly deployed app can be brought online without leaving the panel.
- Runtime detection works for standard project layouts.
- App lifecycle actions are available in the UI and reflected in system state.

### 10.2 Process Management

Razad must control applications through systemd rather than introducing a separate process supervisor.

**Requirements**

- Generate `.service` files automatically.
- Wrap native systemd service management.
- Support auto-restart on crash.
- Display live status.
- Expose resource usage where available, including CPU and RAM.
- Ensure process operations are deterministic and traceable.

**Acceptance Criteria**

- A deployed app has a corresponding systemd service.
- Restart behavior survives server reboots.
- Service status updates appear correctly in the UI.

### 10.3 Database Management

Razad must provide basic database provisioning and credential handling.

**Requirements**

- Provision MySQL.
- Provision PostgreSQL.
- Provision Redis.
- Create database, user, and password in one action.
- Display connection information.
- Support manual backup in v1.
- Scheduled backup is not required in v1.

**Acceptance Criteria**

- A user can create a database stack without leaving the panel.
- Connection details are visible after provisioning.
- Backups can be triggered manually.

### 10.4 Domain and SSL Management

Razad must simplify domain attachment and HTTPS enablement.

**Requirements**

- Bind a custom domain to an application.
- Generate Nginx configuration automatically.
- Provision SSL certificates using Certbot and Let’s Encrypt.
- Handle common reverse proxy patterns needed for app exposure.

**Acceptance Criteria**

- A domain can be attached and routed correctly to an app.
- HTTPS becomes active after certificate provisioning.
- Configuration changes can be applied predictably.

### 10.5 Log Management

Razad must provide live operational visibility through logs.

**Requirements**

- Stream logs in real time via WebSocket.
- Filter logs by app.
- Filter by severity where supported.
- Filter by time range.
- Configure retention rules.

**Acceptance Criteria**

- Logs update without page refresh.
- A user can inspect a specific app’s recent runtime output quickly.
- Log retrieval does not require shell access.

### 10.6 AI Agent: Razad AI

Razad AI is a proactive server assistant that monitors system signals and can assist the user safely.

**Core Objectives**

- Observe logs and metrics.
- Detect anomalies.
- Explain likely issues.
- Recommend corrective actions.
- Execute only pre-approved, non-destructive actions.

**Supported Provider Model**

BYOK support must be available for:
- OpenAI
- Anthropic
- Google Gemini
- Ollama

**AI Interface Requirements**

- User can ask questions in natural language.
- AI can explain server status.
- AI can surface alerts based on observed signals.
- AI can run explicit whitelisted actions only.
- All AI actions must be recorded in audit logs.

**Safety Requirements**

AI must never be allowed to:
- execute arbitrary shell commands,
- delete applications,
- delete databases,
- drop tables,
- modify production environment variables directly,
- uninstall runtimes,
- modify firewall rules,
- perform destructive operations.

**Acceptance Criteria**

- AI can detect a crashed app and suggest or trigger an allowed recovery action.
- All actions taken by AI are visible in audit logs.
- The system blocks any action outside the explicit whitelist.

---

## 11. Allowed Action Registry

The AI layer may only execute actions in the approved registry.

### Allowed Actions

- `restart_app`
- `reload_nginx`
- `clear_app_cache`
- `restart_database_service`
- `scale_worker_count` (up only)
- `send_alert_notification`
- `run_predefined_healthcheck`

### Explicitly Blocked Actions

- `delete_app`
- `delete_database`
- `drop_table`
- `modify_env_production`
- `uninstall_runtime`
- `modify_firewall_rules`
- `execute_arbitrary_command`

### Registry Policy

Any new action must be added through explicit product review before it can be exposed to AI execution.

---

## 12. Technical Architecture

### 12.1 High-Level Stack

- **Core daemon:** Go
- **Frontend:** SvelteKit + `@sveltejs/adapter-static`
- **Frontend embedding:** built assets embedded into Go binary
- **Process management:** systemd wrapper
- **Database management:** MySQL, PostgreSQL, Redis
- **AI layer:** proactive agent with BYOK support
- **Deployment model:** single binary for self-hosted mode
- **Environment:** Linux VPS / dedicated server / managed fleet node
- **Cloud model:** optional control plane for BYO VPS and managed infrastructure

### 12.2 Repository Structure

The repository should follow the existing structure with distinct packages for API, agent, process, proxy, database, SSL, runtime detection, configuration, and WebSocket log streaming.

### 12.3 Runtime Detection Model

Razad should inspect app codebases and infer likely runtime from project signals such as dependency manifests, entry files, or language-specific markers.

Examples:
- `package.json` → Node.js / Bun
- `composer.json` → PHP
- `go.mod` → Go
- `requirements.txt` / `pyproject.toml` → Python
- Ruby project markers → Ruby

Runtime detection must remain heuristic-driven and editable by the user when needed.

---

## 13. Deployment and Installation Requirements

### Requirements

- Installable through a shell installer.
- Self-hosted on Ubuntu 22.04+ and Debian 12+.
- Must run as a systemd service.
- Must not require Docker.
- Must be possible to install in under 2 minutes under normal conditions.
- Must support a minimum server spec of 512MB RAM and 1 vCPU.
- Cloud mode must support customer-owned VPS onboarding and Razad-managed infrastructure provisioning.

### Installation Experience Goals

- One-command installation.
- Clear post-install output.
- Safe defaults.
- Minimal manual configuration.

---

## 14. Security Requirements

Security is a first-class requirement because Razad can touch production services.

### Requirements

- Encrypt environment variables at rest.
- Restrict AI to whitelist-only execution.
- Record every privileged action in audit logs.
- Prevent arbitrary shell execution from AI.
- Validate all user input in API and UI layers.
- Keep secrets out of logs.
- Ensure service configuration generation is deterministic and sanitized.

### Security Posture

Razad should default to conservative behavior. When a request has any meaningful operational risk, the product should require explicit user confirmation or block the action outright.

---

## 15. Observability Requirements

Razad must surface enough operational data for small-team server management.

### Required Signals

- App running state
- Service status
- Basic CPU usage
- Basic RAM usage
- App log stream
- Error events
- AI-triggered actions

### Dashboard Expectations

- A user should understand the health of the server at a glance.
- The dashboard must reduce the need to SSH into the machine for common checks.

---

## 16. UI / UX Requirements

### Core UI Characteristics

- Clean and functional dashboard.
- App-centric navigation.
- Clear separation between apps, databases, domains, logs, and AI.
- Status indicators that are easy to read.
- Fast access to lifecycle actions.

### UX Goals

- Minimize cognitive load.
- Make dangerous actions visually obvious.
- Avoid hiding important operational state.
- Optimize for speed of diagnosis and recovery.

---

## 17. API Requirements

Razad’s backend should expose internal APIs for the frontend and for future integrations.

### API Principles

- Stable and predictable.
- Versioned where necessary.
- Authenticated.
- Bound to the product’s privilege model.
- Designed for panel operations, not arbitrary automation exposure.

### Core API Domains

- Apps
- Databases
- Domains
- SSL
- Logs
- AI actions
- Settings
- Audit events

---

## 18. Logging and Auditability

Auditability is mandatory for trust.

### Requirements

- Log every AI-initiated action.
- Log user-initiated privileged operations.
- Include timestamp, actor, action name, target resource, and result.
- Keep audit logs separate from application runtime logs.

### Why This Matters

When the AI is allowed to act, the system must always be able to answer:
- What happened?
- Who triggered it?
- Why was it allowed?
- What changed?

---

## 19. License Decision

### Chosen License: AGPLv3

AGPLv3 is the appropriate default for Razad because:

- The project is intended to be open source.
- The product may later support a hosted or cloud version.
- AGPLv3 discourages third parties from taking the code, running it as a SaaS, and withholding their changes.
- It preserves the open-source ecosystem while protecting long-term product value.

### Implication

If a company deploys a modified Razad as a networked service, the license encourages disclosure of source changes under AGPLv3 obligations.

---

## 20. Success Metrics

### Product Success Metrics for v1.1

- A Node.js app can be deployed from Git to running production in under 3 minutes.
- Installation completes in under 2 minutes using the installer under normal conditions.
- AI can detect a crashed app and trigger an approved recovery flow without manual shell intervention.
- Users can manage apps, databases, domains, logs, and AI actions from the panel without SSH for routine operations.
- At least one end-to-end production workflow works reliably on Ubuntu 22.04+ and Debian 12+.
- The control plane can onboard both BYO VPS users and managed infrastructure users.

### Operational Success Metrics

- Low error rate during install.
- Low incidence of failed service generation.
- Reliable log streaming.
- Reliable certificate issuance flow.
- No unauthorized AI action execution.

---

## 21. Risks and Mitigations

### Risk 1: AI Takes Unsafe Actions
**Mitigation:** strict whitelist, audit logging, explicit blocking of destructive actions.

### Risk 2: systemd and proxy configuration errors break deployments
**Mitigation:** deterministic config templates, validation before apply, rollback-friendly updates.

### Risk 3: Runtime detection fails on non-standard codebases
**Mitigation:** heuristic detection with manual override.

### Risk 4: User expectations exceed v1 scope
**Mitigation:** clearly define non-goals and progressive roadmap.

### Risk 5: License choice reduces adoption in some segments
**Mitigation:** document the rationale clearly and keep the value proposition sharp.

### Risk 6: Managed infrastructure creates unnecessary architectural coupling
**Mitigation:** keep execution nodes isolated from the cloud control plane and use a formal agent protocol.

---

## 22. Milestone Definition for v1.1

Razad v1.1 should be considered complete when the following are true:

1. A user can install the product on a supported Linux server.
2. The dashboard loads and connects to the daemon.
3. The user can deploy an app from Git or upload.
4. The user can manage process lifecycle actions.
5. The user can provision at least one supported database type.
6. The user can bind a domain and enable SSL.
7. Logs stream in real time.
8. The AI assistant can observe, explain, and safely act within the whitelist.
9. Audit logs capture privileged activity.
10. The system meets the stated performance and ease-of-install goals.
11. The platform can describe and support the path from self-hosted to cloud-managed operation.

---

## 23. Out-of-Scope Clarifications

To avoid scope creep, the following are not part of Razad v1.1:

- A general-purpose remote shell.
- Full Linux administration replacement.
- Container-first workflows.
- Destructive AI remediation.
- Multi-tenant SaaS billing as a core dependency of the OSS product.
- Large enterprise governance features.
- Public cloud generalization beyond the needs of Razad-managed infrastructure.

Razad should stay focused on practical server operations for a small team.

---

## 24. Open Questions

The following items should be finalized during implementation planning:

- Exact authentication model for single-server v1.
- Password and secret storage details.
- How runtime isolation will be handled per app.
- Whether config changes apply immediately or through explicit confirmation.
- How user confirmations are surfaced for AI actions.
- Exact backup format for future compatibility.
- How managed infrastructure provisioning is abstracted in the control plane.
- Which provider APIs are used first for managed VPS provisioning.

---

## 25. Final Product Positioning

Razad is not trying to become the biggest infrastructure platform.
It is trying to become the most sensible one for small teams that want control without ceremony.

The product promise is simple:

- keep deployment lightweight,
- keep management visible,
- keep automation safe,
- and keep the server on the right path.

