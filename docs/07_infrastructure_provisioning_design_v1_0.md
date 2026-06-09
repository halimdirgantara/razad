# Razad
## Infrastructure Provisioning Design v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define how Razad Cloud provisions, enrolls, and manages infrastructure for customers who do not own a VPS.

---

## 1. Purpose

This document defines the provisioning design for Razad Cloud when the customer does not have their own VPS.

The goal is to support managed infrastructure while keeping the architecture aligned with the rest of Razad:
- lightweight where possible,
- secure by default,
- tenant-aware,
- and operationally clear.

---

## 2. Provisioning Models

### 2.1 BYO VPS
The customer owns the VPS. Razad Cloud manages it through node enrollment.

### 2.2 Managed VPS
Razad Cloud provisions a VPS from a provider and attaches it to the customer account.

### 2.3 Managed Node Pool
Razad operates a pool of nodes and assigns capacity to tenants.

For early implementation, the recommended path is:
1. BYO VPS
2. Managed VPS
3. Node pool expansion later

---

## 3. Provisioning Goals

- create infrastructure with minimal manual steps,
- bind a provisioned node to a tenant safely,
- support predictable node lifecycle,
- keep provider integration isolated behind adapters,
- avoid direct coupling between cloud UI and node internals.

---

## 4. Provisioning Lifecycle

1. user requests server creation,
2. billing or entitlement is checked,
3. provider selection is resolved,
4. node provisioning job is created,
5. provider API creates the VPS,
6. bootstrap config is generated,
7. agent is installed or bootstrapped,
8. node enrolls to control plane,
9. node becomes active,
10. user sees the server in the dashboard.

---

## 5. Provider Adapter Model

Each provider integration should implement a common adapter interface.

Responsibilities:
- create server
- delete server
- fetch server status
- obtain network info
- optionally resize or reimage
- report provisioning failure details

Provider adapters should be isolated from core business logic.

---

## 6. Node Bootstrap Design

Bootstrap should:
- install the agent,
- register the node identity,
- exchange enrollment token for session identity,
- fetch initial policy,
- begin heartbeat reporting,
- expose health and logs.

Bootstrap must fail closed if identity validation cannot be completed.

---

## 7. Tenant Isolation Model

Managed infrastructure must protect tenants from each other.

Recommended early-stage isolation:
- separate Linux user per tenant or app set,
- dedicated service identities,
- isolated filesystem paths,
- scoped credentials,
- strict organization/project ownership.

Later isolation can evolve to stronger boundary mechanisms if needed.

---

## 8. State Machine

### Provisioning States
- queued
- validating
- provisioning
- bootstrapping
- enrolling
- active
- failed
- canceled
- deleted

### Node States
- offline
- online
- degraded
- revoked

### Rules
- A node should not be marked active until the agent has enrolled successfully.
- A failed provisioning job must preserve failure reason.
- A deleted node should be detached safely and audited.

---

## 9. Failure Handling

Possible failures:
- provider API unavailable
- quota exceeded
- bootstrap script failure
- agent enrollment failure
- network timeout
- permission mismatch
- billing denial

Recovery:
- surface the exact failure stage,
- keep job history,
- allow retry only if safe,
- avoid creating orphan resources silently.

---

## 10. Billing and Entitlement Checks

Before provisioning:
- confirm account entitlement,
- confirm quota,
- confirm provider capacity,
- confirm plan supports the requested size.

Billing logic should not be embedded inside provider adapters.

---

## 11. Summary

Infrastructure provisioning must be an extension of Razad, not a separate product hidden behind the same UI.

The first version should be thin, controlled, and easy to reason about.
