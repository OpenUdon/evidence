package async

import (
	"testing"
	"time"

	"github.com/OpenUdon/evidence/digest"
)

func TestExecutionRequestNormalizesAndValidates(t *testing.T) {
	record := ExecutionRequest{
		Version: " evidence.async.execution-request.v1 ",
		Attempt: AttemptMetadata{
			EvidenceID: " ev-1 ",
			AttemptID:  " attempt-1 ",
			RecordedAt: time.Date(2026, 6, 3, 12, 0, 0, 0, time.FixedZone("T", 3600)),
		},
		Operation: OperationRef{
			Action:      " PUT ",
			Method:      " put ",
			SourceKind:  " openapi ",
			OperationID: " Databases_CreateOrUpdate ",
		},
		Runtime: RuntimeHints{
			Retry:  map[string]any{"max_attempts": 3},
			Waiter: map[string]any{"until": "exists"},
		},
		Transport: map[string]string{" request_id ": " abc "},
		Digests: []digest.Record{
			{Algorithm: " sha256 ", Value: " b "},
			{Algorithm: "sha256", Value: "a"},
		},
	}
	normalized := NormalizeExecutionRequest(record)
	if normalized.Version != ExecutionRequestVersion {
		t.Fatalf("version = %q", normalized.Version)
	}
	if normalized.Attempt.EvidenceID != "ev-1" || normalized.Attempt.AttemptID != "attempt-1" {
		t.Fatalf("attempt not normalized: %#v", normalized.Attempt)
	}
	if normalized.Attempt.RecordedAt.Location() != time.UTC {
		t.Fatalf("recorded_at not UTC: %s", normalized.Attempt.RecordedAt.Location())
	}
	if normalized.Operation.Action != "put" || normalized.Operation.Method != "PUT" || normalized.Operation.OperationID != "Databases_CreateOrUpdate" {
		t.Fatalf("operation not normalized: %#v", normalized.Operation)
	}
	if normalized.Transport["request_id"] != "abc" {
		t.Fatalf("transport not normalized: %#v", normalized.Transport)
	}
	if len(normalized.Digests) != 2 || normalized.Digests[0].Value != "a" || normalized.Digests[1].Value != "b" {
		t.Fatalf("digests not sorted/trimmed: %#v", normalized.Digests)
	}
	if diagnostics := ValidateExecutionRequest(record); len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestExecutionRequestDigestIsDeterministic(t *testing.T) {
	base := ExecutionRequest{
		Version:   ExecutionRequestVersion,
		Attempt:   AttemptMetadata{EvidenceID: "ev-1", AttemptID: "attempt-1"},
		Operation: OperationRef{OperationID: "op"},
		Transport: map[string]string{"b": "2", "a": "1"},
	}
	left, err := DigestExecutionRequest(base)
	if err != nil {
		t.Fatalf("left digest: %v", err)
	}
	right, err := DigestExecutionRequest(ExecutionRequest{
		Version:   ExecutionRequestVersion,
		Attempt:   AttemptMetadata{EvidenceID: " ev-1 ", AttemptID: " attempt-1 "},
		Operation: OperationRef{OperationID: " op "},
		Transport: map[string]string{"a": "1", "b": "2"},
	})
	if err != nil {
		t.Fatalf("right digest: %v", err)
	}
	if left != right {
		t.Fatalf("digest mismatch: %#v != %#v", left, right)
	}
}

func TestValidationRequiresAttemptAndOperation(t *testing.T) {
	diagnostics := ValidateExecutionResponse(ExecutionResponse{})
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"async.evidence_id_missing", "async.attempt_id_missing", "async.operation_id_missing", "async.outcome_missing"} {
		if !codes[code] {
			t.Fatalf("missing diagnostic %s in %#v", code, diagnostics)
		}
	}
}

func TestObservationRecordsDoNotDefineConvergencePolicy(t *testing.T) {
	status := StatusObservation{
		Version: StatusObservationVersion,
		Attempt: AttemptMetadata{EvidenceID: "ev-status", AttemptID: "attempt-1"},
		Operation: OperationRef{
			OperationID: "Azure-AsyncOperation",
		},
		Status:          "Succeeded",
		TerminalityHint: " success ",
	}
	if diagnostics := ValidateStatusObservation(status); len(diagnostics) != 0 {
		t.Fatalf("status diagnostics = %#v", diagnostics)
	}
	read := ConfirmationReadObservation{
		Version:   ConfirmationReadObservationVersion,
		Attempt:   AttemptMetadata{EvidenceID: "ev-read", AttemptID: "attempt-1"},
		Operation: OperationRef{OperationID: "Databases_Get"},
		Outcome:   " missing ",
	}
	normalized := NormalizeConfirmationReadObservation(read)
	if normalized.Outcome != "missing" {
		t.Fatalf("outcome = %q", normalized.Outcome)
	}
	if diagnostics := ValidateConfirmationReadObservation(read); len(diagnostics) != 0 {
		t.Fatalf("read diagnostics = %#v", diagnostics)
	}
}
