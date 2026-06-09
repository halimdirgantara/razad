# Razad
## Git Workflow & Branching Strategy v1.0

**Product Name:** Razad  
**Document Version:** 1.0  
**Status:** Draft  
**Purpose:** Define source control rules for a disciplined platform engineering workflow.

---

## 1. Purpose

Razad is a platform product with multiple architectural layers. The git workflow must support careful review, controlled releases, and low-risk integration.

---

## 2. Branch Model

Recommended long-lived branches:

- `main` — production-ready code
- `develop` — integration branch if a staged release flow is preferred

Recommended short-lived branches:
- `feature/<name>`
- `fix/<name>`
- `chore/<name>`
- `hotfix/<name>`
- `refactor/<name>`

If the team is small, a trunk-based flow can be used with short-lived feature branches and frequent merges to `main`.

---

## 3. Branching Rules

- No direct commits to `main`.
- Every change must go through review.
- Each branch should target a single concern.
- Large features should be broken into smaller mergeable units.
- Rebase or merge strategy must be consistent within the team.

---

## 4. Commit Message Convention

Recommended style:

```text
type(scope): summary
```

Examples:
- `feat(app): add deployment start flow`
- `fix(proxy): validate nginx config before reload`
- `chore(ci): add build pipeline`
- `refactor(auth): isolate session middleware`

### Allowed Types
- `feat`
- `fix`
- `refactor`
- `chore`
- `docs`
- `test`
- `ci`
- `perf`
- `sec`

---

## 5. Pull Request Rules

Every pull request should include:
- clear description,
- linked task or milestone,
- screenshots if UI is affected,
- testing notes,
- rollback considerations if risky.

Pull requests must not mix unrelated work.

---

## 6. Merge Criteria

A branch may merge only when:
- the code builds,
- tests pass,
- behavior is documented,
- security-sensitive changes are reviewed,
- migration changes are safe,
- no scope creep is introduced.

---

## 7. Release Strategy

Recommended release stages:
- feature branch
- integration review
- staging validation
- release candidate
- production release
- post-release monitoring

---

## 8. Hotfix Strategy

Hotfixes must:
- address only one production issue,
- avoid unrelated refactors,
- be merged back to the active development branch,
- be tagged clearly in release notes.

---

## 9. Version Tagging

Use semantic versioning:
- `v0.x` for early platform development
- `v1.0.0` when MVP is release-ready
- patch versions for fixes

---

## 10. Repository Hygiene

- Keep generated artifacts out of source control.
- Keep migrations in order.
- Keep docs in sync with code.
- Keep feature flags explicit.
- Avoid force-pushing shared branches unless absolutely necessary.

---

## 11. Recommended Team Discipline

- small pull requests
- frequent integration
- no long-lived divergence
- review before merge
- release notes for every meaningful change

---

## 12. Summary

Razad benefits from a workflow that is:
- simple,
- reviewable,
- traceable,
- and safe for a system that touches live servers.
