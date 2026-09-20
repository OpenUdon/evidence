# Product

Evidence is a public Go module for reusable trust, artifact, and approval
evidence primitives. It gives OpenUdon, Ramen, and Authoring one shared place
for deterministic records that describe what was produced, reviewed, redacted,
approved, or handed across a trust boundary.

## Users

- OpenUdon package/review code that needs stable artifact manifests, package
  digests, diagnostics, and approval evidence.
- Ramen planning and governance code that needs reusable plan/report evidence
  without importing OpenUdon.
- Authoring code that needs generic diagnostics, artifact summaries, transcript
  redaction, and durable authoring evidence.
- Operators and reviewers who need clear, reproducible evidence records before
  approving generated workflows or desired-state changes.

## Product Scope

Evidence owns:

- Digest and manifest primitives for files, byte streams, directories, and
  generated artifact sets.
- Generic artifact records with stable paths, media types, sizes, digests, and
  classifications.
- Content-addressed descriptors and lifecycle assessments that downstream
  products can embed in static catalogs and package evidence.
- Diagnostic records and severities that can be rendered by downstream tools.
- Redaction helpers for secret-like values in prompts, transcripts, reports,
  plans, and generated artifacts.
- Approval evidence primitives, including who/what/when metadata supplied by
  downstream callers.

## Non-Goals

- No OpenUdon package layout, prompt, workflow-intent, review-template, or
  trusted-runner semantics.
- No Ramen desired-state profile, graph, plan, state, reconciliation, or
  provider semantics.
- No UWS workflow schema or document behavior.
- No API source parsing or operation ranking.
- No credential lookup, live API calls, model-provider calls, or execution.
- No durable database ownership.
- No catalog identity, registry protocol, storage backend, membership,
  cryptographic key trust, signature policy, or publication authorization.

## Design Principles

- Deterministic by default: the same inputs should produce the same evidence
  records.
- Product neutral: records can be embedded by OpenUdon, Ramen, or Authoring
  without importing those products.
- Caller-bound policy: product-specific decisions, identity, clocks, and
  approval workflows are injected by consumers.
- Safe reporting: diagnostics and records should support redaction before
  durable storage.
