package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// LoadModuleMetadataFromGitSource clones the repo referenced by source (HTTP, HTTPS, SSH),
// checks out the requested ref, and loads module.yaml from the target directory.
func LoadModuleMetadataFromGitSource(source, cacheDir string) (*ModuleMetadata, string, error) {
	moduleDir, err := resolveModuleGitSource(source, cacheDir)
	if err != nil {
		return nil, "", err
	}

	meta, err := LoadModuleMetadata(moduleDir)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(meta.Type) == "" {
		return nil, "", fmt.Errorf("module metadata missing type in %s", moduleDir)
	}
	return meta, moduleDir, nil
}

func resolveModuleGitSource(source, cacheDir string) (string, error) {
	repoURL, modulePath, gitRef, err := parseGitModuleRef(source)
	if err != nil {
		return "", err
	}

	if cacheDir == "" {
		cacheDir = filepath.Join(".pltf", "cache", "modules")
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create module git cache dir %s: %w", cacheDir, err)
	}

	repoKey := hashRepoKey(repoURL, gitRef)
	repoDir := filepath.Join(cacheDir, repoKey)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := gitClone(repoURL, repoDir); err != nil {
			return "", err
		}
	}

	if err := gitCheckout(repoDir, gitRef); err != nil {
		return "", err
	}

	cleanPath := filepath.Clean(modulePath)
	moduleDir := filepath.Join(repoDir, cleanPath)
	rel, err := filepath.Rel(repoDir, moduleDir)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("module path %q escapes repository", modulePath)
	}
	return moduleDir, nil
}

func parseGitModuleRef(ref string) (repoURL, modulePath, gitRef string, err error) {
	normalized, err := normalizeGitRef(ref)
	if err != nil {
		return "", "", "", err
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid git module ref %q: %w", ref, err)
	}
	pathParts := strings.SplitN(u.Path, "//", 2)
	repoURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, pathParts[0])
	modulePath = "."
	if len(pathParts) == 2 {
		modulePath = strings.TrimPrefix(pathParts[1], "/")
		if modulePath == "" {
			modulePath = "."
		}
	}
	gitRef = u.Query().Get("ref")
	return repoURL, modulePath, gitRef, nil
}

func normalizeGitRef(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("git ref is empty")
	}
	if strings.HasPrefix(raw, "git@") {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid git ref %q", raw)
		}
		return "ssh://" + parts[0] + "/" + parts[1], nil
	}
	if strings.Contains(raw, "://") {
		return raw, nil
	}
	return "", fmt.Errorf("git ref %q missing scheme", raw)
}
