# Status M01

## Milestone

Harness and boundary setup for the public Evidence module.

## State

`complete`

## Tasks

| Item | State | Notes |
|---|---|---|
| Create tracked harness snapshot under `../tofu/evidence`. | `[+]` | Includes `AGENTS.md`, memory bank docs, and evolution seed files. |
| Define product scope and non-goals. | `[+]` | Evidence owns deterministic evidence primitives, not product policy. |
| Define architecture boundaries. | `[+]` | Authoring, OpenUdon, and Ramen consume Evidence; Evidence does not import them. |
| Define initial milestone sequence. | `[+]` | M02-M06 cover digest/artifact, diagnostics/redaction, approval, and migrations. |
| Add root module skeleton in `../evidence`. | `[+]` | Uses module path `github.com/OpenUdon/evidence`. |

## Verification

Expected checks:

```bash
git -C ../tofu diff --check -- evidence
(cd ../evidence && go test ./...)
(cd ../evidence && go vet ./...)
(cd ../evidence && git diff --check)
```

## Notes

Evidence starts as a small upstream library. The first implementation milestone
should avoid product-specific names and focus on deterministic records and
tests that can be consumed by Authoring, OpenUdon, and Ramen.
