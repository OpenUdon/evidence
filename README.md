# Evidence

`github.com/OpenUdon/evidence` is the small shared trust/evidence primitive
module for OpenUdon projects. It is intentionally product-neutral: packages here
define reusable record shapes and deterministic helpers, not product workflows,
runtime execution, policy engines, or storage.

License: Apache-2.0.

## Packages

- `artifact`: deterministic artifact records, safe relative path validation,
  manifests, manifest digests, and product label options for preserving
  domain-specific validation wording.
- `digest`: canonical digest records and SHA-256 helpers.
- `diagnostic`: product-neutral diagnostic records, severity normalization, and
  deterministic diagnostic sorting.
- `redact`: product-neutral redaction helpers for secret-like strings and
  JSON/YAML-like documents, including quoted and unquoted assignments while
  preserving authentication scheme names such as Bearer and Basic.
- `approval`: neutral approval evidence records, approver normalization,
  requirement evaluation, validation diagnostics, expiry-aware validation, and
  deterministic approval record digests.

## Shared Behaviors

- Redaction covers common secret assignment forms such as `password=value` and
  token-like provider credentials while allowing products to add stricter
  policy patterns.
- Approval validation ignores approvers without explicit timestamps, defaults
  the validation clock when callers omit one, and uses collision-free sorting
  keys for deterministic digests.
- Artifact path validation accepts product labels for root, path, and input
  names so downstream CLIs can keep stable error messages while sharing one
  implementation.

## Boundary

Ramen and OpenUdon may use these primitives behind their own public wire
formats, such as `ramen.approval.v1`, `ramen.policy.v1`,
`openudon.approval.v1`, handoff package digests, run evidence, and product
diagnostics. Those product-specific schemas, commands, state models, package
layouts, governance rules, executor boundaries, and trusted-runner semantics
remain documented and implemented in their owning modules.

Do not add executor interfaces, CLI plumbing, run orchestration, policy engines,
Ramen reconciliation behavior, OpenUdon authoring behavior, or product-specific
approval states to this module.

## Checks

```bash
go test ./...
go vet ./...
git diff --check
```
