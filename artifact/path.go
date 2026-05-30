package artifact

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// Options customizes the nouns used in artifact validation error messages so a
// downstream product can surface domain-specific wording (for example "package
// root" or "required handoff input") while sharing one validation
// implementation. The zero value reproduces the default "artifact" wording.
type Options struct {
	// RootLabel names the root directory; default "artifact root".
	RootLabel string
	// PathLabel names a relative path; default "artifact path".
	PathLabel string
	// InputLabel names a validated file; default "artifact input".
	InputLabel string
}

func resolveLabels(opts ...Options) (root, pathLabel, input string) {
	root, pathLabel, input = "artifact root", "artifact path", "artifact input"
	for _, o := range opts {
		if s := strings.TrimSpace(o.RootLabel); s != "" {
			root = s
		}
		if s := strings.TrimSpace(o.PathLabel); s != "" {
			pathLabel = s
		}
		if s := strings.TrimSpace(o.InputLabel); s != "" {
			input = s
		}
	}
	return root, pathLabel, input
}

// CleanRelativePath returns a canonical slash-separated relative artifact path.
func CleanRelativePath(inputPath string, opts ...Options) (string, error) {
	rootLabel, _, _ := resolveLabels(opts...)
	for _, r := range inputPath {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("path must not contain control characters: %q", inputPath)
		}
	}
	inputPath = strings.TrimSpace(inputPath)
	if inputPath == "" {
		return "", fmt.Errorf("path must be non-empty")
	}
	if strings.Contains(inputPath, `\`) {
		return "", fmt.Errorf("path must use slash separators: %q", inputPath)
	}
	if path.IsAbs(inputPath) || strings.HasPrefix(inputPath, "/") {
		return "", fmt.Errorf("path must be relative: %q", inputPath)
	}
	if hasWindowsVolumeName(inputPath) {
		return "", fmt.Errorf("path must not include a volume prefix: %q", inputPath)
	}
	for _, segment := range strings.Split(inputPath, "/") {
		if segment == ".." {
			return "", fmt.Errorf("path must not contain '..' segments: %q", inputPath)
		}
	}
	clean := path.Clean(inputPath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path must stay inside %s: %q", rootLabel, inputPath)
	}
	return clean, nil
}

// ValidateRoot rejects missing, non-directory, and symlink artifact roots.
func ValidateRoot(root string, opts ...Options) error {
	rootLabel, _, _ := resolveLabels(opts...)
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("%s is required", rootLabel)
	}
	info, err := os.Lstat(filepath.Clean(root))
	if err != nil {
		return fmt.Errorf("%s: %w", rootLabel, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s must not be a symlink: %s", rootLabel, root)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory: %s", rootLabel, root)
	}
	return nil
}

// ValidateRegularFiles rejects missing, symlinked, directory, and special-file
// artifact inputs.
func ValidateRegularFiles(root string, paths []string, opts ...Options) error {
	if err := ValidateRoot(root, opts...); err != nil {
		return err
	}
	_, pathLabel, _ := resolveLabels(opts...)
	root = filepath.Clean(root)
	for _, inputPath := range paths {
		clean, err := CleanRelativePath(inputPath, opts...)
		if err != nil {
			return fmt.Errorf("%s %q is unsafe: %w", pathLabel, inputPath, err)
		}
		if err := validateRegularFile(root, clean, opts...); err != nil {
			return err
		}
	}
	return nil
}

func validateRegularFile(root, clean string, opts ...Options) error {
	_, _, input := resolveLabels(opts...)
	segments := strings.Split(clean, "/")
	current := root
	for i, segment := range segments {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("%s %s: %w", input, clean, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s must not be a symlink: %s", input, clean)
		}
		last := i == len(segments)-1
		if !last {
			if !info.IsDir() {
				return fmt.Errorf("%s parent must be a directory: %s", input, clean)
			}
			continue
		}
		if info.IsDir() {
			return fmt.Errorf("%s must be a regular file, not a directory: %s", input, clean)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s must be a regular file: %s", input, clean)
		}
	}
	return nil
}

func hasWindowsVolumeName(inputPath string) bool {
	if len(inputPath) < 2 || inputPath[1] != ':' {
		return false
	}
	first := inputPath[0]
	return ('A' <= first && first <= 'Z') || ('a' <= first && first <= 'z') || runtime.GOOS == "windows" && filepath.VolumeName(inputPath) != ""
}
