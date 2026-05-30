package artifact

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCleanRelativePathRejectsUnsafePaths(t *testing.T) {
	for _, input := range []string{"", "../secret.txt", "a/../secret.txt", "/tmp/file", `a\b.txt`, "C:/tmp/file"} {
		if _, err := CleanRelativePath(input); err == nil {
			t.Fatalf("expected unsafe path error for %q", input)
		}
	}
	got, err := CleanRelativePath("a/./b.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a/b.txt" {
		t.Fatalf("clean path = %q", got)
	}
}

func TestFileRecordHashesRegularFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "expected", "quality.json"), "{}\n")

	record, err := FileRecord(root, FileOptions{
		Path:      "expected/quality.json",
		MediaType: "application/json",
		Required:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.Path != "expected/quality.json" || record.SizeBytes != 3 || !record.Required {
		t.Fatalf("record = %#v", record)
	}
	if record.Digest.String() != "sha256:ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356" {
		t.Fatalf("digest = %q", record.Digest.String())
	}
}

func TestFileRecordRejectsMissingFile(t *testing.T) {
	_, err := FileRecord(t.TempDir(), FileOptions{Path: "missing.txt"})
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestFileRecordRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "target.txt"), "secret\n")
	if err := os.Symlink(filepath.Join(root, "target.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}
	_, err := FileRecord(root, FileOptions{Path: "link.txt"})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink error, got %v", err)
	}
}

func TestNewManifestSortsAndDeduplicatesFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "b.txt"), "b\n")
	writeFile(t, filepath.Join(root, "a.txt"), "a\n")

	manifest, err := NewManifest(ManifestOptions{
		Root: root,
		Files: []FileOptions{
			{Path: "b.txt", Kind: "text"},
			{Path: "a.txt", Kind: "text"},
			{Path: "b.txt", Kind: "duplicate"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := artifactPaths(manifest); !slices.Equal(got, []string{"a.txt", "b.txt"}) {
		t.Fatalf("paths = %#v", got)
	}
	if manifest.Artifacts[1].Kind != "text" {
		t.Fatalf("duplicate did not preserve first metadata: %#v", manifest.Artifacts[1])
	}
}

func TestDirectoryManifestCollectsRegularFilesInOrder(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "z.txt"), "z\n")
	writeFile(t, filepath.Join(root, "nested", "a.txt"), "a\n")

	manifest, err := DirectoryManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := artifactPaths(manifest); !slices.Equal(got, []string{"nested/a.txt", "z.txt"}) {
		t.Fatalf("paths = %#v", got)
	}
	digestA, err := manifest.Digest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Artifacts[0], manifest.Artifacts[1] = manifest.Artifacts[1], manifest.Artifacts[0]
	digestB, err := manifest.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if digestA != digestB {
		t.Fatalf("manifest digest changed with artifact order: %s != %s", digestA.String(), digestB.String())
	}
}

func artifactPaths(manifest Manifest) []string {
	paths := make([]string, 0, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		paths = append(paths, artifact.Path)
	}
	return paths
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
