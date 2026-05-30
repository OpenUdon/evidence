package digest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSHA256BytesRecord(t *testing.T) {
	got := SHA256Bytes([]byte("hello\n"))
	if got.Algorithm != AlgorithmSHA256 {
		t.Fatalf("algorithm = %q", got.Algorithm)
	}
	if got.Value != "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03" {
		t.Fatalf("value = %q", got.Value)
	}
	if got.String() != "sha256:"+got.Value {
		t.Fatalf("string = %q", got.String())
	}
}

func TestSHA256ReaderRejectsNil(t *testing.T) {
	if _, err := SHA256Reader(nil); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("expected nil reader error, got %v", err)
	}
}

func TestSHA256File(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "artifact.txt")
	if err := os.WriteFile(path, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, size, err := SHA256File(path)
	if err != nil {
		t.Fatal(err)
	}
	if size != 6 {
		t.Fatalf("size = %d", size)
	}
	if got.String() != "sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03" {
		t.Fatalf("digest = %q", got.String())
	}
}
