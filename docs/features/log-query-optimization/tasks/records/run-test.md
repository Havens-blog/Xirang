---
status: "completed"
started: "2026-09-16 15:46"
completed: "2026-09-16 15:46"
time_spent: ""
---

# Task Record: T-test-run Run API Functional Test

## Summary
Ran the log-query-optimization live API functional suite (5 journey packages under tests/, shared harness tests/logquerye2e/harness.go) against the deployed backend at localhost:8001, plus related unit tests. Environment verified first: dev JWT via mktok.py, live probe GET /api/v1/cam/logs/types -> 200 code=0; go vet clean on all 6 LQO test packages. Results: 51/51 live API tests PASS across all 5 journeys (log-query-lifecycle 10, field-filter-quick-values 11, detail-grouping-pagination 10, aggregate-drilldown-clear 10, waf-source-browsing 10; 0 skips). First run of aggregate-drilldown-clear had 2 transient failures (TestStep2_3_DrilldownFiltersToGroupValue: drilldown total 1569 > baseline 1296; TestStep3b_DrilldownPlusManualFilterCoexist: 740 -> 1624 out of band) — both are cross-call count-invariant checks against the fast-moving live WAF dataset (aggregate total grew 39.6M -> 40.7M during the run; rolling window() recomputes now per call); entry-level filter-semantics assertions passed in both runs; both tests PASSED on immediate re-run with zero changes to test or production code. Unit tests: go test ./internal/logquery/... and ./internal/shared/cloudx/logquery/... all ok (tencent pkg has no test files). No justfile in repo: dev/probe orchestration satisfied by the already-running external deployment, per-journey go test run sequentially (known OOM mitigation), teardown n/a (no local artifacts). Full report: tests/results/latest.md + per-journey raw outputs in tests/results/.

## Changes

### Files Created
无

### Files Modified
无

### Key Decisions
- Adapted skill orchestration to repo reality (no justfile): probe via curl+harness.RequireLive instead of just probe recipe; sequential per-package go test instead of per-journey just test.
- Re-ran only the 2 failing tests once to classify transient live-data drift vs code defect; passed with no modifications, so no fix task created.
- Report + raw outputs written to tests/results/ for evidence.

## Cases Generated
N/A

## Cases Evaluated
N/A

## Scripts Created
无

## Test Results
N/A

## Acceptance Criteria
- [x] All test cases MUST pass
- [x] Tests verify actual functional behavior against the live API surface (no placeholders/mocks)

## Notes
8001 is the deployed build predating the LQO latency backend changes; tests carry skip guards + tolerance bands by design and were run as-is. No test assertion or production code was modified. Candidate future hardening (not blocking): Step2_3 uses a strict got<=baseline cross-call check while sibling tests use a 2x band; observed live drift exceeds the strict form occasionally.
