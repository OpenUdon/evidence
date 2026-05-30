// Package artifact provides deterministic artifact records and manifests.
package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/OpenUdon/evidence/digest"
)

const (
	// ManifestVersion is the default artifact manifest schema version.
	ManifestVersion = "evidence.artifact-manifest.v1"
)

// Record describes one artifact without embedding artifact content.
type Record struct {
	Path           string        `json:"path"`
	Kind           string        `json:"kind,omitempty"`
	MediaType      string        `json:"media_type,omitempty"`
	Classification string        `json:"classification,omitempty"`
	Required       bool          `json:"required,omitempty"`
	SizeBytes      int64         `json:"size_bytes"`
	Digest         digest.Record `json:"digest"`
}

// Manifest describes a deterministic set of artifacts.
type Manifest struct {
	Version   string   `json:"version"`
	Root      string   `json:"root,omitempty"`
	Artifacts []Record `json:"artifacts"`
}

// FileOptions supplies caller-owned metadata for a file artifact record.
type FileOptions struct {
	Path           string
	Kind           string
	MediaType      string
	Classification string
	Required       bool
}

// ManifestOptions configures manifest generation.
type ManifestOptions struct {
	Version string
	Root    string
	Files   []FileOptions
}

// FileRecord returns a record for one safe relative file under root.
func FileRecord(root string, opts FileOptions) (Record, error) {
	if err := ValidateRoot(root); err != nil {
		return Record{}, err
	}
	clean, err := CleanRelativePath(opts.Path)
	if err != nil {
		return Record{}, err
	}
	if err := ValidateRegularFiles(root, []string{clean}); err != nil {
		return Record{}, err
	}
	record, size, err := digest.SHA256File(filepath.Join(filepath.Clean(root), filepath.FromSlash(clean)))
	if err != nil {
		return Record{}, fmt.Errorf("digest artifact %s: %w", clean, err)
	}
	return Record{
		Path:           clean,
		Kind:           strings.TrimSpace(opts.Kind),
		MediaType:      strings.TrimSpace(opts.MediaType),
		Classification: strings.TrimSpace(opts.Classification),
		Required:       opts.Required,
		SizeBytes:      size,
		Digest:         record,
	}, nil
}

// NewManifest returns a deterministic manifest for caller-supplied files.
func NewManifest(opts ManifestOptions) (Manifest, error) {
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = ManifestVersion
	}
	root := strings.TrimSpace(opts.Root)
	if root == "" {
		root = "."
	}
	seen := map[string]FileOptions{}
	for _, file := range opts.Files {
		clean, err := CleanRelativePath(file.Path)
		if err != nil {
			return Manifest{}, err
		}
		file.Path = clean
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = file
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	manifest := Manifest{
		Version: version,
		Root:    strings.TrimSpace(opts.Root),
	}
	for _, path := range paths {
		record, err := FileRecord(root, seen[path])
		if err != nil {
			return Manifest{}, err
		}
		manifest.Artifacts = append(manifest.Artifacts, record)
	}
	return manifest, nil
}

// DirectoryManifest returns a deterministic manifest of every regular file
// below root. Symlinked files or directories are rejected.
func DirectoryManifest(root string, version ...string) (Manifest, error) {
	if err := ValidateRoot(root); err != nil {
		return Manifest{}, err
	}
	var files []FileOptions
	root = filepath.Clean(root)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		clean, err := CleanRelativePath(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifact input must not be a symlink: %s", clean)
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("artifact input must be a regular file: %s", clean)
		}
		files = append(files, FileOptions{Path: clean})
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}
	opts := ManifestOptions{Root: root, Files: files}
	if len(version) > 0 {
		opts.Version = version[0]
	}
	return NewManifest(opts)
}

// Digest returns a digest over the canonical JSON manifest.
func (m Manifest) Digest() (digest.Record, error) {
	clone := Manifest{
		Version:   strings.TrimSpace(m.Version),
		Root:      strings.TrimSpace(m.Root),
		Artifacts: append([]Record(nil), m.Artifacts...),
	}
	sort.SliceStable(clone.Artifacts, func(i, j int) bool {
		return clone.Artifacts[i].Path < clone.Artifacts[j].Path
	})
	data, err := json.Marshal(clone)
	if err != nil {
		return digest.Record{}, err
	}
	return digest.SHA256Bytes(data), nil
}
