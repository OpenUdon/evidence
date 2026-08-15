package artifact

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/OpenUdon/evidence/digest"
)

func TestDescriptorNormalizeValidateAndDigest(t *testing.T) {
	value := Descriptor{
		Version:   "  " + DescriptorVersion + " ",
		MediaType: " Application/Vnd.Example+JSON; Charset=UTF-8 ",
		SizeBytes: 42,
		Digest:    digest.SHA256Bytes([]byte("content")),
		Annotations: map[string]string{
			" title ": " Example ",
		},
	}
	original := value.Annotations[" title "]
	normalized := NormalizeDescriptor(value)
	if normalized.Version != DescriptorVersion || normalized.MediaType != "application/vnd.example+json; charset=UTF-8" {
		t.Fatalf("normalized descriptor = %#v", normalized)
	}
	if normalized.Annotations["title"] != "Example" || value.Annotations[" title "] != original {
		t.Fatalf("annotation normalization mutated input: input=%#v output=%#v", value.Annotations, normalized.Annotations)
	}
	if err := ValidateDescriptor(normalized); err != nil {
		t.Fatal(err)
	}
	first, err := DigestDescriptor(value)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestDescriptor(normalized)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("digest changed after normalization: %s != %s", first.String(), second.String())
	}
	reordered := normalized
	reordered.Annotations = map[string]string{"z": "last", "a": "first"}
	other := normalized
	other.Annotations = map[string]string{}
	other.Annotations["a"] = "first"
	other.Annotations["z"] = "last"
	reorderedDigest, err := DigestDescriptor(reordered)
	if err != nil {
		t.Fatal(err)
	}
	otherDigest, err := DigestDescriptor(other)
	if err != nil {
		t.Fatal(err)
	}
	if reorderedDigest != otherDigest {
		t.Fatalf("annotation insertion order changed digest: %s != %s", reorderedDigest.String(), otherDigest.String())
	}
}

func TestDescriptorValidationFailures(t *testing.T) {
	valid := testDescriptor("content")
	tests := map[string]func(*Descriptor){
		"version":    func(value *Descriptor) { value.Version = "other" },
		"media type": func(value *Descriptor) { value.MediaType = "not-a-media-type" },
		"size":       func(value *Descriptor) { value.SizeBytes = -1 },
		"algorithm":  func(value *Descriptor) { value.Digest.Algorithm = "sha512" },
		"digest":     func(value *Descriptor) { value.Digest.Value = "xyz" },
		"annotation": func(value *Descriptor) { value.Annotations = map[string]string{" ": "value"} },
		"annotation collision": func(value *Descriptor) {
			value.Annotations = map[string]string{"title": "one", " title ": "two"}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if err := ValidateDescriptor(value); err == nil {
				t.Fatalf("invalid descriptor unexpectedly passed: %#v", value)
			}
		})
	}
}

func TestAssessmentNormalizesSortsAndRoundTrips(t *testing.T) {
	assessedAt := time.Date(2026, 8, 15, 1, 2, 3, 4, time.FixedZone("offset", 3600))
	value := Assessment{
		Version:    " " + AssessmentVersion + " ",
		Subject:    testDescriptor("subject"),
		Status:     LifecycleStatus(" ACTIVE "),
		AssessedAt: assessedAt,
		ExpiresAt:  assessedAt.Add(24 * time.Hour),
		Supporting: []Descriptor{testDescriptor("z"), testDescriptor("a")},
	}
	original := append([]Descriptor(nil), value.Supporting...)
	normalized := NormalizeAssessment(value)
	if normalized.Status != LifecycleActive || normalized.AssessedAt.Location() != time.UTC {
		t.Fatalf("normalized assessment = %#v", normalized)
	}
	if !slices.EqualFunc(value.Supporting, original, func(a, b Descriptor) bool { return a.Digest == b.Digest }) {
		t.Fatal("NormalizeAssessment mutated supporting input order")
	}
	if strings.Compare(normalized.Supporting[0].Digest.String(), normalized.Supporting[1].Digest.String()) >= 0 {
		t.Fatalf("supporting descriptors not sorted: %#v", normalized.Supporting)
	}
	data, err := CanonicalAssessmentJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip Assessment
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAssessment(roundTrip); err != nil {
		t.Fatal(err)
	}
	activeWithoutOptionals := value
	activeWithoutOptionals.ExpiresAt = time.Time{}
	activeWithoutOptionals.Supporting = nil
	minimalJSON, err := CanonicalAssessmentJSON(activeWithoutOptionals)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(minimalJSON), "expires_at") || strings.Contains(string(minimalJSON), "successor") {
		t.Fatalf("zero optional lifecycle fields were serialized: %s", minimalJSON)
	}
}

func TestAssessmentLifecycleRules(t *testing.T) {
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	valid := Assessment{
		Version:    AssessmentVersion,
		Subject:    testDescriptor("subject"),
		Status:     LifecycleActive,
		AssessedAt: now,
		ExpiresAt:  now.Add(time.Hour),
	}
	if err := ValidateAssessment(valid); err != nil {
		t.Fatal(err)
	}
	if got := EffectiveStatus(valid, now.Add(time.Hour)); got != LifecycleStale {
		t.Fatalf("EffectiveStatus = %q, want stale", got)
	}
	if got := EffectiveStatus(valid, time.Time{}); got != LifecycleActive {
		t.Fatalf("zero-clock EffectiveStatus = %q, want active", got)
	}

	superseded := valid
	superseded.Status = LifecycleSuperseded
	superseded.Successor = digest.SHA256Bytes([]byte("successor"))
	if err := ValidateAssessment(superseded); err != nil {
		t.Fatal(err)
	}
	superseded.Successor = superseded.Subject.Digest
	if err := ValidateAssessment(superseded); err == nil {
		t.Fatal("self-supersession unexpectedly passed")
	}
}

func TestAssessmentValidationFailures(t *testing.T) {
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	valid := Assessment{Version: AssessmentVersion, Subject: testDescriptor("subject"), Status: LifecycleActive, AssessedAt: now}
	tests := map[string]func(*Assessment){
		"version":         func(value *Assessment) { value.Version = "other" },
		"subject":         func(value *Assessment) { value.Subject.Digest.Value = "bad" },
		"status":          func(value *Assessment) { value.Status = "unknown" },
		"assessed at":     func(value *Assessment) { value.AssessedAt = time.Time{} },
		"time order":      func(value *Assessment) { value.ExpiresAt = now },
		"extra successor": func(value *Assessment) { value.Successor = digest.SHA256Bytes([]byte("next")) },
		"duplicate": func(value *Assessment) {
			value.Supporting = []Descriptor{testDescriptor("same"), testDescriptor("same")}
		},
		"subject repeated": func(value *Assessment) { value.Supporting = []Descriptor{value.Subject} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if err := ValidateAssessment(value); err == nil {
				t.Fatalf("invalid assessment unexpectedly passed: %#v", value)
			}
		})
	}

	missingSuccessor := valid
	missingSuccessor.Status = LifecycleSuperseded
	if err := ValidateAssessment(missingSuccessor); err == nil {
		t.Fatal("superseded assessment without successor unexpectedly passed")
	}
}

func testDescriptor(content string) Descriptor {
	return Descriptor{
		Version:   DescriptorVersion,
		MediaType: "application/json",
		SizeBytes: int64(len(content)),
		Digest:    digest.SHA256Bytes([]byte(content)),
	}
}
