# Tech Stack

## Module

```text
github.com/OpenUdon/evidence
```

Go version:

```text
1.26.3
```

Evidence should start with the Go standard library wherever practical. Add
third-party dependencies only when they provide clear value for stable parsing,
canonicalization, or validation that should not be reimplemented locally.

## Planned Dependencies

- Standard library hashing, IO, path, time, and encoding packages for the first
  digest/artifact slices.
- No dependency on Authoring, OpenUdon, Ramen, udon, provider SDKs, or model
  SDKs.
- Optional future dependency on UWS only if a generic artifact helper truly
  needs public UWS types; prefer keeping Evidence UWS-neutral.

## Commands

Repository checks:

```bash
git -C ../tofu diff --check -- evidence
git -C ../evidence status --short
```

Go checks:

```bash
go test ./...
go vet ./...
git diff --check
../skills/harness/tackle-memory-bank-api-loop --model lane-audit .
```

Dependent checks after exported API changes:

```bash
(cd ../authoring && go test ./...)
(cd ../openudon && go test ./...)
(cd ../ramen && go test ./...)
```

## Artifact And Record Expectations

- Records should be deterministic and JSON/YAML friendly.
- Sorting must be stable and documented where it affects durable evidence.
- Paths should use slash-separated relative paths unless a caller explicitly
  asks for host paths.
- Digest algorithms should be explicit in record fields.
- Public content descriptors currently accept only SHA-256 and normalize media
  types with the Go standard library. Lifecycle assessments use explicit UTC
  times and deterministic supporting-descriptor order.
- Redaction should preserve enough shape for review while removing secret
  values.
- Async evidence records under `async/` are neutral record envelopes only:
  operation identity, attempts, execution request/response summaries, status
  observations, confirmation read observations, runtime hints, and digest
  references. They do not define convergence, package policy, desired hashes,
  trusted-runner behavior, or provider polling logic.

## Runtime Assumptions

- No live credentials.
- No live network calls.
- No model-provider calls.
- No execution of generated artifacts.
- Tests run in public CI with only mock or in-memory inputs.
