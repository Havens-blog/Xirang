---
status: "completed"
started: "2026-10-08 22:16"
completed: "2026-10-08 22:34"
time_spent: "~18m"
---

# Task Record: fix-1 fix unit-test: just unit-test failure in quality gate

## Summary
Fixed two deterministic, pre-existing test failures in e-cam-service/internal/cert/repository surfaced by the quality gate (unrelated to the remove-ecmdb-stack feature which only deleted ecmdb dirs / edited config). (1) TestCertificateRepo_GetByIDAndAttachPrivateKey: certificateRepository.Create discarded the InsertOne result and never populated cert.ID (bson:"_id,omitempty" defeats driver write-back), so GetByID right after Create hit mongo:ErrNoDocuments. Fixed by generating the ObjectID client-side in Create when zero, mirroring the existing crdRegistrationRepository.Create pattern. (2) TestCertificateRepo_ListPage: testFingerprint generated rotations of the hex cycle 0123456789abcdef, so testFingerprint(1)[:10]='123456789a' matched all four seeded fingerprints at the rotation boundary (expected 1, got 4). Fixed the generator to use a unique per-seed 2-hex-char marker prefix plus 'f' fill so no substring collides across seeds.

## Changes

### Files Created
无

### Files Modified
- e-cam-service/internal/cert/repository/certificate.go
- e-cam-service/internal/cert/repository/certificate_test.go

### Key Decisions
- Fix Create to set cert.ID = primitive.NewObjectID() when zero, matching crd_registration.go's established pattern; verified production callers (verify_window_service_test.go) already pre-set ID so the zero-guard is a no-op for them
- Rewrite testFingerprint with marker-prefix+fill rather than changing the search assertion, to preserve the partial-fingerprint-fragment search intent of TestCertificateRepo_ListPage; verified only cert_ledger_repo_test.go:78 slices the fingerprint

## Test Results
- **Tests Executed**: Yes
- **Passed**: 35
- **Failed**: 0
- **Coverage**: 62.4%

## Acceptance Criteria
- [x] TestCertificateRepo_ListPage passes
- [x] TestCertificateRepo_GetByIDAndAttachPrivateKey passes
- [x] Full ./internal/cert/repository/ package passes with no regressions
- [x] Full ./internal/cert/... subtree passes (Create change is additive, no regression in service/deployer callers)
- [x] go build ./... + go vet + gofmt clean on changed files

## Notes
Root cause was NOT environmental/MongoDB data isolation as initially suspected by the dispatcher — both failures are deterministic and reproduce identically across retries. Failure 1 is a test-data design bug (rotation collision); failure 2 is a production code bug (Create not populating _id). The remove-ecmdb-stack feature did not touch the cert package; these are pre-existing latent bugs only surfaced because the quality gate runs against a live MongoDB (tests t.Skip when DB is unreachable). Fix boundaries forbade the full `just unit-test` suite, so verification was scoped to the affected cert subtree per dispatcher instruction.
