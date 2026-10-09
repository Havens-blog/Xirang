---
status: "completed"
started: "2026-10-08 22:07"
completed: "2026-10-08 22:10"
time_spent: "~3m"
---

# Task Record: T-quick-doc-drift Detect Spec Drift

## Summary
Detected spec drift between project-level specs and the remove-ecmdb-stack feature. Scanned docs/business-rules/ and docs/conventions/ for ecmdb/cmdb/ports(8000/8097/3333) references. Found one drifted rule: BIZ-auth-005 in docs/business-rules/auth.md referenced a stale transitional clause '过渡期含 ecmdb-web' (ecmdb-web directory and nginx /api/cmdb route removed by feature tasks 2-3). Removed the clause; rule now covers only 控制台 / e-cam-web. No orphaned or implicit-new rules. No conventions drift. Git commit skipped (D:\Haven is not a git repo); traceability marker written to docs/features/remove-ecmdb-stack/specs/.drift-fixed.

## Changes

### Files Created
- docs/features/remove-ecmdb-stack/specs/.drift-fixed

### Files Modified
- docs/business-rules/auth.md

### Key Decisions
无

## Document Metrics
specs scanned: 3 business-rules + 11 conventions = 14 files; drifted: 1 (BIZ-auth-005); orphaned: 0; implicit-new: 0

## Referenced Documents
- docs/features/remove-ecmdb-stack/manifest.md
- docs/features/remove-ecmdb-stack/tasks/records/2-rewire-nginx-default-route.md
- docs/features/remove-ecmdb-stack/tasks/records/3-delete-ecmdb-directories.md
- docs/decisions/0003-remove-ecmdb.md

## Review Status
final

## Acceptance Criteria
- [x] All project-level specs in docs/business-rules/ and docs/conventions/ checked for drift against remove-ecmdb-stack feature scope
- [x] Drifted specs auto-fixed
- [x] No orphaned or stale ecmdb/cmdb references remain in project-level specs

## Notes
Drift-only mode (no prd-spec.md/tech-design.md in feature — quick mode). Discovery via task records (tasks 1-4 completed) instead of git diff since D:\Haven is not a git repo. Non-interactive: [auto-specs] tag recorded in .drift-fixed marker (commit skipped, non-git). Vocabulary index generation skipped — no docs/vocabulary.md existed prior and drift-only quick mode has no consolidation output to index.
