# Status A01 - Content Descriptors And Lifecycle Assessments

## Goal

Add deterministic neutral records that Browsertools can publish and OpenUdon
can verify without moving browser, registry, storage, or policy semantics into
Evidence.

## State

Completed.

## Task Ledger

| Item | State | Notes |
|---|---|---|
| Define content descriptors | `[+]` | Added media type, non-negative size, SHA-256 identity, normalized collision-checked annotations, validation, canonical JSON, and digest helpers. |
| Define lifecycle assessments | `[+]` | Added active/stale/revoked/superseded records with explicit assessed/expiry times, successor rules, supporting descriptors, explicit-clock status evaluation, canonical JSON, and digests. |
| Harden malformed and deterministic cases | `[+]` | Tests cover algorithms/hex, time inversion, successor consistency, duplicates, self-supersession, input immutability, map/order stability, optional wire fields, JSON round trips, and exact whole-key redaction extensions that avoid browser-domain substring false positives. |
| Reconcile downstream adoption | `[+]` | Browsertools P02 and OpenUdon A01 name the public types, retain product policy, and require compatibility gates; actual adoption closes in those dependent milestones. |
| Update public docs and evolution | `[+]` | README, product, architecture, tech stack, and evolution v3 document the storage/network/signature/trust boundary. |
| Run module and consumer gates | `[+]` | Added standalone CI; Evidence test/vet/diff and current Browsertools/OpenUdon full workspace tests pass. |

## Acceptance

- [x] Records are deterministic, versioned, JSON-friendly, and free of product names.
- [x] Validation fails closed without network, clocks, key lookup, or persistence.
- [x] Browsertools P02 and OpenUdon A01 carry explicit adoption and compatibility tasks; the overall goal requires both consumers to use the records.
- [x] Evidence still owns no registry, browser, UWS, credential, or execution behavior.
