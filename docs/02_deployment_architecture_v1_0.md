# Razad
## Deployment Architecture v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define how Razad is deployed in self-hosted, BYO VPS cloud, and managed infrastructure modes.

---

## 1. Purpose

This document defines the deployment architecture for Razad, including runtime placement, installation flow, environment layout, and operational boundaries.

---

## 2. Deployment Modes

### 2.1 Self-Hosted Mode
Razad runs on a customer-owned Linux server as a single installation.

Characteristics:
- One daemon instance
- Embedded UI assets
- Local config and secrets
- systemd service integration
- No dependence on Razad Cloud

### 2.2 BYO VPS Cloud Mode
The customer owns the VPS, but Razad Cloud manages it remotely.

Characteristics:
- Node agent installed on customer server
- Control plane in Razad Cloud
- Remote orchestration over secure agent protocol

### 2.3 Managed Infrastructure Mode
Razad provisions and operates infrastructure on behalf of the customer.

Characteristics:
- Customer uses Razad Cloud UI
- Infrastructure is provisioned through provider APIs
- Agent bootstrap is automated
- Same execution model as BYO VPS

---

## 3. Runtime Topology

### Self-Hosted
```text
Browser -> Razad daemon -> systemd/Nginx/DB on same host
```

### Cloud + BYO VPS
```text
Browser -> Razad Cloud -> Node Agent -> Customer VPS services
```

### Cloud + Managed
```text
Browser -> Razad Cloud -> Provisioner -> Managed Node Agent -> Node services
```

---

## 4. Installation Layout

Recommended directories on Linux host:

- `/etc/razad/` for config
- `/var/lib/razad/` for state
- `/var/log/razad/` for logs
- `/opt/razad/` for binaries or packaged assets
- `/etc/systemd/system/razad.service` for daemon registration

---

## 5. Self-Hosted Start Flow

1. Install binary and configs.
2. Create or bootstrap admin.
3. Register systemd service.
4. Start daemon.
5. Load dashboard.
6. Create first app.
7. Generate service and proxy configs.
8. Start app.
9. Expose health and logs.

---

## 6. Cloud-Controlled Node Flow

1. User creates server in cloud.
2. Cloud issues enrollment token.
3. Agent is installed or bootstrapped on node.
4. Agent enrolls outbound.
5. Cloud stores node identity.
6. Cloud pushes commands only through the agent.
7. Node performs local actions and returns status.

---

## 7. Release Artifacts

The deployment package should include:
- install script
- daemon binary
- config template
- systemd unit template
- Nginx template
- migration tool
- optional frontend build assets

---

## 8. Environment Separation

### Development
- local config
- debug logging
- no production secrets
- mock provider keys where needed

### Staging
- realistic config
- real integrations where safe
- limited test data

### Production
- encrypted secrets
- strict audit logging
- TLS required
- config validation before apply

---

## 9. Reverse Proxy Flow

The proxy layer must be configured only after:
- app runtime is known,
- backend process is healthy,
- domain ownership is verified,
- config template is validated.

---

## 10. Deployment Safety Rules

- Never overwrite a known-good config without validation.
- Never expose secrets in logs.
- Never allow cloud services to bypass node authorization.
- Never let AI trigger unsupported commands.
- Never deploy partially initialized state as if it were healthy.

---

## 11. Rollback Strategy

If a deployment fails:
- preserve the last known working version,
- keep prior Nginx config available,
- keep prior systemd service definition if needed,
- mark deployment as failed,
- present actionable diagnostics to the user.

---

## 12. Summary

Razad deployment is designed to be:
- deterministic,
- Linux-native,
- safe by default,
- and expandable into cloud-managed operation.
