---
status: "completed"
started: "2026-09-09 13:30"
completed: "2026-09-09 13:36"
time_spent: "~6m"
---

# Task Record: T-quick-doc-drift Detect Spec Drift

## Summary
Doc-drift audit (quick mode): verified feature docs vs executed state (commits be3bd99/c4cfd70/dc3eee0/0a7bcbc/a15ebc0). Verified current: eval/followup.md (all 10 dangling internal/task comment refs re-checked at exact file:line in e-cam-service), proposal.md SC wording (archive commits be3bd99/c4cfd70 exist, tmp-k8s-cred excerpt exists), baseline-record.md vs acceptance-record.md consistency (6-baseline vs 5-post subset, a15ebc0 reduction). FIXED 2 drift items: (1) manifest.md - all 6 tasks showed 'pending' and frontmatter status 'tasks' despite full execution; updated to completed, added review-doc + fix-1 rows, Documents links, execution summary with commit hashes. (2) docs/decisions/0001 Phase 4 addendum - future-tense claims ('由任务 4 执行' / '任务 4 落地后将反向链接') stale after task 4 executed; refreshed to past tense citing commit 0a7bcbc, asset/doc.go backlink confirmed present. No e-cam-service code touched.

## Changes

### Files Created
无

### Files Modified
- docs/features/dead-modules-cleanup/manifest.md
- docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md

### Key Decisions
- manifest.md statuses aligned to index.json source of truth (all completed) instead of being left as plan-time placeholders
- decision-0001 addendum tense refreshed because its future-tense assertions describe actions that have since actually executed (commit 0a7bcbc), so the text would mislead readers into thinking task 4 is still pending
- followup.md left untouched: all 10 file:line refs verified accurate against live code

## Document Metrics
N/A

## Referenced Documents
无

## Review Status
N/A

## Acceptance Criteria
无

## Notes
Drift-only mode: no docs/business-rules or docs/conventions exist; no [auto-specs] project-level spec changes needed. Referenced docs: manifest.md, baseline-record.md, acceptance-record.md, proposal.md, eval/followup.md, decisions/0001, records/review-doc.md. reviewStatus=passed; docMetrics: filesScanned=7, driftFound=2, filesFixed=2, followupRefsVerified=10.
