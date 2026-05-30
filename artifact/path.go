package artifact

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// CleanRelativePath returns a canonical slash-separated relative artifact path.
func CleanRelativePath(inputPath string) (string, error) {
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
		return "", fmt.Errorf("path must stay inside artifact root: %q", inputPath)
	}
	return clean, nil
}

// ValidateRoot rejects missing, non-directory, and symlink artifact roots.
func ValidateRoot(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("artifact root is required")
	}
	info, err := os.Lstat(filepath.Clean(root))
	if err != nil {
		return fmt.Errorf("artifact root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact root must not be a symlink: %s", root)
	}
	if !info.IsDir() {
		return fmt.Errorf("artifact root must be a directory: %s", root)
	}
	return nil
}

// ValidateRegularFiles rejects missing, symlinked, directory, and special-file
// artifact inputs.
func ValidateRegularFiles(root string, paths []string) error {
	if err := ValidateRoot(root); err != nil {
		return err
	}
	root = filepath.Clean(root)
	for _, inputPath := range paths {
		clean, err := CleanRelativePath(inputPath)
		if err != nil {
			return fmt.Errorf("artifact path %q is unsafe: %w", inputPath, err)
		}
		if err := validateRegularFile(root, clean); err != nil {
			return err
		}
	}
	return nil
}

func validateRegularFile(root, clean string) error {
	segments := strings.Split(clean, "/")
	current := root
	for i, segment := range segments {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("artifact input %s: %w", clean, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifact input must not be a symlink: %s", clean)
		}
		last := i == len(segments)-1
		if !last {
			if !info.IsDir() {
				return fmt.Errorf("artifact input parent must be a directory: %s", clean)
			}
			continue
		}
		if info.IsDir() {
			return fmt.Errorf("artifact input must be a regular file, not a directory: %s", clean)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("artifact input must be a regular file: %s", clean)
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
