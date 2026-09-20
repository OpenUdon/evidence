# Milestones

This file owns Evidence milestone sequencing, acceptance criteria, current
state, and the status-file index.

## Status ID Pattern

Evidence status files use one uppercase domain letter and a zero-padded number
from `01` through `99`:

```text
M01, M02, M03, ... M09, M10, M11, ...
A01, A02, A03, ...
T01, T02, T03, ...
```

Task/status files use the lane ID:

```text
tabilet/memory-bank/status-M01.md
tabilet/memory-bank/status-M02.md
```

Do not reuse IDs after a status file exists.

Lane meanings:

- `M`: completed legacy work and future cross-cutting public contracts.
- `A`: artifact integrity primitives: digest, artifact, diagnostic, and
  redaction records and algorithms.
- `T`: trust records: approval, attestation, and async execution evidence.

Do not reclassify completed M history. Keep cancelled IDs with `[X]` rows and
never create aggregate `status.md`. Lane letters classify ownership, not
execution order. Independent A and T work may proceed together only when
milestone sections record non-overlapping packages, resolved prerequisites,
and downstream impacts; shared contracts remain in M. Prefer one active
implementation milestone per lane.

## Current State

Evidence has completed M01-M07, including approval evidence, downstream
adoption, and shared async operation evidence contracts. M08 migrated the
private harness to permanent domain lanes. The
repository exists as the shared upstream package for deterministic trust,
artifact, diagnostic, redaction, and approval evidence primitives used by
Authoring, OpenUdon, and Ramen.

A01 is the active artifact-integrity milestone. It adds product-neutral
content descriptors and lifecycle assessments needed by both Browsertools and
OpenUdon for a static browser-capability catalog. It deliberately excludes
registry identity, network transport, storage, cryptographic signatures, and
publication policy.

The key architectural decision is to keep Evidence product-neutral and mostly
deterministic. Runtime-dependent behavior such as identity, clocks, policy, and
storage is bound by downstream consumers through narrow interfaces.

## Delivery Strategy

Build Evidence in small provider-free slices:

1. `M01`: establish the public module harness, boundaries, status tracking,
   and initial docs.
2. `M02`: implement digest and artifact manifest primitives.
3. `M03`: implement generic diagnostic records and redaction helpers.
4. `M04`: implement approval evidence records and caller-bound attestation
   interfaces.
5. `M05`: migrate shared OpenUdon package/report evidence code to Evidence.
6. `M06`: migrate Ramen plan/governance evidence code to Evidence where it is
   product-neutral.
7. `M07`: define shared async operation evidence records only after Ramen and
   OpenUdon agree on neutral fields that do not encode product policy.
8. `M08`: migrate the private harness to permanent parallel lanes.
9. `A01`: add canonical content descriptors and lifecycle assessments for
   digest-addressed artifact distribution.

## Status Files

| Milestone | Status File | Summary |
|---|---|---|
| M01 | [status-M01.md](status-M01.md) | Harness and boundary setup. |
| M02 | [status-M02.md](status-M02.md) | Digest and artifact manifest primitives. |
| M03 | [status-M03.md](status-M03.md) | Diagnostics and redaction primitives. |
| M04 | [status-M04.md](status-M04.md) | Approval evidence records. |
| M05 | [status-M05.md](status-M05.md) | OpenUdon adoption of neutral evidence primitives. |
| M06 | [status-M06.md](status-M06.md) | Ramen adoption of neutral evidence primitives. |
| M07 | [status-M07.md](status-M07.md) | Shared async operation evidence contracts. |
| M08 | [status-M08.md](status-M08.md) | Parallel-lane harness migration. |
| A01 | [status-A01.md](status-A01.md) | Content descriptors and lifecycle assessments. |

## Candidate Directions

Candidates have no lane, ID, status file, or execution-order entry until a
fresh scope and dependency review promotes them.

| Direction | Why Deferred | Promotion Trigger |
|---|---|---|
| Additional product-neutral record families | A01 covers the agreed Browsertools/OpenUdon distribution gap; no other shared record gap has been demonstrated. | At least two downstream products agree on another neutral field set and boundary tests. |
| Cryptographic attestations or signatures | Key lookup, identity trust, revocation, and authorization policy cannot be safely hidden in a deterministic record package. | A bounded public verification contract separates caller-supplied trust policy from neutral record encoding. |
| Further downstream helper migration | Current consumers already use the approved neutral packages; remaining helpers may encode product policy. | A duplication review identifies behavior that is demonstrably product-neutral and compatibility-tested. |

## Milestones

### M01 Harness And Boundary Setup

**Goal.** Establish Evidence as a public Go module with Ramen-style memory bank
harness docs, explicit ownership boundaries, and baseline commands.

Acceptance:

- `AGENTS.md`, `tabilet/memory-bank/product.md`, `tabilet/memory-bank/architecture.md`,
  `tabilet/memory-bank/tech-stack.md`, `tabilet/memory-bank/milestone.md`, and
  `tabilet/memory-bank/status-M01.md` exist through the tracked snapshot.
- The root checkout has the same symlink-facing harness pattern used by Ramen.
- `go.mod` exists with module path `github.com/OpenUdon/evidence`.
- Documentation states that Evidence owns deterministic shared evidence
  primitives while downstream products own policy and runtime binding.

### M02 Digest And Artifact Manifests

**Goal.** Define stable digest helpers and artifact manifest records for files,
byte streams, directories, and generated artifact sets.

Acceptance:

- Digest records include algorithm and encoded value.
- Manifest records sort paths deterministically.
- Tests cover byte, file, directory, missing path, and ordering behavior.

### M03 Diagnostics And Redaction

**Goal.** Define generic diagnostic envelopes and reusable redaction helpers.

Acceptance:

- Diagnostic records support severity, code, message, location, and optional
  structured detail.
- Redaction helpers work for common string, map, and JSON/YAML-like document
  shapes.
- Tests cover stable ordering and secret-like value handling.

### M04 Approval Evidence Records

**Goal.** Define product-neutral approval evidence records and optional
interfaces for caller-bound identity and time.

Acceptance:

- Records can represent reviewer/operator identity supplied by a caller,
  approval subject, timestamps, reasons, and artifact digest references.
- Evidence does not decide product policy satisfaction.
- Tests use fake clocks and identities only.

### M05 OpenUdon Migration

**Goal.** Move product-neutral OpenUdon digest, artifact, diagnostic, and
approval evidence helpers into Evidence without changing package behavior.

Acceptance:

- OpenUdon package/review/handoff tests pass.
- OpenUdon-specific templates, prompts, package layout, and policy stay in
  OpenUdon.

### M06 Ramen Migration

**Goal.** Move product-neutral Ramen plan/report/governance evidence helpers
into Evidence.

Acceptance:

- Ramen tests pass.
- Ramen desired-state profile, graph, plan semantics, state, and reconciliation
  behavior stay in Ramen.

### M07 Shared Async Operation Evidence Contracts

**Goal.** Define product-neutral async operation evidence records that Ramen
and OpenUdon can both embed without moving package policy, trusted-runner
behavior, or desired-state convergence into Evidence.

Acceptance:

- Records cover execution request/response, status observation, confirmation
  read observation, and attempt/evidence linkage metadata.
- Runtime hints are represented as execution metadata, not desired-state hash
  or policy inputs.
- Evidence remains deterministic, redaction-aware, append-only, and free of
  Ramen/OpenUdon imports.

### M08 Parallel-Lane Harness Migration

**Goal.** Adopt permanent domain lanes and runner-compatible task ledgers
without changing Evidence APIs.

Acceptance:

- M01-M07 history remains permanent and every allocated milestone has a
  parseable status file.
- Future artifact-integrity and trust-record work has explicit ownership.
- Candidate work remains unnumbered until approval.
- Structural, module, and unattended no-action runner checks pass.

### A01 Content Descriptors And Lifecycle Assessments

**Goal.** Give Browsertools and OpenUdon one deterministic, product-neutral
record vocabulary for content-addressed artifacts and their current
assessment without encoding registry or product policy.

Acceptance:

- `artifact.Descriptor` records an explicit media type, byte size, SHA-256
  digest, and deterministically normalized annotations.
- `artifact.Assessment` binds a descriptor to `active`, `stale`, `revoked`, or
  `superseded` status, assessment/expiry times, an optional successor digest,
  and supporting artifact descriptors.
- Normalize, validate, canonical JSON, and digest helpers are deterministic;
  invalid algorithms, sizes, statuses, time ordering, duplicate supporting
  descriptors, and malformed successor relationships fail closed.
- Browsertools P02 and OpenUdon A01 explicitly depend on and carry adoption
  tasks for the records; their consumer gates must pass before the overall
  cross-repository goal closes. Evidence imports no UWS, browser, registry,
  storage, network, identity, or runtime packages.
- Cryptographic signatures remain deferred; A01 describes supplied assessment
  evidence and never decides whether a caller trusts its issuer.
