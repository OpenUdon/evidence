package diagnostic

import (
	"slices"
	"testing"
)

func TestNormalizeSeverity(t *testing.T) {
	tests := map[string]string{
		"":            SeverityWarning,
		"WARN":        SeverityWarning,
		"information": SeverityInfo,
		"blocker":     SeverityBlocking,
		"advice":      SeverityAdvisory,
		"custom":      SeverityWarning,
	}
	for input, want := range tests {
		if got := NormalizeSeverity(input); got != want {
			t.Fatalf("NormalizeSeverity(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestHasErrorsTreatsBlockingAsError(t *testing.T) {
	if !HasErrors([]Record{{Severity: SeverityBlocking}}) {
		t.Fatal("blocking severity should count as an error")
	}
	if HasErrors([]Record{{Severity: SeverityWarning}, {Severity: SeverityInfo}}) {
		t.Fatal("warning/info diagnostics should not count as errors")
	}
}

func TestSortNormalizesAndOrdersDiagnostics(t *testing.T) {
	records := []Record{
		{Code: "z.info", Severity: "info", Message: "later"},
		{Code: "b.warn", Severity: "warning", Message: "middle"},
		{Code: "a.err", Severity: "ERROR", Message: "first", Location: Location{Path: "b.txt"}},
		{Code: "a.err", Severity: "error", Message: "first", Location: Location{Path: "a.txt"}},
	}
	got := Sort(records)
	codes := []string{got[0].Code + ":" + got[0].Location.Path, got[1].Code + ":" + got[1].Location.Path, got[2].Code, got[3].Code}
	want := []string{"a.err:a.txt", "a.err:b.txt", "b.warn", "z.info"}
	if !slices.Equal(codes, want) {
		t.Fatalf("order = %#v, want %#v", codes, want)
	}
	if got[0].Severity != SeverityError {
		t.Fatalf("severity was not normalized: %#v", got[0])
	}
}

func TestNormalizeTrimsFieldsAndClearsNegativeLocation(t *testing.T) {
	got := Normalize(Record{
		Code:        " code ",
		Severity:    "",
		Message:     " message ",
		Remediation: " fix ",
		Location:    Location{Path: " artifact.json ", Line: -1, Column: -2},
	})
	if got.Code != "code" || got.Message != "message" || got.Remediation != "fix" {
		t.Fatalf("record was not trimmed: %#v", got)
	}
	if got.Severity != SeverityWarning || got.Location.Path != "artifact.json" || got.Location.Line != 0 || got.Location.Column != 0 {
		t.Fatalf("record was not normalized: %#v", got)
	}
}
