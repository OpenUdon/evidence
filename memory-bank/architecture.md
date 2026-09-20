# Architecture

Evidence is the shared deterministic layer for artifact and trust records. It
owns common record shapes and stable algorithms while downstream products own
policy, package layout, workflow semantics, and persistence.

## System Boundary

- `github.com/OpenUdon/evidence` owns generic digest, artifact, diagnostic,
  redaction, and approval evidence primitives.
- `github.com/OpenUdon/authoring` owns authoring sessions, transcripts,
  structured-output helpers, and progressive authoring orchestration.
- `github.com/OpenUdon/uws` owns public UWS document semantics, model, schema
  lookup, validation, and artifact discovery.
- `github.com/OpenUdon/apitools` owns API source metadata and prompt-safe
  operation/auth/security summaries.
- `github.com/OpenUdon/openudon` owns OpenUdon authoring, review package
  generation, approval templates, package digests as product policy, and
  trusted-runner handoff.
- `github.com/OpenUdon/ramen` owns desired-state project semantics, graphing,
  planning, reconciliation state, governance, and trusted executor boundaries.

Evidence must not import Authoring, OpenUdon, or Ramen. Authoring, OpenUdon,
and Ramen may import Evidence.

## Execution Model

Evidence uses a deterministic function-first model:

```text
bytes / paths / caller metadata
  -> evidence algorithm
  -> stable record / diagnostics / redacted output
```

Runtime-dependent behavior is optional and bound by interfaces:

```go
type Clock interface {
    Now() time.Time
}

type IdentityProvider interface {
    Subject(context.Context) (Subject, error)
}
```

Concrete interface names will be settled during implementation. The ownership
rule should remain stable: Evidence owns common algorithms and records;
products bind identity, policy, package layout, approval workflow, and storage.

## Harness Layout

Private planning uses permanent `status-<LANE><NN>.md` ledgers. `M` retains
legacy/cross-cutting contracts, `A` owns artifact-integrity primitives, and `T`
owns approval/attestation/async trust records. The unattended runner reads the
second column of `Item | State | Notes` tables. Candidates remain unnumbered.

## Package Layout

```text
artifact/    file and generated-artifact records, classification, manifests
digest/      stable digest helpers for bytes, streams, files, and manifests
diagnostic/  generic diagnostic records, severities, sorting, rendering inputs
redact/      redaction rules and safe string/map/document helpers
approval/    generic approval evidence records and caller-bound attestations
```

`artifact/`, `digest/`, `diagnostic/`, and `redact/` are implemented. Later
package names may be adjusted during implementation. Keep product-specific
adapters out of this module.
`artifact/` also defines versioned content descriptors and lifecycle
assessments. These bind exact SHA-256 bytes to caller-supplied active, stale,
revoked, or superseded evidence. Registry identity, persistence, network
transport, issuer trust, signatures, and authorization remain downstream.
`async/` now defines product-neutral async execution evidence records for
request, response, status observation, confirmation read observation, attempt
metadata, validation, normalization, and stable digests. Ramen and OpenUdon
still own convergence, package policy, trusted-runner behavior, and storage.

## Data Flow

OpenUdon flow:

```text
generated workflow package
  -> artifact records and digests
  -> review diagnostics and approval evidence
  -> OpenUdon package/handoff artifacts
```

Ramen flow:

```text
project / plan / generated UWS
  -> artifact records and diagnostics
  -> governance approval evidence
  -> trusted executor handoff metadata
```

Authoring flow:

```text
session transcript / generated draft artifacts
  -> redacted transcript and artifact summaries
  -> downstream review or package evidence
```

## Security Boundary

- Evidence stores no credential values and performs no credential lookup.
- Inputs are untrusted bytes or records supplied by consumers.
- Redaction helpers are safety tools, not authorization boundaries.
- Approval evidence records describe supplied approvals; they do not decide
  whether a product policy is satisfied.
- Default tests must not require network, live providers, model credentials, or
  private repositories.
- Lifecycle evaluation accepts an explicit caller time; it never reads a
  hidden clock or treats an assessment as authorization.
