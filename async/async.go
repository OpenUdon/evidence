// Package async provides product-neutral evidence records for asynchronous
// operation submission, execution, status polling, and confirmation reads.
package async

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/OpenUdon/evidence/diagnostic"
	"github.com/OpenUdon/evidence/digest"
)

const (
	ExecutionRequestVersion            = "evidence.async.execution-request.v1"
	ExecutionResponseVersion           = "evidence.async.execution-response.v1"
	StatusObservationVersion           = "evidence.async.status-observation.v1"
	ConfirmationReadObservationVersion = "evidence.async.confirmation-read-observation.v1"
)

type OperationRef struct {
	SubjectKind string `json:"subject_kind,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	Action      string `json:"action,omitempty"`
	Method      string `json:"method,omitempty"`
	SourceKind  string `json:"source_kind,omitempty"`
	SourceID    string `json:"source_id,omitempty"`
	SourcePath  string `json:"source_path,omitempty"`
	OperationID string `json:"operation_id"`
}

type RuntimeHints struct {
	Retry      map[string]any `json:"retry,omitempty"`
	Waiter     map[string]any `json:"waiter,omitempty"`
	Pagination map[string]any `json:"pagination,omitempty"`
}

type AttemptMetadata struct {
	EvidenceID string    `json:"evidence_id"`
	AttemptID  string    `json:"attempt_id"`
	Sequence   int64     `json:"sequence,omitempty"`
	Actor      string    `json:"actor,omitempty"`
	Source     string    `json:"source,omitempty"`
	RecordedAt time.Time `json:"recorded_at,omitempty"`
}

type ExecutionRequest struct {
	Version   string            `json:"version"`
	Attempt   AttemptMetadata   `json:"attempt"`
	RequestID string            `json:"request_id,omitempty"`
	Operation OperationRef      `json:"operation"`
	Runtime   RuntimeHints      `json:"runtime,omitempty"`
	Transport map[string]string `json:"transport,omitempty"`
	Digests   []digest.Record   `json:"digests,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ExecutionResponse struct {
	Version           string            `json:"version"`
	Attempt           AttemptMetadata   `json:"attempt"`
	RequestEvidenceID string            `json:"request_evidence_id,omitempty"`
	ResponseID        string            `json:"response_id,omitempty"`
	Operation         OperationRef      `json:"operation"`
	Outcome           string            `json:"outcome"`
	StatusCode        int               `json:"status_code,omitempty"`
	StatusClass       string            `json:"status_class,omitempty"`
	CorrelationIDs    map[string]string `json:"correlation_ids,omitempty"`
	PayloadDigests    []digest.Record   `json:"payload_digests,omitempty"`
	ErrorSummary      string            `json:"error_summary,omitempty"`
	StartedAt         time.Time         `json:"started_at,omitempty"`
	FinishedAt        time.Time         `json:"finished_at,omitempty"`
}

type StatusObservation struct {
	Version           string            `json:"version"`
	Attempt           AttemptMetadata   `json:"attempt"`
	RequestEvidenceID string            `json:"request_evidence_id,omitempty"`
	Operation         OperationRef      `json:"operation"`
	Status            string            `json:"status,omitempty"`
	TerminalityHint   string            `json:"terminality_hint,omitempty"`
	CorrelationIDs    map[string]string `json:"correlation_ids,omitempty"`
	PayloadDigests    []digest.Record   `json:"payload_digests,omitempty"`
	ObservedAt        time.Time         `json:"observed_at,omitempty"`
}

type ConfirmationReadObservation struct {
	Version           string            `json:"version"`
	Attempt           AttemptMetadata   `json:"attempt"`
	RequestEvidenceID string            `json:"request_evidence_id,omitempty"`
	Operation         OperationRef      `json:"operation"`
	Outcome           string            `json:"outcome"`
	ProjectedDigests  []digest.Record   `json:"projected_digests,omitempty"`
	CorrelationIDs    map[string]string `json:"correlation_ids,omitempty"`
	ObservedAt        time.Time         `json:"observed_at,omitempty"`
}

func NormalizeExecutionRequest(record ExecutionRequest) ExecutionRequest {
	record.Version = strings.TrimSpace(record.Version)
	record.Attempt = NormalizeAttempt(record.Attempt)
	record.RequestID = strings.TrimSpace(record.RequestID)
	record.Operation = NormalizeOperation(record.Operation)
	record.Transport = normalizeStringMap(record.Transport)
	record.Digests = normalizeDigests(record.Digests)
	record.Metadata = normalizeStringMap(record.Metadata)
	return record
}

func NormalizeExecutionResponse(record ExecutionResponse) ExecutionResponse {
	record.Version = strings.TrimSpace(record.Version)
	record.Attempt = NormalizeAttempt(record.Attempt)
	record.RequestEvidenceID = strings.TrimSpace(record.RequestEvidenceID)
	record.ResponseID = strings.TrimSpace(record.ResponseID)
	record.Operation = NormalizeOperation(record.Operation)
	record.Outcome = strings.ToLower(strings.TrimSpace(record.Outcome))
	record.StatusClass = strings.ToLower(strings.TrimSpace(record.StatusClass))
	record.CorrelationIDs = normalizeStringMap(record.CorrelationIDs)
	record.PayloadDigests = normalizeDigests(record.PayloadDigests)
	record.ErrorSummary = strings.TrimSpace(record.ErrorSummary)
	record.StartedAt = utc(record.StartedAt)
	record.FinishedAt = utc(record.FinishedAt)
	return record
}

func NormalizeStatusObservation(record StatusObservation) StatusObservation {
	record.Version = strings.TrimSpace(record.Version)
	record.Attempt = NormalizeAttempt(record.Attempt)
	record.RequestEvidenceID = strings.TrimSpace(record.RequestEvidenceID)
	record.Operation = NormalizeOperation(record.Operation)
	record.Status = strings.TrimSpace(record.Status)
	record.TerminalityHint = strings.ToLower(strings.TrimSpace(record.TerminalityHint))
	record.CorrelationIDs = normalizeStringMap(record.CorrelationIDs)
	record.PayloadDigests = normalizeDigests(record.PayloadDigests)
	record.ObservedAt = utc(record.ObservedAt)
	return record
}

func NormalizeConfirmationReadObservation(record ConfirmationReadObservation) ConfirmationReadObservation {
	record.Version = strings.TrimSpace(record.Version)
	record.Attempt = NormalizeAttempt(record.Attempt)
	record.RequestEvidenceID = strings.TrimSpace(record.RequestEvidenceID)
	record.Operation = NormalizeOperation(record.Operation)
	record.Outcome = strings.ToLower(strings.TrimSpace(record.Outcome))
	record.ProjectedDigests = normalizeDigests(record.ProjectedDigests)
	record.CorrelationIDs = normalizeStringMap(record.CorrelationIDs)
	record.ObservedAt = utc(record.ObservedAt)
	return record
}

func NormalizeAttempt(attempt AttemptMetadata) AttemptMetadata {
	attempt.EvidenceID = strings.TrimSpace(attempt.EvidenceID)
	attempt.AttemptID = strings.TrimSpace(attempt.AttemptID)
	attempt.Actor = strings.TrimSpace(attempt.Actor)
	attempt.Source = strings.TrimSpace(attempt.Source)
	attempt.RecordedAt = utc(attempt.RecordedAt)
	if attempt.Sequence < 0 {
		attempt.Sequence = 0
	}
	return attempt
}

func NormalizeOperation(operation OperationRef) OperationRef {
	operation.SubjectKind = strings.TrimSpace(operation.SubjectKind)
	operation.SubjectID = strings.TrimSpace(operation.SubjectID)
	operation.Action = strings.ToLower(strings.TrimSpace(operation.Action))
	operation.Method = strings.ToUpper(strings.TrimSpace(operation.Method))
	operation.SourceKind = strings.TrimSpace(operation.SourceKind)
	operation.SourceID = strings.TrimSpace(operation.SourceID)
	operation.SourcePath = strings.TrimSpace(operation.SourcePath)
	operation.OperationID = strings.TrimSpace(operation.OperationID)
	return operation
}

func ValidateExecutionRequest(record ExecutionRequest) []diagnostic.Record {
	record = NormalizeExecutionRequest(record)
	var out []diagnostic.Record
	out = append(out, validateVersion(record.Version, ExecutionRequestVersion, "version")...)
	out = append(out, validateAttempt(record.Attempt)...)
	out = append(out, validateOperation(record.Operation)...)
	return diagnostic.Sort(out)
}

func ValidateExecutionResponse(record ExecutionResponse) []diagnostic.Record {
	record = NormalizeExecutionResponse(record)
	var out []diagnostic.Record
	out = append(out, validateVersion(record.Version, ExecutionResponseVersion, "version")...)
	out = append(out, validateAttempt(record.Attempt)...)
	out = append(out, validateOperation(record.Operation)...)
	if record.Outcome == "" {
		out = append(out, diag("async.outcome_missing", "execution response outcome is required", "outcome"))
	}
	return diagnostic.Sort(out)
}

func ValidateStatusObservation(record StatusObservation) []diagnostic.Record {
	record = NormalizeStatusObservation(record)
	var out []diagnostic.Record
	out = append(out, validateVersion(record.Version, StatusObservationVersion, "version")...)
	out = append(out, validateAttempt(record.Attempt)...)
	out = append(out, validateOperation(record.Operation)...)
	return diagnostic.Sort(out)
}

func ValidateConfirmationReadObservation(record ConfirmationReadObservation) []diagnostic.Record {
	record = NormalizeConfirmationReadObservation(record)
	var out []diagnostic.Record
	out = append(out, validateVersion(record.Version, ConfirmationReadObservationVersion, "version")...)
	out = append(out, validateAttempt(record.Attempt)...)
	out = append(out, validateOperation(record.Operation)...)
	if record.Outcome == "" {
		out = append(out, diag("async.outcome_missing", "confirmation read outcome is required", "outcome"))
	}
	return diagnostic.Sort(out)
}

func DigestExecutionRequest(record ExecutionRequest) (digest.Record, error) {
	return digestNormalized(NormalizeExecutionRequest(record))
}

func DigestExecutionResponse(record ExecutionResponse) (digest.Record, error) {
	return digestNormalized(NormalizeExecutionResponse(record))
}

func DigestStatusObservation(record StatusObservation) (digest.Record, error) {
	return digestNormalized(NormalizeStatusObservation(record))
}

func DigestConfirmationReadObservation(record ConfirmationReadObservation) (digest.Record, error) {
	return digestNormalized(NormalizeConfirmationReadObservation(record))
}

func validateVersion(got, want, field string) []diagnostic.Record {
	if got == want {
		return nil
	}
	return []diagnostic.Record{diag("async.version_invalid", fmt.Sprintf("async evidence version must be %s", want), field)}
}

func validateAttempt(attempt AttemptMetadata) []diagnostic.Record {
	var out []diagnostic.Record
	if attempt.EvidenceID == "" {
		out = append(out, diag("async.evidence_id_missing", "evidence_id is required", "attempt.evidence_id"))
	}
	if attempt.AttemptID == "" {
		out = append(out, diag("async.attempt_id_missing", "attempt_id is required", "attempt.attempt_id"))
	}
	return out
}

func validateOperation(operation OperationRef) []diagnostic.Record {
	if operation.OperationID == "" {
		return []diagnostic.Record{diag("async.operation_id_missing", "operation_id is required", "operation.operation_id")}
	}
	return nil
}

func diag(code, message, field string) diagnostic.Record {
	return diagnostic.Record{
		Code:     code,
		Severity: diagnostic.SeverityError,
		Message:  message,
		Location: diagnostic.Location{
			Field: field,
		},
	}
}

func normalizeStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeDigests(records []digest.Record) []digest.Record {
	out := make([]digest.Record, 0, len(records))
	for _, record := range records {
		record.Algorithm = strings.TrimSpace(record.Algorithm)
		record.Value = strings.TrimSpace(record.Value)
		if !record.IsZero() {
			out = append(out, record)
		}
	}
	slices.SortFunc(out, func(a, b digest.Record) int {
		return strings.Compare(a.String(), b.String())
	})
	return out
}

func digestNormalized(record any) (digest.Record, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return digest.Record{}, err
	}
	return digest.SHA256Bytes(data), nil
}

func utc(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC()
}
