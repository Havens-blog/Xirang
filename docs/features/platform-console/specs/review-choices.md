---
feature: "platform-console"
reviewed: "2026-09-30"
---

# Review Choices

> Non-interactive mode: all `[CROSS]` items auto-approved for integration; all `[LOCAL]` items skipped.
> Overlap policy: `[skip]` keep both (no existing project-level spec files yet — `docs/business-rules/` and `docs/conventions/` did not exist prior to this run, so no overlaps to resolve).

## Approved for Integration

### Business rules -> docs/business-rules/

- BIZ-001 -> docs/business-rules/auth.md (as BIZ-auth-001)
- BIZ-002 -> docs/business-rules/auth.md (as BIZ-auth-002)
- BIZ-003 -> docs/business-rules/auth.md (as BIZ-auth-003)
- BIZ-004 -> docs/business-rules/auth.md (as BIZ-auth-004)
- BIZ-005 -> docs/business-rules/auth.md (as BIZ-auth-005)
- BIZ-006 -> docs/business-rules/auth.md (as BIZ-auth-006)
- BIZ-007 -> docs/business-rules/auth.md (as BIZ-auth-007)
- BIZ-008 -> docs/business-rules/auth.md (as BIZ-auth-008)
- BIZ-009 -> docs/business-rules/identity.md (as BIZ-identity-001)
- BIZ-010 -> docs/business-rules/identity.md (as BIZ-identity-002)
- BIZ-011 -> docs/business-rules/identity.md (as BIZ-identity-003)
- BIZ-012 -> docs/business-rules/identity.md (as BIZ-identity-004)
- BIZ-013 -> docs/business-rules/identity.md (as BIZ-identity-005)
- BIZ-014 -> docs/business-rules/resilience.md (as BIZ-resilience-001)
- BIZ-015 -> docs/business-rules/resilience.md (as BIZ-resilience-002)
- BIZ-017 -> docs/business-rules/auth.md (as BIZ-auth-009)
- BIZ-018 -> docs/business-rules/auth.md (as BIZ-auth-010)

### Technical specs -> docs/conventions/

- TECH-001 -> docs/conventions/api.md (as TECH-api-001)
- TECH-002 -> docs/conventions/error-handling.md (as TECH-error-001)
- TECH-003 -> docs/conventions/api.md (as TECH-api-002)
- TECH-004 -> docs/conventions/api.md (as TECH-api-003)
- TECH-006 -> docs/conventions/api.md (as TECH-api-004)
- TECH-007 -> docs/conventions/error-handling.md (as TECH-error-002)
- TECH-008 -> docs/conventions/error-handling.md (as TECH-error-003)
- TECH-009 -> docs/conventions/error-handling.md (as TECH-error-004)
- TECH-010 -> docs/conventions/error-handling.md (as TECH-error-005)
- TECH-011 -> docs/conventions/error-handling.md (as TECH-error-006)
- TECH-012 -> docs/conventions/error-handling.md (as TECH-error-007)
- TECH-013 -> docs/conventions/frontend-state.md (as TECH-frontend-state-001)
- TECH-014 -> docs/conventions/frontend-state.md (as TECH-frontend-state-002)
- TECH-015 -> docs/conventions/frontend-state.md (as TECH-frontend-state-003)
- TECH-016 -> docs/conventions/frontend-state.md (as TECH-frontend-state-004)
- TECH-017 -> docs/conventions/frontend-state.md (as TECH-frontend-state-005)
- TECH-019 -> docs/conventions/data-model.md (as TECH-data-model-001)
- TECH-020 -> docs/conventions/data-model.md (as TECH-data-model-002)
- TECH-021 -> docs/conventions/data-model.md (as TECH-data-model-003)
- TECH-022 -> docs/conventions/data-model.md (as TECH-data-model-004)
- TECH-023 -> docs/conventions/data-model.md (as TECH-data-model-005)
- TECH-024 -> docs/conventions/data-model.md (as TECH-data-model-006)
- TECH-025 -> docs/conventions/security.md (as TECH-security-001)
- TECH-026 -> docs/conventions/security.md (as TECH-security-002)
- TECH-027 -> docs/conventions/security.md (as TECH-security-003)
- TECH-028 -> docs/conventions/security.md (as TECH-security-004)
- TECH-029 -> docs/conventions/performance.md (as TECH-performance-001)
- TECH-030 -> docs/conventions/performance.md (as TECH-performance-002)
- TECH-031 -> docs/conventions/performance.md (as TECH-performance-003)
- TECH-032 -> docs/conventions/testing.md (as TECH-testing-001)
- TECH-033 -> docs/conventions/testing.md (as TECH-testing-002)
- TECH-034 -> docs/conventions/testing.md (as TECH-testing-003)
- TECH-035 -> docs/conventions/testing.md (as TECH-testing-004)
- TECH-036 -> docs/conventions/deployment.md (as TECH-deployment-001)
- TECH-037 -> docs/conventions/deployment.md (as TECH-deployment-002)
- TECH-038 -> docs/conventions/cross-repo-copy.md (as TECH-cross-repo-copy-001)
- TECH-039 -> docs/conventions/cross-repo-copy.md (as TECH-cross-repo-copy-002)
- TECH-040 -> docs/conventions/cross-repo-copy.md (as TECH-cross-repo-copy-003)
- TECH-041 -> docs/conventions/tech-stack.md (as TECH-tech-stack-001)
- TECH-042 -> docs/conventions/tech-stack.md (as TECH-tech-stack-002)

## Skipped (LOCAL)

- BIZ-016 (零/单/多租户登录分流) — stays in feature
- TECH-005 (eiam API 契约基线登记入 parity-checklist) — stays in feature
- TECH-018 (写前守卫保证范围仅控制台 7 个 list store) — stays in feature

## Related Existing Entries

- (none — `docs/business-rules/` and `docs/conventions/` did not exist prior to this run)
