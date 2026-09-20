# Status M07

## Milestone

Shared async operation evidence contracts.

## State

`complete`

## Tasks

| Item | State | Notes |
|---|---|---|
| Define execution request evidence. | `[+]` | Added `async.ExecutionRequest` with neutral operation identity, attempt metadata, non-secret transport metadata, runtime hints, digest references, validation, normalization, and stable digest helper. No Ramen desired-state, OpenUdon package policy, or assumption that OpenUdon produces Ramen requests. |
| Define execution response evidence. | `[+]` | Added `async.ExecutionResponse` with outcome/status summaries, timing, error summary, payload digest references, correlation IDs, validation, normalization, and stable digest helper. |
| Define status observation evidence. | `[+]` | Added `async.StatusObservation` with status marker, terminality hint, timestamps, correlation IDs, and payload digest references without encoding convergence policy. |
| Define confirmation read observation evidence. | `[+]` | Added `async.ConfirmationReadObservation` for missing/exists/read-result evidence that downstream products can embed while leaving Ramen to decide whether it satisfies convergence. |
| Define attempt and ordering metadata. | `[+]` | Added `async.AttemptMetadata` with `evidence_id`, `attempt_id`, sequence, actor/source annotations, and recorded timestamp normalization. |
| Document redaction and hash boundaries. | `[+]` | Architecture and tech-stack now state that runtime hints are execution metadata and Evidence must not define desired-state hash semantics, package policy, convergence, or provider polling logic. |

## Review Findings

| Severity | Finding | Resolution |
|---|---|---|
| High | Initial validation normalized empty/wrong record versions to the expected version, so malformed records could pass validators. | Fixed normalization to trim but not default versions; validators now require the caller-supplied version to match the expected record version. |
| Low | The M52 review showed Evidence records must not assume OpenUdon produces Ramen-native execution requests. | Execution request docs and M29 notes now say product flows produce/embed records independently. |

## Verification

Completed checks:

```bash
go test ./...
go vet ./...
git diff --check
git -C ../tofu diff --check -- evidence openudon ramen
```

## Notes

This milestone follows OpenUdon M52's classification: Evidence owns only
neutral record shapes. Ramen owns convergence and state; OpenUdon owns package
and trusted-runner lifecycle.
