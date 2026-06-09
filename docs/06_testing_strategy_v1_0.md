# Razad
## Testing Strategy v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define how Razad is tested across unit, integration, system, and failure-mode levels.

---

## 1. Purpose

Razad affects live server operations. Testing must prove that the product behaves correctly under normal, edge, and failure conditions.

---

## 2. Testing Goals

- verify core workflows,
- catch regressions early,
- protect security boundaries,
- validate Linux integration,
- ensure config generation is safe,
- prove AI cannot perform forbidden actions,
- keep release risk low.

---

## 3. Test Pyramid

### 3.1 Unit Tests
Focus on:
- service logic
- validation
- policy rules
- config parsing
- utility functions
- error conversion

### 3.2 Integration Tests
Focus on:
- database interaction
- API endpoints
- migration correctness
- systemd config generation logic
- proxy config generation
- AI action gating

### 3.3 End-to-End Tests
Focus on:
- install -> login -> deploy -> domain -> SSL -> logs -> audit
- restart after crash
- AI suggestion and approval flow
- database provisioning

---

## 4. Critical Test Scenarios

### Installation
- fresh install succeeds
- repeat install is safe
- unsupported OS is rejected
- permissions failure is reported clearly

### Authentication
- valid login succeeds
- invalid login fails
- session expiration works
- logout invalidates session

### App Lifecycle
- create app from Git
- deploy app
- stop app
- restart app
- delete app
- deployment failure is visible

### Proxy and SSL
- domain binding succeeds
- Nginx config validates
- SSL issues successfully
- invalid config does not overwrite working config

### Database
- PostgreSQL provisioning works
- credentials are created safely
- backup can be generated
- restore path is gated and validated

### AI Safety
- allowed action succeeds
- disallowed action is blocked
- AI audit event is recorded
- AI cannot execute arbitrary commands

### Observability
- logs stream in real time
- health snapshot appears
- node offline state is visible
- audit trail is queryable

---

## 5. Negative Testing

Negative tests are mandatory for:
- unauthorized access,
- invalid input,
- malformed config,
- forbidden AI action,
- token replay,
- bad domain binding,
- broken Nginx template,
- failed certificate issuance,
- corrupted secrets.

---

## 6. Security Testing

Security-sensitive tests must confirm:
- secrets are not returned in plaintext,
- logs are redacted,
- AI action allowlist is enforced,
- cross-tenant access is denied,
- privileged endpoints require auth,
- audit events are created.

---

## 7. Environment Strategy

### Local
- fast developer feedback
- mocked integrations where practical

### Staging
- real systemd/Nginx/DB behavior where possible
- production-like configuration

### Production-like Validation
- full install flow
- real app deployment
- SSL issuance
- live log streaming
- recovery drills

---

## 8. Regression Strategy

Any bug fixed in a production path must receive:
- a regression test,
- a documented root cause,
- a note about the affected module.

---

## 9. Release Criteria

A build may be released only if:
- core tests pass,
- security tests pass,
- critical integration flows pass,
- no breaking schema migration is pending,
- known failures are documented and acceptable.

---

## 10. Summary

Razad testing must prove not just that the product works, but that it fails safely.
