# Razad
## Development Standards v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define coding, structure, naming, testing, and operational standards for the codebase.

---

## 1. Purpose

This document defines standards that keep the Razad codebase consistent as it grows.

---

## 2. Core Standards

### 2.1 Readability First
- Prefer explicit code over clever code.
- Prefer small functions over large ones.
- Prefer clear names over abbreviated names.

### 2.2 Module Boundaries
- Each domain module owns its own service, repository, and handler logic.
- Shared logic belongs in shared packages only when truly reusable.

### 2.3 Single Responsibility
- One function should do one thing well.
- One module should own one business concern.
- UI should not contain backend business logic.

---

## 3. Go Standards

- Use idiomatic Go formatting.
- Keep package names short and meaningful.
- Avoid cyclic imports.
- Keep handlers thin.
- Put business logic in services.
- Put persistence logic in repositories.
- Return typed errors where practical.

### Error Handling
- Wrap errors with context.
- Do not swallow errors silently.
- Convert internal errors into stable public error responses.

---

## 4. Frontend Standards

- Use SvelteKit patterns consistently.
- Keep API calls isolated in a client layer.
- Use components for repeated UI structures.
- Keep screens operationally focused.
- Surface errors clearly.
- Avoid deep state mutation in components.

---

## 5. API Standards

- Use versioned routes.
- Use JSON consistently.
- Return stable error envelopes.
- Make destructive actions explicit.
- Require authentication for privileged endpoints.

---

## 6. Naming Standards

### Files
- Use lowercase with hyphens or domain-consistent naming where applicable.

### Functions
- Use verbs for actions.
- Use nouns for data retrieval methods.
- Keep names aligned to the domain language.

### Database
- Table names should be plural and stable.
- Column names should be consistent and predictable.

---

## 7. Testing Standards

- Every module should have at least basic unit coverage.
- Critical flows require integration tests.
- Security-sensitive logic requires negative tests.
- API handlers should be tested with request/response cases.
- Migration changes should be reviewed with schema intent.

---

## 8. Documentation Standards

- Update docs when behavior changes.
- Keep architecture docs aligned with implementation.
- Document every public endpoint and significant module boundary.
- Keep operator-facing instructions practical.

---

## 9. Security Standards

- Never log secrets.
- Never hardcode provider keys.
- Never execute arbitrary shell commands from AI.
- Never expose raw internal errors to the UI without sanitization.
- Validate every user-supplied value that influences system commands or config files.

---

## 10. Performance Standards

- Prefer efficient queries.
- Avoid unnecessary round-trips.
- Keep log streaming bounded.
- Keep config reloads safe and minimal.
- Do not optimize prematurely unless a real bottleneck exists.

---

## 11. Review Standards

Before merge:
- verify code style,
- verify tests,
- check security impact,
- verify docs,
- review error handling,
- check migration safety.

---

## 12. Summary

The codebase should remain:
- explicit,
- testable,
- predictable,
- secure,
- and easy to hand off.
