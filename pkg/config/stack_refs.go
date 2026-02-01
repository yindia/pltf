package config

import (
	"fmt"
	"strings"

	"pltf/stacks"
)

// ResolveStackRef resolves a stack reference. If the ref is a plain name,
// it is resolved from embedded stacks; otherwise it is treated as a path/git ref.
func ResolveStackRef(ref, baseDir string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("stack reference is empty")
	}
	if looksLikeStackName(ref) {
		return stacks.ResolveEmbeddedStack(ref)
	}
	return ResolveSpecPath(ref, baseDir)
}

func looksLikeStackName(ref string) bool {
	if isGitRef(ref) {
		return false
	}
	if strings.Contains(ref, "://") {
		return false
	}
	if strings.ContainsAny(ref, "/\\") {
		return false
	}
	if strings.HasSuffix(ref, ".yaml") || strings.HasSuffix(ref, ".yml") {
		return false
	}
	return true
}
