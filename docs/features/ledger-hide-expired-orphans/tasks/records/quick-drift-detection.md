---
status: "completed"
started: "2026-09-21 12:10"
completed: "2026-09-21 12:10"
time_spent: ""
---

# Task Record: T-quick-doc-drift Detect Spec Drift

## Summary
Spec drift detection for ledger-hide-expired-orphans: SKIP with reason — no spec surface exists to drift against. docs/business-rules/ and docs/conventions/ do not exist in D:\Haven\docs, e-cam-web, or e-cam-service, and zero files anywhere under D:\Haven\docs contain a `domains:` frontmatter, so no spec can overlap the changed-file set. Feature's changed files were identified from commits directly (git diff main...HEAD empty because feature commits are already merged to main): e-cam-web 6393454 (src/api/cert.ts, src/views/cert/ledger/{CertTable.vue,format.ts,index.vue} + tests), fbf31aa (logs diagnose card tests), 57e918a (justfile); e-cam-service 1f4ecd5 (internal/cert/service/ledger_service.go, internal/cert/web/ledger_handler.go + tests). No spec files were fixed (none exist); no commits made. forge:consolidate-specs skill not invoked because its premise (existing specs with domains frontmatter) is unsatisfied — a scan would be a no-op.

## Changes

### Files Created
无

### Files Modified
无

### Key Decisions
无

## Document Metrics
drift-check: skip (no spec files); specs-overlapping-changes: 0; files-fixed: 0

## Referenced Documents
- docs/features/ledger-hide-expired-orphans/tasks/quick-drift-detection.md

## Review Status
final

## Acceptance Criteria
- [x] Identify files changed by this feature via git diff
- [x] List spec files in docs/business-rules/ and docs/conventions/ with domains frontmatter and verify overlapping ones
- [x] Auto-fix drifted specs and commit with [auto-specs] tag

## Notes
Dependency task 3 (联调验证) is not complete, but this check did not depend on post-integration state — it was purely structural (spec existence). If specs for the cert-ledger domain are added later, re-run drift detection against commits 6393454 (e-cam-web) and 1f4ecd5 (e-cam-service).
