// Package digest provides stable digest records and helpers.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// AlgorithmSHA256 is the canonical name used in Evidence digest records.
	AlgorithmSHA256 = "sha256"
)

// Record describes a digest with an explicit algorithm and encoded value.
type Record struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// String returns the conventional algorithm:value form.
func (r Record) String() string {
	algorithm := strings.TrimSpace(r.Algorithm)
	value := strings.TrimSpace(r.Value)
	if algorithm == "" {
		return value
	}
	if value == "" {
		return algorithm + ":"
	}
	return algorithm + ":" + value
}

// IsZero reports whether the record carries no digest data.
func (r Record) IsZero() bool {
	return strings.TrimSpace(r.Algorithm) == "" && strings.TrimSpace(r.Value) == ""
}

// SHA256Bytes returns a SHA-256 digest record for data.
func SHA256Bytes(data []byte) Record {
	sum := sha256.Sum256(data)
	return Record{Algorithm: AlgorithmSHA256, Value: hex.EncodeToString(sum[:])}
}

// SHA256String returns the conventional sha256:<hex> digest for data.
func SHA256String(data []byte) string {
	return SHA256Bytes(data).String()
}

// SHA256Reader consumes r and returns its SHA-256 digest.
func SHA256Reader(r io.Reader) (Record, error) {
	if r == nil {
		return Record{}, fmt.Errorf("digest reader is nil")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return Record{}, err
	}
	return Record{Algorithm: AlgorithmSHA256, Value: hex.EncodeToString(hash.Sum(nil))}, nil
}

// SHA256File returns the SHA-256 digest and byte size for a file.
func SHA256File(path string) (Record, int64, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Record{}, 0, fmt.Errorf("digest file path is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return Record{}, 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Record{}, 0, err
	}
	record, err := SHA256Reader(file)
	if err != nil {
		return Record{}, 0, err
	}
	return record, info.Size(), nil
}
