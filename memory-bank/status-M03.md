# Status M03

## Milestone

Diagnostics and redaction primitives.

## State

`complete`

## Tasks

| Item | State | Notes |
|---|---|---|
| Add generic diagnostic records. | `[+]` | `diagnostic.Record` carries severity, code, message, location, remediation, and structured detail. |
| Add diagnostic helpers. | `[+]` | Severity normalization, error detection, and deterministic sorting are product-neutral. |
| Add string redaction helpers. | `[+]` | Redacts known credential-shaped values, bearer tokens, JWT-like values, private keys, and sensitive assignments. |
| Add document redaction helpers. | `[+]` | Supports nested `map[string]any`, `map[string]string`, `map[any]any`, `[]any`, `[]string`, and string values. |
| Add tests for milestone acceptance. | `[+]` | Covers diagnostic ordering, severity handling, secret-like strings, sensitive keys, nested documents, and custom options. |

## Verification

Expected checks:

```bash
(cd ../evidence && go test ./...)
(cd ../evidence && go vet ./...)
(cd ../evidence && git diff --check)
git -C ../tofu diff --check -- evidence
```

## Notes

This slice intentionally provides reusable mechanics only. Downstream products
still decide which diagnostics are blocking, which redaction paths are required
for their schemas, and how redacted evidence is persisted or reviewed.
