# Razad
## Security Architecture v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Scope:** Self-hosted OSS, Razad Cloud BYO VPS, Razad Managed Infrastructure  
**Primary Goal:** Define the trust model, security boundaries, secret handling, authorization model, and AI safety controls.

---

## 1. Purpose

This document defines how Razad should protect users, servers, applications, credentials, and automation privileges.

Razad is not a passive dashboard. It can create services, issue certificates, manage proxies, provision databases, and execute approved actions on live systems. Because of that, security is a core architectural concern rather than a feature layer.

---

## 2. Security Objectives

1. Prevent unauthorized access to control plane and node resources.
2. Protect secrets at rest and in transit.
3. Ensure privileged actions are authenticated, authorized, and audited.
4. Restrict AI to explicit, non-destructive, whitelisted actions.
5. Preserve tenant isolation across organizations, projects, and nodes.
6. Reduce blast radius through least privilege and scoped execution.
7. Make security controls understandable enough for small teams to operate safely.

---

## 3. Security Principles

### 3.1 Least Privilege

Every component should operate with the minimum privilege required to perform its job.

### 3.2 Explicit Trust Boundaries

Cloud, node, browser, AI provider, and operating system must each be treated as separate trust zones.

### 3.3 Fail Closed

When validation, authorization, or policy evaluation fails, the request must be denied.

### 3.4 Audit Everything Important

All privileged actions, especially AI-driven actions, must generate immutable audit events.

### 3.5 No Arbitrary Execution

Neither users nor AI should gain unrestricted shell execution through the product.

### 3.6 Safe-by-Default Automation

Automation should prefer observation, recommendation, and reversible actions.

---

## 4. Threat Model

### 4.1 Primary Threats

- Unauthorized UI access
- Stolen API tokens
- Compromised node agent
- Cross-tenant data leakage
- Secret exfiltration
- Malicious or buggy AI behavior
- Config injection into Nginx/systemd/database templates
- Replay or spoofing of node commands
- Provider credential leakage
- Abuse of managed infrastructure accounts

### 4.2 Secondary Threats

- Log poisoning
- Privilege escalation through service wrappers
- SSRF through remote fetch or provisioning flows
- Weak enrollment token handling
- Unsafe rollback or recovery behavior
- Weak password handling
- Improper deletion of security-relevant audit history

---

## 5. Trust Boundaries

```text
[Browser/User]
      |
      v
[Web UI] -----> [Control Plane API]
                         |
                         v
                  [Policy Engine]
                         |
                         v
                  [Node Agent Channel]
                         |
                         v
               [Execution Plane on Server]
                         |
       +-----------------+------------------+
       |                 |                  |
    systemd            nginx              db services
```

### 5.1 Browser Trust Boundary

The browser is untrusted input. It may only interact with public APIs and should never receive secrets that are not strictly needed.

### 5.2 Control Plane Trust Boundary

The cloud control plane manages identity, policy, billing, node registry, and orchestration. It must not directly assume local system privileges on nodes.

### 5.3 Node Trust Boundary

The node agent is the local authority for server-level mutations. Cloud can request actions, but the node must verify authorization and policy before execution.

### 5.4 AI Provider Trust Boundary

External AI providers are untrusted from a security standpoint. They may help generate analysis or recommendations, but they must never receive unnecessary secrets or direct privileged access.

### 5.5 Operating System Trust Boundary

The OS is the final execution environment. systemd, Nginx, and database services must be wrapped and constrained, not blindly exposed.

---

## 6. Identity and Authentication

### 6.1 User Authentication

Razad should support secure login for both self-hosted and cloud modes.

**Required properties**
- Passwords must be hashed using a strong adaptive hash.
- Sessions must be protected against fixation and theft.
- Authentication cookies must use secure flags in production.
- Login failure responses must not reveal account existence.

### 6.2 Session Model

Recommended default for self-hosted mode:
- server-side session storage
- signed session cookie
- CSRF protection for state-changing requests

Recommended default for cloud mode:
- session-based web login
- API tokens or short-lived access tokens for automation and node communication

### 6.3 API Authentication

All API requests that mutate state must require authenticated identity.

Possible auth types:
- session cookie for browser traffic
- bearer token for service integrations
- node enrollment token for agent bootstrap
- signed node session for ongoing agent communication

### 6.4 Node Authentication

Nodes must not trust unsigned incoming commands.

Recommended approach:
- enrollment token issued by control plane
- one-time or short-lived bootstrap credential
- derived node identity key pair
- signed requests or mutual transport authentication for persistent sessions

---

## 7. Authorization Model

### 7.1 Scope Levels

Razad should enforce authorization at these scopes:

- user
- organization
- project
- server
- app
- database instance
- domain
- AI action

### 7.2 Role Model

Suggested roles:
- owner
- admin
- operator
- viewer

### 7.3 Permission Style

Use capability-based or explicit permission checks rather than broad implicit trust.

Examples:
- `app:create`
- `app:restart`
- `domain:bind`
- `database:provision`
- `node:enroll`
- `ai:approve_action`
- `billing:manage`

### 7.4 Authorization Rules

- A user may only operate on resources inside their organization scope.
- A project-scoped resource may only be accessed by users with project access.
- Node-level actions must be validated against node ownership and node status.
- AI action execution requires both policy allowance and resource ownership validation.

---

## 8. Secret Management

### 8.1 Secret Categories

- user passwords
- session signing keys
- node enrollment tokens
- database credentials
- API provider keys
- SSH or provider access keys
- encryption keys for env vars
- SSL private keys

### 8.2 Storage Rules

- Secrets must never be stored plaintext in application metadata.
- Environment variables marked secret must be encrypted at rest.
- Provider credentials must be isolated from normal application config.
- Secrets must not appear in logs, audit payloads, or client responses.

### 8.3 Encryption at Rest

Razad should use application-level encryption for sensitive fields.

Recommended pattern:
- master key stored outside the database where possible
- per-record or per-scope encrypted payloads
- key rotation path defined early

### 8.4 Secret Exposure Controls

- show only last four characters or masked values in UI
- never return full secret after initial creation unless the platform design explicitly allows re-display via secure re-authentication
- redact secrets in error traces

---

## 9. Transport Security

### 9.1 External Transport

- All cloud traffic must use HTTPS.
- Browser sessions must only operate over TLS in production.
- Certificates should be valid and automatically renewed.

### 9.2 Node Transport

- Node-agent communication must be encrypted.
- The channel must resist spoofing and replay.
- Node commands should be signed or established over mutually authenticated sessions.

### 9.3 Internal Service Transport

If the cloud control plane is decomposed into multiple services, internal traffic should also be protected with service identity and transport encryption where appropriate.

---

## 10. Tenant Isolation

### 10.1 Isolation Targets

Isolation must exist at multiple levels:
- database row scoping
- API authorization
- node identity scoping
- filesystem path scoping
- OS user or container scoping for managed workloads

### 10.2 Data Isolation

Every record must be clearly scoped to an organization, project, or server.

### 10.3 Execution Isolation

For managed infrastructure, customer workloads should not share unnecessary privilege boundaries. Early versions may use separate Linux users; stronger isolation can later move to rootless containers or equivalent.

### 10.4 Cloud Tenant Isolation

The cloud control plane must never allow one tenant to enumerate or mutate another tenant’s nodes, apps, domains, secrets, or audit logs.

---

## 11. AI Safety Architecture

This is one of the most important parts of Razad.

### 11.1 AI Security Principle

AI is an advisor and bounded executor. It is never a general-purpose operator.

### 11.2 Allowed AI Behavior

- inspect logs and metrics
- summarize incidents
- suggest corrective actions
- run approved health checks
- restart approved services
- notify users
- trigger other explicitly allowed non-destructive actions

### 11.3 Forbidden AI Behavior

AI must never be allowed to:
- execute arbitrary shell commands
- delete apps
- delete databases
- drop tables
- modify firewall rules
- uninstall runtimes
- alter production environment variables without explicit policy approval
- bypass audit logging

### 11.4 AI Policy Enforcement Layer

The AI output must pass through a policy engine before any execution request is sent to the node.

Policy engine checks:
- action is registered
- action is allowed by current policy
- target resource belongs to the correct tenant
- action is non-destructive
- execution context is valid
- approval requirements are satisfied

### 11.5 Human-in-the-loop Modes

Recommended modes:
- observe only
- recommend
- execute safe actions automatically
- execute only after user confirmation

The default for early versions should be conservative.

---

## 12. Node Security

### 12.1 Node Agent Privilege

The agent should run with minimum privilege and use tightly scoped elevation for system tasks.

### 12.2 Local Command Wrappers

The agent should use explicit wrappers for approved operations rather than constructing arbitrary shell calls.

### 12.3 Filesystem Safety

- app working directories must be isolated
- config generation must sanitize inputs
- uploaded artifacts must be validated before execution
- temp files must be placed in controlled directories

### 12.4 Systemd Safety

Generated unit files must be validated before reload or restart.

### 12.5 Nginx Safety

Generated proxy configs must be syntax-checked before apply.

---

## 13. Billing and Provisioning Security

### 13.1 Provider Credential Protection

Provider keys used for managed infrastructure or provisioning integrations must be scoped, encrypted, and rotatable.

### 13.2 Provisioning Abuse Prevention

Provisioning requests should be rate-limited and authorized by tenant and billing status.

### 13.3 Resource Creation Guardrails

Before creating a node or billing-backed resource, Razad should confirm:
- tenant entitlement
- quota availability
- provider availability
- required approval state

---

## 14. Logging and Audit Security

### 14.1 Audit Requirements

Every significant privileged action must create an audit event with:
- who initiated it
- what was targeted
- what action was requested
- when it happened
- whether it succeeded
- whether AI was involved

### 14.2 Audit Immutability

Audit logs should be append-only from the product’s perspective.

### 14.3 Log Redaction

Sensitive data must be removed or masked before storage in logs.

### 14.4 Incident Traceability

The system should be able to reconstruct:
- a node action history
- an AI decision chain
- a provisioning sequence
- a failed deployment path

---

## 15. Input Validation and Injection Defense

### 15.1 Sources of Untrusted Input

- browser forms
- API payloads
- repository metadata
- uploaded archives
- environment variable values
- domain names
- runtime commands
- AI-generated suggestions
- provider response payloads

### 15.2 Validation Requirements

- validate type, format, and length
- reject unsafe characters where appropriate
- sanitize shell-adjacent inputs
- escape config templates correctly
- normalize paths before use

### 15.3 Injection Surfaces

Particular attention must be given to:
- shell invocation
- systemd unit generation
- Nginx configuration
- SQL queries
- log display rendering
- Markdown or HTML rendering in the UI

---

## 16. Authentication Flows by Mode

### 16.1 Self-hosted

- local account login
- local session cookie
- optional local admin bootstrap

### 16.2 BYO VPS Cloud

- cloud account login
- organization-scoped access
- node enrollment token
- signed node session

### 16.3 Managed Infrastructure

- cloud account login
- billing-gated provisioning permissions
- node agent bootstrap under Razad-controlled lifecycle

---

## 17. Secure Defaults

Razad should ship with conservative defaults:

- HTTPS required in cloud mode
- secrets encrypted at rest
- AI starts in observe or recommend mode
- destructive actions blocked by default
- audit logging enabled by default
- node enrollment tokens expire quickly
- config changes validated before apply

---

## 18. Incident Response Design

### 18.1 Security Incident Categories

- leaked credential
- compromised node
- suspicious AI action
- unauthorized access attempt
- tenant boundary violation
- provisioning abuse

### 18.2 Response Capabilities

Razad should support:
- revoking sessions
- revoking node tokens
- disabling AI execution
- rotating provider credentials
- isolating a server from control plane actions
- preserving audit evidence

---

## 19. Encryption and Key Management

### 19.1 Key Hierarchy

Recommended hierarchy:
- root/master key
- tenant or installation key
- field-level encrypted payloads

### 19.2 Rotation

The architecture should allow key rotation without requiring total data loss.

### 19.3 Backup of Secrets

Backup procedures must treat secrets as sensitive assets and preserve confidentiality even if backup storage is compromised.

---

## 20. Security Controls by Component

### 20.1 UI

- CSRF protection
- secure session handling
- XSS-safe rendering
- no secret leakage in page state

### 20.2 API

- authentication required
- authorization required
- request validation
- rate limiting where appropriate

### 20.3 Agent

- authenticated command intake
- whitelist-only action execution
- local validation before mutation
- audit emission

### 20.4 AI Service

- prompt isolation
- provider key isolation
- policy gate before execution
- no direct privileged access

### 20.5 Cloud Control Plane

- tenant scoping
- billing checks
- node identity checks
- audit trail persistence

---

## 21. Recommended Security Posture by Phase

### Phase 1 — Self-Hosted OSS

Focus on:
- strong local auth
- encrypted secrets
- safe template generation
- audit logging
- explicit AI restrictions

### Phase 2 — BYO VPS Cloud

Add:
- node enrollment security
- secure remote control plane communication
- organization-scoped authorization
- node isolation policies

### Phase 3 — Managed Infrastructure

Add:
- provisioning abuse prevention
- billing-backed authorization
- stronger tenant isolation
- fleet-level incident controls

---

## 22. Security Testing Strategy

### 22.1 Unit Tests

- auth checks
- permission checks
- secret masking
- policy engine decisions
- template sanitization

### 22.2 Integration Tests

- node enrollment
- action execution flow
- SSL issuance flow
- config reload flow
- audit event generation

### 22.3 Negative Tests

- invalid token rejection
- cross-tenant access rejection
- AI forbidden action rejection
- malformed config input rejection
- secret leakage prevention

### 22.4 Security Review Triggers

Any change to the following should require explicit review:
- auth flows
- secret storage
- AI action registry
- node protocol
- provider credential handling
- billing-gated provisioning logic

---

## 23. Open Security Questions

- What exact session strategy should be used in self-hosted mode?
- Should node communication use mTLS from the start or evolve toward it?
- Which encryption library and key management approach should be standardized?
- Should AI recommendations be stored separately from user-visible action logs?
- What minimum assurance should managed infrastructure require before allowing customer workloads?

---

## 24. Security Definition of Done

Security architecture is acceptable for implementation when the following are true:

1. All trust boundaries are explicit.
2. User, node, and cloud authentication flows are defined.
3. Secret storage is encrypted at rest.
4. AI can only execute whitelisted actions.
5. All privileged actions generate audit events.
6. Tenant isolation is enforced at data and execution levels.
7. Templates and commands are validated before execution.
8. The system fails closed when policy or authorization is uncertain.

---

## 25. Closing Statement

Razad’s security model must be conservative enough for production, but simple enough for small teams to understand.

The architecture succeeds only if the product remains useful without ever becoming a privilege shortcut.

