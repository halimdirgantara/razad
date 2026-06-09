# Razad
## API Specification v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Scope:** Self-hosted OSS, Razad Cloud BYO VPS, Razad Managed Infrastructure  
**Primary Goal:** Define the HTTP API, WebSocket channels, auth model, resource contracts, and error semantics for Razad.

---

## 1. Purpose

This document defines the API surface for Razad. It covers:

- resource-oriented REST endpoints,
- WebSocket streams for live logs and node events,
- auth and authorization behavior,
- request and response conventions,
- error formats,
- node-agent communication contract,
- cloud control plane APIs,
- safety constraints for AI-driven actions.

The API is designed to serve three consumers:

1. the web UI,
2. the node agent and cloud control plane,
3. future integrations and automation tooling.

---

## 2. API Design Principles

1. **Versioned and stable** — breaking changes require version bumps.
2. **Resource-oriented** — endpoints represent domain objects and actions clearly.
3. **Tenant-aware** — every sensitive request is scoped to an organization or local installation.
4. **Fail closed** — invalid permissions, policies, or signatures must reject requests.
5. **Audit-friendly** — privileged mutations should create audit events.
6. **Safe-by-default** — destructive or unsafe actions are excluded from the public action registry.
7. **Implementation-neutral** — the contract should support both self-hosted and cloud-managed modes.

---

## 3. Base Concepts

## 3.1 Product Modes

Razad operates in three modes:

- **Self-hosted**: API serves a single local installation.
- **BYO VPS Cloud**: API serves a remote cloud control plane and node agents.
- **Managed Infrastructure**: API also provisions and manages infrastructure nodes.

## 3.2 API Layers

- **Public UI API**: used by the web frontend.
- **Node API**: used by the node agent for execution-plane operations.
- **Control Plane API**: used by cloud services and management operations.
- **Internal Service API**: used by backend services if the cloud is split into multiple services.

---

## 4. Conventions

## 4.1 Base URL

Self-hosted examples:

```text
https://razad.local/api/v1
```

Cloud examples:

```text
https://app.razad.io/api/v1
```

Node agent examples:

```text
https://node-01.example.com/api/v1
```

## 4.2 HTTP Methods

- `GET` for reading
- `POST` for creation, command execution, and non-idempotent actions
- `PUT` for full replacement
- `PATCH` for partial update
- `DELETE` for removal

## 4.3 Content Type

All JSON endpoints use:

```text
Content-Type: application/json
```

## 4.4 ID Format

All externally visible resource IDs should be opaque identifiers such as UUIDs or comparable non-sequential IDs.

## 4.5 Pagination

List endpoints should support cursor or page-based pagination.

Recommended response fields:

- `items`
- `next_cursor` or `page`
- `limit`
- `total` where practical

## 4.6 Sorting and Filtering

Common query parameters:

- `q` for search
- `status` for state filtering
- `project_id`
- `server_id`
- `app_id`
- `created_from`
- `created_to`
- `limit`
- `cursor`

## 4.7 Date and Time

All timestamps must use ISO 8601 UTC.

Example:

```text
2026-06-07T12:00:00Z
```

## 4.8 Nullability

Optional fields may be `null` but must be documented.

---

## 5. Authentication

## 5.1 Self-Hosted Mode

Recommended auth patterns:

- secure cookie sessions for browser traffic,
- CSRF protection for state-changing browser requests,
- local bootstrap admin setup during initial install.

## 5.2 Cloud Mode

Recommended auth patterns:

- browser session or short-lived access token,
- API tokens for automation,
- organization-scoped authorization.

## 5.3 Node Agent Authentication

Node agents must authenticate using one of the following phases:

1. enrollment token for bootstrap,
2. signed node identity for persistent sessions,
3. optional mutual transport authentication if implemented.

## 5.4 Authorization

All sensitive requests must check:

- user identity,
- organization scope,
- project scope,
- server ownership,
- action policy,
- node trust state.

### Authorization Rules

- A user can only access resources inside their organization.
- A node can only execute commands approved for its tenant.
- AI actions require policy approval and resource ownership validation.
- Destructive operations must be unavailable in the v1 action registry.

---

## 6. Standard Response Format

## 6.1 Success Response

```json
{
  "success": true,
  "data": {},
  "meta": {}
}
```

## 6.2 Error Response

```json
{
  "success": false,
  "error": {
    "code": "string",
    "message": "Human readable message",
    "details": {},
    "request_id": "req_123"
  }
}
```

## 6.3 Error Codes

Recommended codes:

- `unauthorized`
- `forbidden`
- `not_found`
- `validation_error`
- `conflict`
- `rate_limited`
- `policy_denied`
- `node_unreachable`
- `action_not_allowed`
- `config_invalid`
- `internal_error`

---

## 7. Core Resource Model

The API should expose these resource groups:

- authentication
- organizations
- projects
- servers
- node agents
- apps
- deployments
- environment variables
- domains
- SSL certificates
- databases
- logs
- AI actions
- policies
- audit events
- billing
- provisioning jobs
- metrics and health

---

## 8. Authentication Endpoints

## 8.1 Login

`POST /auth/login`

Authenticate a user and create a session.

**Request**

```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

**Response**

```json
{
  "success": true,
  "data": {
    "user": {},
    "session": {
      "expires_at": "2026-06-07T12:00:00Z"
    }
  }
}
```

## 8.2 Logout

`POST /auth/logout`

Invalidate the current session.

## 8.3 Current User

`GET /auth/me`

Return the authenticated user and current scopes.

## 8.4 Session Refresh

`POST /auth/refresh`

Refresh a session or issue a new access token depending on mode.

---

## 9. Organization Endpoints

## 9.1 List Organizations

`GET /organizations`

## 9.2 Create Organization

`POST /organizations`

**Request fields**
- `name`
- `slug` (optional)

## 9.3 Get Organization

`GET /organizations/{organization_id}`

## 9.4 Update Organization

`PATCH /organizations/{organization_id}`

## 9.5 Delete Organization

`DELETE /organizations/{organization_id}`

Soft delete is recommended unless the implementation explicitly requires hard removal.

## 9.6 List Members

`GET /organizations/{organization_id}/members`

## 9.7 Add Member

`POST /organizations/{organization_id}/members`

## 9.8 Remove Member

`DELETE /organizations/{organization_id}/members/{member_id}`

---

## 10. Project Endpoints

## 10.1 List Projects

`GET /projects?organization_id={id}`

## 10.2 Create Project

`POST /projects`

**Request**

```json
{
  "organization_id": "org_123",
  "name": "Production",
  "slug": "production"
}
```

## 10.3 Get Project

`GET /projects/{project_id}`

## 10.4 Update Project

`PATCH /projects/{project_id}`

## 10.5 Delete Project

`DELETE /projects/{project_id}`

---

## 11. Server and Node Endpoints

## 11.1 List Servers

`GET /servers?organization_id={id}`

## 11.2 Create Managed Server Record

`POST /servers`

Creates a server record for either self-hosted, BYO VPS, or managed infrastructure.

**Request fields**
- `organization_id`
- `project_id` (optional)
- `name`
- `mode`
- `provider_type`
- `region` (optional)
- `hostname` (optional)
- `ip_address` (optional)

## 11.3 Get Server

`GET /servers/{server_id}`

## 11.4 Update Server

`PATCH /servers/{server_id}`

## 11.5 Delete Server

`DELETE /servers/{server_id}`

## 11.6 List Node Agents

`GET /servers/{server_id}/agents`

## 11.7 Enroll Node Agent

`POST /servers/{server_id}/agents/enroll`

Used for bootstrap registration.

**Request fields**
- `enrollment_token`
- `agent_version`
- `public_key`
- `host_facts`

## 11.8 Node Heartbeat

`POST /agents/{agent_id}/heartbeat`

**Request fields**
- `status`
- `metrics`
- `timestamp`

## 11.9 Node Health Snapshot

`POST /agents/{agent_id}/health-snapshots`

Used to submit CPU, RAM, disk, and load information.

---

## 12. App Endpoints

## 12.1 List Apps

`GET /apps?project_id={id}&server_id={id}`

## 12.2 Create App

`POST /apps`

**Request fields**
- `organization_id`
- `project_id`
- `server_id`
- `name`
- `source_type` (`git` or `upload`)
- `repository_url` (optional)
- `runtime` (optional if auto-detect)
- `start_command` (optional)
- `working_directory` (optional)

## 12.3 Get App

`GET /apps/{app_id}`

## 12.4 Update App

`PATCH /apps/{app_id}`

## 12.5 Delete App

`DELETE /apps/{app_id}`

## 12.6 Deploy App

`POST /apps/{app_id}/deploy`

Triggers a deployment from Git or an uploaded artifact.

**Request fields**
- `source_ref`
- `deployment_note` (optional)
- `force` (optional)

## 12.7 Start App

`POST /apps/{app_id}/start`

## 12.8 Stop App

`POST /apps/{app_id}/stop`

## 12.9 Restart App

`POST /apps/{app_id}/restart`

## 12.10 Inspect App Status

`GET /apps/{app_id}/status`

## 12.11 App Metrics

`GET /apps/{app_id}/metrics`

---

## 13. Deployment Endpoints

## 13.1 List Deployments

`GET /apps/{app_id}/deployments`

## 13.2 Get Deployment

`GET /deployments/{deployment_id}`

## 13.3 Rollback Deployment

`POST /deployments/{deployment_id}/rollback`

Rollback should only be enabled if a safe rollback target exists.

## 13.4 Redeploy Latest

`POST /apps/{app_id}/deployments/redeploy-latest`

---

## 14. Environment Variable Endpoints

## 14.1 List Variables

`GET /apps/{app_id}/env`

## 14.2 Add Variable

`POST /apps/{app_id}/env`

**Request fields**
- `key`
- `value`
- `is_secret`

## 14.3 Update Variable

`PATCH /apps/{app_id}/env/{env_id}`

## 14.4 Delete Variable

`DELETE /apps/{app_id}/env/{env_id}`

### Secret Handling Rule

Secret values must be masked in responses after creation.

---

## 15. Domain and SSL Endpoints

## 15.1 List Domains

`GET /domains?organization_id={id}`

## 15.2 Create Domain

`POST /domains`

**Request fields**
- `organization_id`
- `project_id`
- `name`

## 15.3 Get Domain

`GET /domains/{domain_id}`

## 15.4 Update Domain

`PATCH /domains/{domain_id}`

## 15.5 Delete Domain

`DELETE /domains/{domain_id}`

## 15.6 Bind Domain to App

`POST /domains/{domain_id}/bind`

**Request fields**
- `app_id`
- `path_prefix` (optional)
- `proxy_mode` (optional)

## 15.7 Unbind Domain

`POST /domains/{domain_id}/unbind`

## 15.8 Issue SSL Certificate

`POST /domains/{domain_id}/ssl/issue`

## 15.9 Renew SSL Certificate

`POST /domains/{domain_id}/ssl/renew`

## 15.10 SSL Status

`GET /domains/{domain_id}/ssl`

---

## 16. Database Endpoints

## 16.1 List Databases

`GET /databases?project_id={id}&server_id={id}`

## 16.2 Create Database

`POST /databases`

**Request fields**
- `organization_id`
- `project_id`
- `server_id`
- `name`
- `engine` (`mysql`, `postgresql`, `redis`)
- `version` (optional)

## 16.3 Get Database

`GET /databases/{database_id}`

## 16.4 Update Database

`PATCH /databases/{database_id}`

## 16.5 Delete Database

`DELETE /databases/{database_id}`

## 16.6 Database Credentials

`GET /databases/{database_id}/credentials`

## 16.7 Rotate Database Password

`POST /databases/{database_id}/credentials/rotate`

## 16.8 Manual Backup

`POST /databases/{database_id}/backups`

## 16.9 List Backups

`GET /databases/{database_id}/backups`

## 16.10 Restore Backup

`POST /backups/{backup_id}/restore`

Restore must require confirmation and must be treated as a sensitive operation.

---

## 17. Logs Endpoints

## 17.1 Get App Logs

`GET /apps/{app_id}/logs`

Supports query parameters such as:
- `level`
- `from`
- `to`
- `cursor`
- `limit`

## 17.2 Get Server Logs

`GET /servers/{server_id}/logs`

## 17.3 Live Log Stream

`GET /ws/logs?app_id={id}`

WebSocket channel for real-time log consumption.

---

## 18. AI Endpoints

## 18.1 Ask AI

`POST /ai/chat`

Used for natural-language server questions.

**Request fields**
- `organization_id`
- `scope_type`
- `scope_id`
- `message`

## 18.2 AI Recommendations

`GET /ai/recommendations?organization_id={id}`

## 18.3 AI Actions

`GET /ai/actions?organization_id={id}`

## 18.4 AI Action Detail

`GET /ai/actions/{ai_action_id}`

## 18.5 Approve AI Action

`POST /ai/actions/{ai_action_id}/approve`

## 18.6 Reject AI Action

`POST /ai/actions/{ai_action_id}/reject`

## 18.7 Execute AI Action

`POST /ai/actions/{ai_action_id}/execute`

Execution is allowed only if policy and whitelist checks pass.

---

## 19. Policy Endpoints

## 19.1 Get Policy

`GET /policies/{policy_id}`

## 19.2 List Policies

`GET /policies?organization_id={id}`

## 19.3 Update Policy

`PATCH /policies/{policy_id}`

### Policy Fields

- `name`
- `mode`
- `is_active`
- `allowed_actions`
- `approval_required_actions`
- `notification_rules`

---

## 20. Audit Endpoints

## 20.1 List Audit Events

`GET /audit-events?organization_id={id}`

## 20.2 Get Audit Event

`GET /audit-events/{audit_event_id}`

Audit events should be immutable and read-only through the API.

---

## 21. Billing Endpoints

These are most relevant in cloud mode.

## 21.1 Get Billing Account

`GET /billing/account`

## 21.2 Update Billing Details

`PATCH /billing/account`

## 21.3 List Subscriptions

`GET /billing/subscriptions`

## 21.4 List Invoices

`GET /billing/invoices`

## 21.5 List Usage Records

`GET /billing/usage`

---

## 22. Provisioning Endpoints

## 22.1 List Jobs

`GET /provisioning/jobs?organization_id={id}`

## 22.2 Create Job

`POST /provisioning/jobs`

Used for managed infrastructure provisioning or related orchestration tasks.

## 22.3 Get Job

`GET /provisioning/jobs/{job_id}`

## 22.4 Cancel Job

`POST /provisioning/jobs/{job_id}/cancel`

Cancellation should be best-effort and state-aware.

---

## 23. Node Agent API Contract

The agent API is a restricted subset intended for trusted node communication.

## 23.1 Agent Handshake

`POST /agent/handshake`

Used to establish identity and session context.

**Request fields**
- `node_id`
- `public_key`
- `agent_version`
- `enrollment_token` (bootstrap only)

## 23.2 Agent Command Receive

`POST /agent/commands`

Receives signed or authorized commands from the control plane.

## 23.3 Agent Command Status

`GET /agent/commands/{command_id}`

## 23.4 Agent Publish Event

`POST /agent/events`

## 23.5 Agent Emit Logs

`POST /agent/logs`

## 23.6 Agent Health Ping

`POST /agent/health`

### Agent Safety Rule

The agent must reject any command that is not in the allowed action registry.

---

## 24. Standard Resource Shapes

## 24.1 App Object

```json
{
  "id": "app_123",
  "organization_id": "org_123",
  "project_id": "proj_123",
  "server_id": "srv_123",
  "name": "web",
  "slug": "web",
  "source_type": "git",
  "repository_url": "https://github.com/example/repo",
  "runtime": "nodejs",
  "status": "running",
  "created_at": "2026-06-07T12:00:00Z",
  "updated_at": "2026-06-07T12:00:00Z"
}
```

## 24.2 Server Object

```json
{
  "id": "srv_123",
  "organization_id": "org_123",
  "name": "production-1",
  "mode": "managed",
  "provider_type": "hetzner",
  "region": "sgp1",
  "status": "online",
  "last_seen_at": "2026-06-07T12:00:00Z"
}
```

## 24.3 AI Action Object

```json
{
  "id": "ai_123",
  "organization_id": "org_123",
  "action_key": "restart_app",
  "action_status": "pending_approval",
  "app_id": "app_123",
  "server_id": "srv_123",
  "result_summary": null,
  "created_at": "2026-06-07T12:00:00Z"
}
```

---

## 25. Error Handling Rules

1. Validation errors should return field-level details when safe.
2. Authorization errors must not leak whether the resource exists outside the caller’s scope.
3. Node failures should return actionable status codes when possible.
4. AI policy denials should explain the policy result without exposing sensitive internals.
5. Internal errors should include a request ID for debugging.

---

## 26. Rate Limiting and Abuse Controls

Recommended to apply rate limiting to:
- login attempts,
- node enrollment,
- provisioning requests,
- AI chat calls,
- destructive or high-risk actions.

Cloud mode should also consider per-organization quotas.

---

## 27. Idempotency

High-risk or retried operations should support idempotency where relevant.

Recommended idempotent areas:
- provisioning jobs,
- node enrollment,
- deploy actions,
- certificate issuance requests,
- AI action execution requests.

Use an `Idempotency-Key` header for requests that may be retried.

---

## 28. Events and WebSocket Payloads

WebSocket events should be small, typed, and consistent.

Example event:

```json
{
  "type": "app.log",
  "timestamp": "2026-06-07T12:00:00Z",
  "data": {
    "app_id": "app_123",
    "level": "error",
    "message": "Process crashed"
  }
}
```

---

## 29. Versioning Strategy

- `v1` is stable for initial implementation.
- Breaking changes require a major version path.
- New fields may be added in a backward-compatible manner.
- Endpoints may be deprecated but should remain supported for a defined window.

---

## 30. Recommended Endpoint Priority for MVP

The minimum set to ship first should be:

- auth
- organizations
- projects
- servers
- apps
- app deployments
- environment variables
- domains
- SSL issuance
- database provisioning
- logs
- audit events
- AI chat and AI actions
- node handshake and heartbeat

Billing and provisioning orchestration can be staged if the initial release remains self-hosted-first.

---

## 31. Implementation Notes

- Prefer explicit handler names aligned with domain nouns.
- Keep node endpoints separate from public UI endpoints.
- Do not let frontend convenience leak into authorization design.
- Treat AI action execution as a privileged workflow, not a normal API call.
- Keep response envelopes predictable for frontend state management.

---

## 32. Open Questions

- Should node APIs live under the same base path as UI APIs or a separate service namespace?
- Should action execution use synchronous responses or async job tracking by default?
- What should the first cloud-mode token format be?
- Should log streaming be per app, per server, or both from the start?
- Which endpoints should be strictly local-only in self-hosted mode?

---

## 33. API Definition of Done

The API specification is complete enough for implementation when:

1. Every major resource has clear CRUD or command endpoints.
2. Auth and authorization behavior is defined by mode.
3. Node communication is separated from public access.
4. AI actions are routed through policy gates.
5. Error format and response envelopes are standardized.
6. Audit and logs are covered.
7. The API can support both self-hosted and cloud-managed operation without semantic drift.

---

## 34. Closing Statement

Razad’s API should be predictable, strict, and safe.

It must expose just enough surface area to build a powerful product, while keeping the dangerous parts gated behind policy, scope, and explicit privileges.

