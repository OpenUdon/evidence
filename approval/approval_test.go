package approval

import (
	"slices"
	"testing"
	"time"

	"github.com/OpenUdon/evidence/digest"
)

func TestNormalizeApprover(t *testing.T) {
	when := time.Date(2026, 5, 30, 12, 0, 0, 0, time.FixedZone("test", -7*60*60))
	got, err := NormalizeApprover(Approver{Identity: " alice@example.com ", Role: " reviewer ", Context: " ticket-1 ", ApprovedAt: when})
	if err != nil {
		t.Fatal(err)
	}
	if got.Identity != "alice@example.com" || got.Role != "reviewer" || got.Context != "ticket-1" {
		t.Fatalf("approver was not trimmed: %#v", got)
	}
	if got.ApprovedAt.Location() != time.UTC {
		t.Fatalf("approved_at was not normalized to UTC: %s", got.ApprovedAt.Location())
	}
	if _, err := NormalizeApprover(Approver{ApprovedAt: when}); err == nil {
		t.Fatalf("missing identity unexpectedly succeeded")
	}
	if _, err := NormalizeApprover(Approver{Identity: "alice"}); err == nil {
		t.Fatalf("missing approved_at unexpectedly succeeded")
	}
}

func TestRequirementsSatisfiedFiltersRolesAndMinApprovals(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	requirements := []Requirement{{
		ID:           "prod-change",
		MinApprovals: 2,
		Roles:        []string{"security", "ops"},
	}}
	approvers := []Approver{
		{Identity: "alice", Role: "ops", ApprovedAt: now},
		{Identity: "alice", Role: "ops", ApprovedAt: now.Add(time.Minute)},
		{Identity: "bob", Role: "viewer", ApprovedAt: now},
	}
	if err := RequirementsSatisfied(requirements, approvers, now); err == nil {
		t.Fatalf("duplicate/wrong-role approvers unexpectedly satisfied requirement")
	}
	approvers = append(approvers, Approver{Identity: "carol", Role: "security", ApprovedAt: now})
	if err := RequirementsSatisfied(requirements, approvers, now); err != nil {
		t.Fatalf("requirement was not satisfied: %v", err)
	}
}

func TestRequirementsSatisfiedExpiry(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	requirements := []Requirement{{
		ID:           "short-lived",
		MinApprovals: 1,
		ExpiresAt:    now.Add(-time.Minute),
	}}
	approvers := []Approver{{Identity: "alice", ApprovedAt: now.Add(-2 * time.Minute)}}
	if err := RequirementsSatisfied(requirements, approvers, now); err == nil {
		t.Fatalf("expired approval requirement unexpectedly succeeded")
	}
}

func TestDigestRecordIsDeterministic(t *testing.T) {
	when := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	a := Record{
		Subject: Subject{ID: "plan", Kind: "ramen-plan", Digest: digest.SHA256Bytes([]byte("plan"))},
		Requirements: []Requirement{
			{ID: "b", Roles: []string{"ops", "security"}},
			{ID: "a", Roles: []string{"security", "ops"}},
		},
		Approvers: []Approver{
			{Identity: "carol", Role: "security", ApprovedAt: when},
			{Identity: "alice", Role: "ops", ApprovedAt: when},
		},
	}
	b := Record{
		Version: RecordVersion,
		Subject: Subject{Kind: "ramen-plan", ID: "plan", Digest: digest.SHA256Bytes([]byte("plan"))},
		Approvers: []Approver{
			{Identity: "alice", Role: "ops", ApprovedAt: when},
			{Identity: "carol", Role: "security", ApprovedAt: when},
		},
		Requirements: []Requirement{
			{ID: "a", Roles: []string{"ops", "security"}},
			{ID: "b", Roles: []string{"security", "ops"}},
		},
	}
	digestA, err := DigestRecord(a)
	if err != nil {
		t.Fatal(err)
	}
	digestB, err := DigestRecord(b)
	if err != nil {
		t.Fatal(err)
	}
	if digestA != digestB {
		t.Fatalf("digests differ:\n%s\n%s", digestA.String(), digestB.String())
	}
}

func TestValidateRecordDiagnostics(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	records := ValidateRecord(Record{
		Version: "bad",
		Subject: Subject{},
		Requirements: []Requirement{{
			ID:           "needs-ops",
			MinApprovals: 1,
			Roles:        []string{"ops"},
		}},
		Approvers: []Approver{{Identity: " ", Role: "ops", ApprovedAt: now}},
	}, ValidationOptions{Now: now, RequireSubjectDigest: true})
	var codes []string
	for _, record := range records {
		codes = append(codes, record.Code)
	}
	for _, code := range []string{
		"approval.approver_invalid",
		"approval.requirement_unsatisfied",
		"approval.subject_digest_missing",
		"approval.subject_missing",
		"approval.version_invalid",
	} {
		if !slices.Contains(codes, code) {
			t.Fatalf("missing diagnostic %s in %#v", code, codes)
		}
	}
}
