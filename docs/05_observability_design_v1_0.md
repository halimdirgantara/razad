# Razad
## Observability Design v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define logs, metrics, traces, events, and operational visibility for Razad.

---

## 1. Purpose

Razad manages production servers. Observability is not optional. Users must be able to understand what is running, what is failing, and what changed.

---

## 2. Observability Goals

- Show server health at a glance.
- Expose app and service logs in real time.
- Track deployment and recovery history.
- Record audit events for privileged actions.
- Provide enough telemetry for AI recommendations and safe automation.

---

## 3. Observability Layers

### 3.1 Logs
- app stdout/stderr
- systemd journal output
- Nginx logs
- database service logs
- audit logs

### 3.2 Metrics
- CPU
- RAM
- disk usage
- load average
- process state
- service state
- restart frequency

### 3.3 Events
- app created
- deployment started
- deployment failed
- domain bound
- SSL issued
- AI action executed
- node heartbeat missed

### 3.4 Audit Trail
- user actions
- AI actions
- provisioning actions
- security-sensitive events

---

## 4. UX Requirements

The dashboard should present:
- current server status,
- app health,
- active alerts,
- recent logs,
- recent deployments,
- recent AI activity.

The user should not need SSH for routine diagnosis.

---

## 5. Log Streaming Design

### Requirements
- stream logs over WebSocket,
- support app-level filtering,
- support server-level views,
- keep stream cursors,
- avoid unbounded memory growth.

### Behavior
- If a stream disconnects, the UI should reconnect.
- If logs overflow, the system should preserve recent entries and expose truncation state.
- Sensitive values must be redacted before display.

---

## 6. Metrics Collection Design

### Sources
- node agent
- local process inspection
- systemd status
- service-specific probes

### Capture Strategy
- periodic snapshot collection
- lightweight enough for low-spec VPS
- store only what is necessary for operational insight

---

## 7. Health Model

### States
- healthy
- warning
- critical
- offline
- unknown

### Rules
- A crashed app should surface visibly.
- A missing heartbeat should degrade server status.
- An SSL certificate nearing expiry should generate a warning.
- A failed deployment should be visible as a state transition, not hidden in logs alone.

---

## 8. Audit Observability

Audit records should answer:
- who did it,
- what happened,
- when it happened,
- what was affected,
- whether AI was involved,
- whether the action succeeded.

Audit logs should be searchable and immutable from the user perspective.

---

## 9. Retention Strategy

### High-frequency data
- Keep recent detail.
- Roll up or prune old data where appropriate.

### Audit data
- Retain longer.
- Avoid destructive pruning.

### Logs
- Retention should be configurable.
- Default should be practical for small VPS systems.

---

## 10. AI Support

AI should read:
- logs,
- metrics,
- deployment history,
- health snapshots,
- audit trail.

AI should not be able to bypass observability policy. It only consumes the data exposed by the system.

---

## 11. Alerting Strategy

The MVP alerting strategy should support:
- app crash,
- node offline,
- SSL expiring soon,
- deployment failure,
- database service failure.

Alert noise should be controlled through debouncing and severity thresholds.

---

## 12. Summary

Observability in Razad must be:
- actionable,
- low overhead,
- secure,
- and useful enough to diagnose production problems quickly.
