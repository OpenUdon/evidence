# AGENTS.md

## Purpose

Evidence is the public shared trust and artifact evidence toolkit for
OpenUdon, Ramen, and Authoring. It owns generic digest, artifact safety,
diagnostic, redaction, and approval-record primitives that can be reused across
workflow authoring, review packaging, desired-state planning, and handoff
boundaries.

Module path:

```text
github.com/OpenUdon/evidence
```

Evidence must not own OpenUdon package semantics, Ramen desired-state profile
semantics, UWS workflow semantics, API source parsing, credential resolution,
model prompting, or live execution.

## Memory Bank First

The tracked canonical Evidence harness snapshot lives in `../tofu/evidence`.
In a normal `../evidence` checkout, `AGENTS.md`, `memory-bank/`, and
`evolution/` may be symlinks to this tracked snapshot so agents can use local
paths while planning history is committed in the `../tofu` repository.

Before substantial changes, read in this order:

1. [memory-bank/product.md](memory-bank/product.md)
2. [memory-bank/architecture.md](memory-bank/architecture.md)
3. [memory-bank/tech-stack.md](memory-bank/tech-stack.md)
4. [memory-bank/milestone.md](memory-bank/milestone.md)
5. The relevant per-milestone status file in [memory-bank/](memory-bank/)

Use the memory bank as the active project source of truth. Do not recreate
duplicate root-level product, architecture, roadmap, or aggregate status
documents.

This project exposes [GOAL.md](GOAL.md), one optional protocol for goal requests
that span multiple status files. Follow it only when a request names it.

A `GOAL.md` run is a deliberate exception to the row-level commit rule below.
For that run, `COMMIT_POLICY: none` — the protocol default — means no commits,
while `COMMIT_POLICY: task` keeps the usual one-commit-per-row cadence.
Precedence is the request, then `GOAL.md`, then this file; only commits are
delegated, and only during the run.

## Boundaries

- `../authoring` owns generic session, transcript, structured-output, and
  authoring-loop orchestration. Authoring may consume Evidence primitives for
  artifact reports, diagnostics, and transcript redaction.
- `../uws` owns public workflow semantics, document model, schema lookup, and
  validation. Evidence may hash, classify, and report on UWS files as generic
  artifacts, but must not define UWS behavior.
- `../apitools` owns API source discovery, metadata, operation summaries,
  auth/security summaries, and ranking. Evidence may carry generic source
  metadata in records supplied by consumers, but must not parse API sources.
- `../openudon` owns OpenUdon workflow authoring, package review, approval
  templates, package layout, and trusted-runner handoff.
- `../ramen` owns desired-state profiles, plans, diffs, reconciliation state,
  graphing, governance, and trusted executor boundaries.

Rule of thumb:

- If it is a deterministic digest, canonical artifact record, diagnostic
  envelope, redaction helper, or approval evidence primitive useful to multiple
  products, it can belong in Evidence.
- If it names OpenUdon packages, Ramen resources, UWS fields, API source
  families, providers, credentials, accounts, or execution behavior, it belongs
  downstream or in the owning sibling module.

## Execution Model

Evidence is mostly deterministic library code. Shared algorithms should accept
plain inputs and return stable records without hidden IO or product behavior.

Where runtime-dependent behavior is unavoidable, keep it behind narrow
interfaces that downstream products bind explicitly. Examples include clocks,
identity attestations, policy decisions, external key lookup, or artifact
readers supplied by a caller.

The upstream package should own common record shapes, digest algorithms,
canonical sorting, validation, and redaction mechanics. OpenUdon, Ramen, and
Authoring should own product-specific policy, artifact taxonomy, package
layout, approval workflow, and persistence.

## Commands

Initial harness/documentation checks:

```bash
git -C ../tofu diff --check -- evidence
git -C ../evidence status --short
```

Planned public module checks once Go code exists:

```bash
go test ./...
go vet ./...
git diff --check
```

When exported APIs change, run dependent checks in sibling consumers as
applicable:

```bash
(cd ../authoring && go test ./...)
(cd ../openudon && go test ./...)
(cd ../ramen && go test ./...)
```

## Safety

- Treat all artifact bytes, generated reports, plans, transcripts, approvals,
  and source documents as untrusted until validated by the consumer.
- Do not store secrets in examples, reports, diagnostics, approval records, or
  tests. Redaction helpers should default toward preserving structure while
  removing values.
- Do not execute workflows, API operations, Terraform/OpenTofu behavior,
  model-provider calls, or trusted-runner actions from Evidence.
- Default tests must be provider-free, model-free, network-free, and credential
  free.
- Keep durable record formats stable and deterministic once consumed by
  downstream repositories.

## Documentation Rules

- Update [memory-bank/milestone.md](memory-bank/milestone.md) when milestone
  scope, sequencing, acceptance criteria, current-state dashboard, boundaries,
  or the status-file index changes.
- When a milestone has multiple implementation tasks, update the matching
  `memory-bank/status-<LANE><NN>.md` file with task rows, state, notes, and scoped
  commit tracking.
- Keep one permanent, zero-padded status file for every milestone indexed by
  `milestone.md`. Never reuse an ID or create aggregate `status.md`.
- Keep candidate directions unnumbered until a fresh scope and dependency
  review promotes them.
- Write task ledgers as `Item | State | Notes` with a backticked marker in the
  second column: `` `[ ]` ``, `` `[+]` ``, `` `[~]` ``, `` `[!]` ``, or
  `` `[X]` ``.
- Treat each row as a commit unit. Parallel artifact-integrity and trust-record
  work requires explicit non-overlapping ownership, resolved prerequisites,
  and downstream impacts in `milestone.md`.
- Update [memory-bank/product.md](memory-bank/product.md) when product scope,
  users, workflows, concepts, or non-goals change.
- Update [memory-bank/architecture.md](memory-bank/architecture.md) when system
  boundaries, package layout, data flow, execution model, or security
  boundaries change.
- Update [memory-bank/tech-stack.md](memory-bank/tech-stack.md) when
  dependencies, commands, runtime assumptions, artifact schemas, or tooling
  choices change.
- Check [evolution/](evolution/) after a major review, milestone, or boundary
  change. Add the next prompt/result version only when product direction,
  architecture boundary, milestone target, or public/private contract direction
  materially changes.
