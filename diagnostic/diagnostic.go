// Package diagnostic provides product-neutral diagnostic records and helpers.
package diagnostic

import (
	"cmp"
	"encoding/json"
	"slices"
	"strings"
)

const (
	SeverityError    = "error"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
	SeverityBlocking = "blocking"
	SeverityAdvisory = "advisory"
)

// Location identifies the artifact, resource, source, or field tied to a
// diagnostic without requiring a product-specific type.
type Location struct {
	Path          string `json:"path,omitempty"`
	Line          int    `json:"line,omitempty"`
	Column        int    `json:"column,omitempty"`
	Address       string `json:"address,omitempty"`
	ModuleAddress string `json:"module_address,omitempty"`
	APISourceKind string `json:"api_source_kind,omitempty"`
	APISourceID   string `json:"api_source_id,omitempty"`
	OperationID   string `json:"operation_id,omitempty"`
	Field         string `json:"field,omitempty"`
}

// Record describes one validation, review, authoring, or runtime evidence
// issue in a product-neutral shape.
type Record struct {
	Code        string         `json:"code"`
	Severity    string         `json:"severity"`
	Message     string         `json:"message"`
	Location    Location       `json:"location,omitempty"`
	Remediation string         `json:"remediation,omitempty"`
	Detail      map[string]any `json:"detail,omitempty"`
}

// NormalizeSeverity returns a stable lowercase severity. Empty and unknown
// severities are treated as warnings, which is the safest non-success default.
func NormalizeSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case SeverityError:
		return SeverityError
	case SeverityWarning, "warn":
		return SeverityWarning
	case SeverityInfo, "information":
		return SeverityInfo
	case SeverityBlocking, "blocker":
		return SeverityBlocking
	case SeverityAdvisory, "advice":
		return SeverityAdvisory
	default:
		return SeverityWarning
	}
}

// IsError reports whether severity should block successful validation.
func IsError(severity string) bool {
	normalized := NormalizeSeverity(severity)
	return normalized == SeverityError || normalized == SeverityBlocking
}

// HasErrors reports whether records contains any blocking/error diagnostic.
func HasErrors(records []Record) bool {
	for _, record := range records {
		if IsError(record.Severity) {
			return true
		}
	}
	return false
}

// Normalize returns a trimmed copy of record with normalized severity.
func Normalize(record Record) Record {
	record.Code = strings.TrimSpace(record.Code)
	record.Severity = NormalizeSeverity(record.Severity)
	record.Message = strings.TrimSpace(record.Message)
	record.Remediation = strings.TrimSpace(record.Remediation)
	record.Location = normalizeLocation(record.Location)
	if len(record.Detail) == 0 {
		record.Detail = nil
	}
	return record
}

// NormalizeAll returns normalized records without mutating the input slice.
func NormalizeAll(records []Record) []Record {
	out := make([]Record, len(records))
	for i, record := range records {
		out[i] = Normalize(record)
	}
	return out
}

// Sort returns normalized diagnostics ordered by severity rank, code,
// location, message, remediation, and detail.
func Sort(records []Record) []Record {
	out := NormalizeAll(records)
	slices.SortStableFunc(out, Compare)
	return out
}

// Compare orders normalized or unnormalized diagnostic records deterministically.
func Compare(a, b Record) int {
	a = Normalize(a)
	b = Normalize(b)
	if diff := cmp.Compare(severityRank(a.Severity), severityRank(b.Severity)); diff != 0 {
		return diff
	}
	for _, pair := range [][2]string{
		{a.Code, b.Code},
		{a.Location.Path, b.Location.Path},
		{a.Location.Address, b.Location.Address},
		{a.Location.ModuleAddress, b.Location.ModuleAddress},
		{a.Location.APISourceKind, b.Location.APISourceKind},
		{a.Location.APISourceID, b.Location.APISourceID},
		{a.Location.OperationID, b.Location.OperationID},
		{a.Location.Field, b.Location.Field},
		{a.Message, b.Message},
		{a.Remediation, b.Remediation},
		{detailKey(a.Detail), detailKey(b.Detail)},
	} {
		if diff := cmp.Compare(pair[0], pair[1]); diff != 0 {
			return diff
		}
	}
	if diff := cmp.Compare(a.Location.Line, b.Location.Line); diff != 0 {
		return diff
	}
	return cmp.Compare(a.Location.Column, b.Location.Column)
}

func severityRank(severity string) int {
	switch NormalizeSeverity(severity) {
	case SeverityError, SeverityBlocking:
		return 0
	case SeverityWarning:
		return 1
	case SeverityAdvisory:
		return 2
	case SeverityInfo:
		return 3
	default:
		return 4
	}
}

func normalizeLocation(location Location) Location {
	location.Path = strings.TrimSpace(location.Path)
	location.Address = strings.TrimSpace(location.Address)
	location.ModuleAddress = strings.TrimSpace(location.ModuleAddress)
	location.APISourceKind = strings.TrimSpace(location.APISourceKind)
	location.APISourceID = strings.TrimSpace(location.APISourceID)
	location.OperationID = strings.TrimSpace(location.OperationID)
	location.Field = strings.TrimSpace(location.Field)
	if location.Line < 0 {
		location.Line = 0
	}
	if location.Column < 0 {
		location.Column = 0
	}
	return location
}

func detailKey(detail map[string]any) string {
	if len(detail) == 0 {
		return ""
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return ""
	}
	return string(data)
}
