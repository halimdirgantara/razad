# Razad
## Architecture Decision Records (ADR) v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Capture major technical decisions and the rationale behind them.

---

## 1. Purpose

This document records the core architectural decisions that shape Razad. The goal is to keep future implementation consistent and to prevent repeated debate on foundational choices.

---

## 2. Decision 001 — Go for the Core Daemon

### Decision
Use Go for the backend daemon, orchestration layer, and local execution plane.

### Context
Razad must run as a lightweight, reliable Linux-native service. It needs strong concurrency support, easy static compilation, and low runtime overhead.

### Rationale
- Single binary deployment is practical.
- Strong standard library for HTTP, crypto, and concurrency.
- Good fit for system-level tooling.
- Stable production footprint.

### Consequences
- Backend code must follow Go idioms.
- Shared domain logic should remain explicit and testable.
- Frontend and backend should be separated cleanly.

---

## 3. Decision 002 — SvelteKit for the Frontend

### Decision
Use SvelteKit for the dashboard and operational UI.

### Rationale
- Fast UI development.
- Good fit for dashboard-style interfaces.
- Easy to produce a static build for embedding.
- Strong developer experience.

### Consequences
- UI should remain thin and API-driven.
- State should be synchronized with backend as source of truth.

---

## 4. Decision 003 — systemd as Process Supervisor

### Decision
Use systemd as the authoritative process supervisor for applications and node services.

### Rationale
- Native to Linux.
- Reliable restart and boot integration.
- Reduces the need for a custom process manager.
- Fits the product philosophy of native-first infrastructure.

### Consequences
- Razad must generate and validate unit files.
- Process management code must respect systemd semantics.
- Debugging should integrate with journal and service state.

---

## 5. Decision 004 — Nginx for Reverse Proxy

### Decision
Use Nginx to route public traffic to applications.

### Rationale
- Mature and widely deployed.
- Easy TLS termination and reverse proxy support.
- Works well with systemd-managed backends.

### Consequences
- Razad must generate valid Nginx configs.
- Config validation must happen before reload.
- The product should preserve a last-known-good proxy configuration.

---

## 6. Decision 005 — PostgreSQL as Primary Metadata Store

### Decision
Use PostgreSQL for control-plane and metadata storage.

### Rationale
- Strong relational integrity.
- Good support for multi-tenant application data.
- Good migration and indexing support.
- Mature ecosystem for production systems.

### Consequences
- Schema design must be explicit and normalized.
- Migration scripts must be versioned and reviewed.
- Secrets must never be stored in plaintext.

---

## 7. Decision 006 — SQLite Optional for Lightweight Self-Hosted Use

### Decision
Allow SQLite as an optional storage backend for limited self-hosted deployments where practical.

### Rationale
- Improves accessibility for small installations.
- Reduces setup friction for local-first use.

### Consequences
- Not all advanced features will be equally strong on SQLite.
- Core schema and abstractions must still be relational-friendly.

---

## 8. Decision 007 — AI Must Be Whitelisted and Non-Destructive

### Decision
Razad AI may only execute explicitly allowed non-destructive actions.

### Rationale
- The product touches live production systems.
- Unbounded AI action is a high-risk security failure.
- Safety must be enforced in code, not just in prompts.

### Consequences
- Action registry must be explicit.
- Policy engine must gate execution.
- Audit logs must record all AI activity.

---

## 9. Decision 008 — Control Plane and Execution Plane Separation

### Decision
Separate cloud control-plane concerns from node-local execution concerns.

### Rationale
- Enables future cloud scale.
- Reduces coupling.
- Makes BYO VPS and managed infrastructure possible without rewrite.

### Consequences
- Agent protocol must be clearly defined.
- Cloud must not directly mutate node state without agent mediation.

---

## 10. Decision 009 — AGPLv3 for the Open Source Core

### Decision
Use AGPLv3 for Razad OSS.

### Rationale
- Protects against closed-source SaaS forks of the platform.
- Preserves open-source reciprocity for networked use.

### Consequences
- Commercial strategy must be compatible with AGPLv3.
- Documentation must be explicit about licensing expectations.

---

## 11. Decision Log Maintenance

New decisions should be numbered and appended as the architecture evolves. Each decision must state:
- the decision,
- the context,
- the rationale,
- the consequences,
- the date or version context.

---

## 12. Summary

Razad is optimized for:
- native Linux execution,
- lightweight deployment,
- safe automation,
- and future cloud expansion.

These decisions should remain stable unless the product direction materially changes.
