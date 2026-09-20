# Status M02

## Milestone

Digest and artifact manifest primitives.

## State

`complete`

## Tasks

| Item | State | Notes |
|---|---|---|
| Add typed digest records. | `[+]` | `digest.Record` carries algorithm and encoded value; SHA-256 helpers cover bytes, readers, and files. |
| Add safe artifact path helpers. | `[+]` | Slash-separated relative paths reject empty, absolute, parent traversal, backslash, control-character, and volume-prefixed inputs. |
| Add artifact file records. | `[+]` | Records include path, caller metadata, byte size, and digest. |
| Add deterministic manifests. | `[+]` | Manifests sort and deduplicate paths; directory collection walks regular files. |
| Add tests for milestone acceptance. | `[+]` | Covers bytes, files, directory ordering, missing file rejection, unsafe paths, and symlink rejection. |

## Verification

Expected checks:

```bash
(cd ../evidence && go test ./...)
(cd ../evidence && go vet ./...)
(cd ../evidence && git diff --check)
git -C ../tofu diff --check -- evidence
```

## Notes

This slice intentionally avoids OpenUdon or Ramen package policy. Downstream
packages can layer their required-file inventory, review state, and approval
rules on top of these deterministic primitives.
