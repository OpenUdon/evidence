// Package approval provides product-neutral approval evidence primitives.
package approval

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/OpenUdon/evidence/artifact"
	"github.com/OpenUdon/evidence/diagnostic"
	"github.com/OpenUdon/evidence/digest"
)

const (
	RecordVersion = "evidence.approval.v1"
)

// Subject identifies the artifact, package, plan, run, or other review target
// that approval evidence is bound to.
type Subject struct {
	Kind      string            `json:"kind,omitempty"`
	ID        string            `json:"id"`
	Digest    digest.Record     `json:"digest,omitempty"`
	Artifacts []artifact.Record `json:"artifacts,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Approver records one human or system identity that approved a subject.
type Approver struct {
	Identity   string    `json:"identity"`
	Role       string    `json:"role,omitempty"`
	Context    string    `json:"context,omitempty"`
	ApprovedAt time.Time `json:"approved_at"`
}

// Requirement describes the approval count and optional role constraints that
// must be satisfied before a subject may cross a trusted boundary.
type Requirement struct {
	ID           string    `json:"id"`
	Reason       string    `json:"reason,omitempty"`
	SubjectKind  string    `json:"subject_kind,omitempty"`
	SubjectID    string    `json:"subject_id,omitempty"`
	Address      string    `json:"address,omitempty"`
	Action       string    `json:"action,omitempty"`
	MinApprovals int       `json:"min_approvals"`
	Roles        []string  `json:"roles,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// Decision is a deterministic evaluation result for one requirement.
type Decision struct {
	RequirementID string     `json:"requirement_id"`
	Satisfied     bool       `json:"satisfied"`
	Needed        int        `json:"needed"`
	Matched       []Approver `json:"matched,omitempty"`
	Missing       int        `json:"missing,omitempty"`
	Expired       bool       `json:"expired,omitempty"`
}

// Record groups approval evidence for one subject.
type Record struct {
	Version      string        `json:"version"`
	Subject      Subject       `json:"subject"`
	Requirements []Requirement `json:"requirements,omitempty"`
	Approvers    []Approver    `json:"approvers,omitempty"`
	Decisions    []Decision    `json:"decisions,omitempty"`
}

// ValidationOptions controls validation and requirement evaluation.
type ValidationOptions struct {
	Now                  time.Time
	RequireSubjectDigest bool
}

// NormalizeApprover trims identity metadata and normalizes approval time to UTC.
func NormalizeApprover(approver Approver) (Approver, error) {
	approver.Identity = strings.TrimSpace(approver.Identity)
	approver.Role = strings.TrimSpace(approver.Role)
	approver.Context = strings.TrimSpace(approver.Context)
	if approver.Identity == "" {
		return Approver{}, fmt.Errorf("approver identity is required")
	}
	if approver.ApprovedAt.IsZero() {
		return Approver{}, fmt.Errorf("approver approved_at is required")
	}
	approver.ApprovedAt = approver.ApprovedAt.UTC()
	return approver, nil
}

// RequirementsSatisfied reports whether approvers satisfy all requirements.
func RequirementsSatisfied(requirements []Requirement, approvers []Approver, now ...time.Time) error {
	decisions := EvaluateRequirements(requirements, approvers, firstTime(now...))
	for _, decision := range decisions {
		if !decision.Satisfied {
			if decision.Expired {
				return fmt.Errorf("approval requirement %s is expired", decision.RequirementID)
			}
			return fmt.Errorf("approval requirement %s needs %d approver(s)", decision.RequirementID, decision.Needed)
		}
	}
	return nil
}

// EvaluateRequirements returns deterministic decisions for requirements.
func EvaluateRequirements(requirements []Requirement, approvers []Approver, now time.Time) []Decision {
	requirements = NormalizeRequirements(requirements)
	approvers = NormalizeApprovers(approvers)
	decisions := make([]Decision, 0, len(requirements))
	for _, req := range requirements {
		needed := req.MinApprovals
		if needed <= 0 {
			needed = 1
		}
		decision := Decision{RequirementID: req.ID, Needed: needed}
		if !req.ExpiresAt.IsZero() && !now.IsZero() && now.After(req.ExpiresAt.UTC()) {
			decision.Expired = true
			decision.Missing = needed
			decisions = append(decisions, decision)
			continue
		}
		seen := map[string]bool{}
		for _, approver := range approvers {
			if approver.Identity == "" {
				continue
			}
			if seen[approver.Identity] {
				continue
			}
			if len(req.Roles) > 0 && !contains(req.Roles, approver.Role) {
				continue
			}
			seen[approver.Identity] = true
			decision.Matched = append(decision.Matched, approver)
		}
		if len(decision.Matched) >= needed {
			decision.Satisfied = true
			decision.Missing = 0
		} else {
			decision.Missing = needed - len(decision.Matched)
		}
		decisions = append(decisions, decision)
	}
	return decisions
}

// DigestRecord returns the digest of a normalized approval record without
// mutating the caller's value.
func DigestRecord(record Record) (digest.Record, error) {
	normalized := NormalizeRecord(record)
	data, err := json.Marshal(normalized)
	if err != nil {
		return digest.Record{}, err
	}
	return digest.SHA256Bytes(data), nil
}

// ValidateRecord returns diagnostics for malformed or unsatisfied approval
// evidence. Diagnostics are sorted deterministically.
func ValidateRecord(record Record, opts ValidationOptions) []diagnostic.Record {
	record = NormalizeRecord(record)
	var records []diagnostic.Record
	if record.Version != RecordVersion {
		records = append(records, diag("approval.version_invalid", "approval record version is unsupported", "version"))
	}
	if record.Subject.ID == "" {
		records = append(records, diag("approval.subject_missing", "approval subject id is required", "subject.id"))
	}
	if opts.RequireSubjectDigest && record.Subject.Digest.IsZero() {
		records = append(records, diag("approval.subject_digest_missing", "approval subject digest is required", "subject.digest"))
	}
	for i, approver := range record.Approvers {
		if _, err := NormalizeApprover(approver); err != nil {
			records = append(records, diag("approval.approver_invalid", err.Error(), fmt.Sprintf("approvers[%d]", i)))
		}
	}
	for i, req := range record.Requirements {
		if req.ID == "" {
			records = append(records, diag("approval.requirement_id_missing", "approval requirement id is required", fmt.Sprintf("requirements[%d].id", i)))
		}
		if req.MinApprovals < 0 {
			records = append(records, diag("approval.requirement_min_invalid", "approval requirement min_approvals must not be negative", fmt.Sprintf("requirements[%d].min_approvals", i)))
		}
	}
	for _, decision := range EvaluateRequirements(record.Requirements, record.Approvers, opts.Now) {
		if decision.Satisfied {
			continue
		}
		code := "approval.requirement_unsatisfied"
		message := fmt.Sprintf("approval requirement %s needs %d approver(s)", decision.RequirementID, decision.Needed)
		if decision.Expired {
			code = "approval.requirement_expired"
			message = fmt.Sprintf("approval requirement %s is expired", decision.RequirementID)
		}
		records = append(records, diagnostic.Record{
			Code:     code,
			Severity: diagnostic.SeverityError,
			Message:  message,
			Location: diagnostic.Location{
				Field: "requirements",
			},
		})
	}
	return diagnostic.Sort(records)
}

// NormalizeRecord returns a deterministic, trimmed approval record.
func NormalizeRecord(record Record) Record {
	record.Version = strings.TrimSpace(record.Version)
	if record.Version == "" {
		record.Version = RecordVersion
	}
	record.Subject = NormalizeSubject(record.Subject)
	record.Requirements = NormalizeRequirements(record.Requirements)
	record.Approvers = NormalizeApprovers(record.Approvers)
	record.Decisions = NormalizeDecisions(record.Decisions)
	return record
}

func NormalizeSubject(subject Subject) Subject {
	subject.Kind = strings.TrimSpace(subject.Kind)
	subject.ID = strings.TrimSpace(subject.ID)
	if len(subject.Artifacts) > 0 {
		subject.Artifacts = append([]artifact.Record(nil), subject.Artifacts...)
		slices.SortStableFunc(subject.Artifacts, func(a, b artifact.Record) int {
			return strings.Compare(a.Path, b.Path)
		})
	}
	if len(subject.Metadata) == 0 {
		subject.Metadata = nil
	}
	return subject
}

func NormalizeApprovers(approvers []Approver) []Approver {
	out := make([]Approver, 0, len(approvers))
	for _, approver := range approvers {
		normalized, err := NormalizeApprover(approver)
		if err == nil {
			out = append(out, normalized)
		} else {
			approver.Identity = strings.TrimSpace(approver.Identity)
			approver.Role = strings.TrimSpace(approver.Role)
			approver.Context = strings.TrimSpace(approver.Context)
			out = append(out, approver)
		}
	}
	slices.SortStableFunc(out, func(a, b Approver) int {
		if diff := strings.Compare(a.Identity, b.Identity); diff != 0 {
			return diff
		}
		if diff := strings.Compare(a.Role, b.Role); diff != 0 {
			return diff
		}
		if diff := strings.Compare(a.Context, b.Context); diff != 0 {
			return diff
		}
		return a.ApprovedAt.Compare(b.ApprovedAt)
	})
	return out
}

func NormalizeRequirements(requirements []Requirement) []Requirement {
	out := make([]Requirement, 0, len(requirements))
	for _, req := range requirements {
		req.ID = strings.TrimSpace(req.ID)
		req.Reason = strings.TrimSpace(req.Reason)
		req.SubjectKind = strings.TrimSpace(req.SubjectKind)
		req.SubjectID = strings.TrimSpace(req.SubjectID)
		req.Address = strings.TrimSpace(req.Address)
		req.Action = strings.TrimSpace(req.Action)
		req.Roles = trimList(req.Roles)
		if !req.ExpiresAt.IsZero() {
			req.ExpiresAt = req.ExpiresAt.UTC()
		}
		out = append(out, req)
	}
	slices.SortStableFunc(out, func(a, b Requirement) int {
		return strings.Compare(a.ID+a.SubjectID+a.Address+a.Action, b.ID+b.SubjectID+b.Address+b.Action)
	})
	return out
}

func NormalizeDecisions(decisions []Decision) []Decision {
	out := make([]Decision, 0, len(decisions))
	for _, decision := range decisions {
		decision.RequirementID = strings.TrimSpace(decision.RequirementID)
		decision.Matched = NormalizeApprovers(decision.Matched)
		out = append(out, decision)
	}
	slices.SortStableFunc(out, func(a, b Decision) int {
		return strings.Compare(a.RequirementID, b.RequirementID)
	})
	return out
}

func trimList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	slices.Sort(out)
	return out
}

func contains(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func firstTime(values ...time.Time) time.Time {
	if len(values) == 0 {
		return time.Now().UTC()
	}
	return values[0].UTC()
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
