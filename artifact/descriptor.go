package artifact

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"slices"
	"strings"
	"time"

	"github.com/OpenUdon/evidence/digest"
)

const (
	// DescriptorVersion is the wire version for content-addressed artifact
	// descriptors.
	DescriptorVersion = "evidence.artifact-descriptor.v1"
	// AssessmentVersion is the wire version for artifact lifecycle assessments.
	AssessmentVersion = "evidence.artifact-assessment.v1"
)

// LifecycleStatus is a caller-supplied assessment of an artifact's current
// distribution lifecycle. It describes evidence and does not grant trust.
type LifecycleStatus string

const (
	LifecycleActive     LifecycleStatus = "active"
	LifecycleStale      LifecycleStatus = "stale"
	LifecycleRevoked    LifecycleStatus = "revoked"
	LifecycleSuperseded LifecycleStatus = "superseded"
)

// Descriptor identifies exact artifact bytes without naming a product,
// storage location, registry, or trust policy.
type Descriptor struct {
	Version     string            `json:"version"`
	MediaType   string            `json:"media_type"`
	SizeBytes   int64             `json:"size_bytes"`
	Digest      digest.Record     `json:"digest"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Assessment binds lifecycle evidence to exact subject bytes. Successor is
// required only for superseded subjects. Supporting descriptors are inert
// evidence references selected by the caller.
type Assessment struct {
	Version    string          `json:"version"`
	Subject    Descriptor      `json:"subject"`
	Status     LifecycleStatus `json:"status"`
	AssessedAt time.Time       `json:"assessed_at"`
	ExpiresAt  time.Time       `json:"expires_at,omitzero"`
	Successor  digest.Record   `json:"successor,omitzero"`
	Supporting []Descriptor    `json:"supporting,omitempty"`
}

// NormalizeDescriptor returns a deterministic copy without mutating input.
func NormalizeDescriptor(value Descriptor) Descriptor {
	value.Version = strings.TrimSpace(value.Version)
	value.MediaType = normalizeMediaType(value.MediaType)
	value.Digest.Algorithm = strings.ToLower(strings.TrimSpace(value.Digest.Algorithm))
	value.Digest.Value = strings.ToLower(strings.TrimSpace(value.Digest.Value))
	value.Annotations = normalizeAnnotations(value.Annotations)
	return value
}

// ValidateDescriptor rejects malformed descriptor fields. Only SHA-256 is
// currently supported so content identity stays unambiguous across consumers.
func ValidateDescriptor(value Descriptor) error {
	if err := validateAnnotationKeys(value.Annotations); err != nil {
		return err
	}
	value = NormalizeDescriptor(value)
	if value.Version != DescriptorVersion {
		return fmt.Errorf("artifact descriptor version must be %q", DescriptorVersion)
	}
	if value.MediaType == "" {
		return fmt.Errorf("artifact descriptor media_type is required")
	}
	parsed, _, err := mime.ParseMediaType(value.MediaType)
	if err != nil || !strings.Contains(parsed, "/") {
		return fmt.Errorf("artifact descriptor media_type %q is invalid", value.MediaType)
	}
	if value.SizeBytes < 0 {
		return fmt.Errorf("artifact descriptor size_bytes must not be negative")
	}
	if err := validateSHA256(value.Digest, "artifact descriptor digest"); err != nil {
		return err
	}
	return nil
}

// CanonicalDescriptorJSON returns deterministic normalized JSON.
func CanonicalDescriptorJSON(value Descriptor) ([]byte, error) {
	if err := ValidateDescriptor(value); err != nil {
		return nil, err
	}
	value = NormalizeDescriptor(value)
	return json.Marshal(value)
}

// DigestDescriptor returns the SHA-256 digest of canonical descriptor JSON.
func DigestDescriptor(value Descriptor) (digest.Record, error) {
	data, err := CanonicalDescriptorJSON(value)
	if err != nil {
		return digest.Record{}, err
	}
	return digest.SHA256Bytes(data), nil
}

// NormalizeAssessment returns a deterministic copy without mutating input.
func NormalizeAssessment(value Assessment) Assessment {
	value.Version = strings.TrimSpace(value.Version)
	value.Subject = NormalizeDescriptor(value.Subject)
	value.Status = LifecycleStatus(strings.ToLower(strings.TrimSpace(string(value.Status))))
	value.AssessedAt = normalizeTime(value.AssessedAt)
	value.ExpiresAt = normalizeTime(value.ExpiresAt)
	value.Successor.Algorithm = strings.ToLower(strings.TrimSpace(value.Successor.Algorithm))
	value.Successor.Value = strings.ToLower(strings.TrimSpace(value.Successor.Value))
	value.Supporting = append([]Descriptor(nil), value.Supporting...)
	for index := range value.Supporting {
		value.Supporting[index] = NormalizeDescriptor(value.Supporting[index])
	}
	slices.SortStableFunc(value.Supporting, compareDescriptor)
	return value
}

// ValidateAssessment rejects malformed or internally inconsistent lifecycle
// evidence. Trust in the caller that supplied the assessment remains external.
func ValidateAssessment(value Assessment) error {
	if err := ValidateDescriptor(value.Subject); err != nil {
		return fmt.Errorf("artifact assessment subject: %w", err)
	}
	for index, supporting := range value.Supporting {
		if err := ValidateDescriptor(supporting); err != nil {
			return fmt.Errorf("artifact assessment supporting[%d]: %w", index, err)
		}
	}
	value = NormalizeAssessment(value)
	if value.Version != AssessmentVersion {
		return fmt.Errorf("artifact assessment version must be %q", AssessmentVersion)
	}
	switch value.Status {
	case LifecycleActive, LifecycleStale, LifecycleRevoked, LifecycleSuperseded:
	default:
		return fmt.Errorf("artifact assessment status %q is invalid", value.Status)
	}
	if value.AssessedAt.IsZero() {
		return fmt.Errorf("artifact assessment assessed_at is required")
	}
	if !value.ExpiresAt.IsZero() && !value.ExpiresAt.After(value.AssessedAt) {
		return fmt.Errorf("artifact assessment expires_at must be after assessed_at")
	}
	if value.Status == LifecycleSuperseded {
		if err := validateSHA256(value.Successor, "artifact assessment successor"); err != nil {
			return err
		}
		if value.Successor == value.Subject.Digest {
			return fmt.Errorf("artifact assessment successor must differ from subject digest")
		}
	} else if !value.Successor.IsZero() {
		return fmt.Errorf("artifact assessment successor is valid only for superseded status")
	}
	seen := map[string]struct{}{value.Subject.Digest.String(): {}}
	for _, supporting := range value.Supporting {
		key := supporting.Digest.String()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("artifact assessment contains duplicate digest %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// EffectiveStatus returns stale when an otherwise active assessment has
// expired at the caller-supplied time. The zero time leaves explicit status
// unchanged; callers must supply their own clock value when expiry matters.
func EffectiveStatus(value Assessment, at time.Time) LifecycleStatus {
	value = NormalizeAssessment(value)
	if value.Status == LifecycleActive && !at.IsZero() && !value.ExpiresAt.IsZero() && !at.Before(value.ExpiresAt) {
		return LifecycleStale
	}
	return value.Status
}

// CanonicalAssessmentJSON returns deterministic normalized JSON.
func CanonicalAssessmentJSON(value Assessment) ([]byte, error) {
	if err := ValidateAssessment(value); err != nil {
		return nil, err
	}
	value = NormalizeAssessment(value)
	return json.Marshal(value)
}

// DigestAssessment returns the SHA-256 digest of canonical assessment JSON.
func DigestAssessment(value Assessment) (digest.Record, error) {
	data, err := CanonicalAssessmentJSON(value)
	if err != nil {
		return digest.Record{}, err
	}
	return digest.SHA256Bytes(data), nil
}

func validateSHA256(value digest.Record, label string) error {
	if value.Algorithm != digest.AlgorithmSHA256 {
		return fmt.Errorf("%s algorithm must be %q", label, digest.AlgorithmSHA256)
	}
	if len(value.Value) != 64 {
		return fmt.Errorf("%s value must be 64 lowercase hexadecimal characters", label)
	}
	decoded, err := hex.DecodeString(value.Value)
	if err != nil || len(decoded) != 32 || value.Value != strings.ToLower(value.Value) {
		return fmt.Errorf("%s value must be 64 lowercase hexadecimal characters", label)
	}
	return nil
}

func normalizeMediaType(value string) string {
	value = strings.TrimSpace(value)
	parsed, params, err := mime.ParseMediaType(value)
	if err != nil {
		return value
	}
	return mime.FormatMediaType(strings.ToLower(parsed), params)
}

func normalizeAnnotations(value map[string]string) map[string]string {
	if len(value) == 0 {
		return nil
	}
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	out := make(map[string]string, len(value))
	for _, key := range keys {
		item := value[key]
		out[strings.TrimSpace(key)] = strings.TrimSpace(item)
	}
	return out
}

func validateAnnotationKeys(value map[string]string) error {
	seen := make(map[string]string, len(value))
	for key := range value {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			return fmt.Errorf("artifact descriptor annotation keys must be non-empty")
		}
		if prior, ok := seen[normalized]; ok && prior != key {
			return fmt.Errorf("artifact descriptor annotation keys %q and %q normalize to the same key", prior, key)
		}
		seen[normalized] = key
	}
	return nil
}

func normalizeTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC().Round(0)
}

func compareDescriptor(a, b Descriptor) int {
	if result := strings.Compare(a.Digest.String(), b.Digest.String()); result != 0 {
		return result
	}
	if result := strings.Compare(a.MediaType, b.MediaType); result != 0 {
		return result
	}
	if a.SizeBytes < b.SizeBytes {
		return -1
	}
	if a.SizeBytes > b.SizeBytes {
		return 1
	}
	return 0
}
