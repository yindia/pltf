package config

import (
	"path/filepath"
	"runtime"
	"strings"
)

// IsGitSource reports whether the provided source string looks like a git URL.
func IsGitSource(src string) bool {
	src = strings.TrimSpace(src)
	if src == "" {
		return false
	}
	lower := strings.ToLower(src)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "ssh://") ||
		strings.HasPrefix(src, "git@")
}

// IsCustomRootSource returns true when the source explicitly references the custom modules root.
func IsCustomRootSource(src string) bool {
	return strings.EqualFold(strings.TrimSpace(src), "custom")
}

// IsLocalSource returns true when the source looks like a local path (relative or absolute).
func IsLocalSource(src string) bool {
	src = strings.TrimSpace(src)
	if src == "" {
		return false
	}
	if IsGitSource(src) || IsCustomRootSource(src) {
		return false
	}
	if filepath.IsAbs(src) {
		return true
	}
	if strings.HasPrefix(src, "./") || strings.HasPrefix(src, "../") {
		return true
	}
	if runtime.GOOS == "windows" {
		if strings.Contains(src, "\\") {
			return true
		}
	}
	return strings.Contains(src, "/")
}
