package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// ResolveSpecPath resolves a local or git-backed spec reference into a local file path.
// Git refs use: https://host/org/repo.git//path/to/spec.yaml?ref=main
func ResolveSpecPath(ref, baseDir string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("spec reference is empty")
	}
	if !isGitRef(ref) {
		if filepath.IsAbs(ref) {
			return filepath.Clean(ref), nil
		}
		if baseDir == "" {
			return filepath.Clean(ref), nil
		}
		return filepath.Clean(filepath.Join(baseDir, ref)), nil
	}

	repoURL, filePath, gitRef, err := parseGitRef(ref)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", fmt.Errorf("git ref %q is missing a file path (use //path/to/spec.yaml)", ref)
	}

	cacheDir := filepath.Join(baseDir, ".pltf", "cache", "git")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create git cache dir %s: %w", cacheDir, err)
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

	cleanPath := filepath.Clean(filePath)
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("git ref path %q escapes repository", filePath)
	}
	localPath := filepath.Join(repoDir, cleanPath)
	rel, err := filepath.Rel(repoDir, localPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("git ref path %q escapes repository", filePath)
	}
	return localPath, nil
}

// ResolveGitRef resolves a git ref to a local path. If the ref omits the //path,
// the repo root (".") is used.
func ResolveGitRef(ref, baseDir string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("git ref is empty")
	}
	if !isGitRef(ref) {
		return "", fmt.Errorf("not a git ref: %s", ref)
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid git ref %q: %w", ref, err)
	}
	if !strings.Contains(u.Path, "//") {
		ref = strings.TrimSuffix(ref, "/") + "//."
	}
	return ResolveSpecPath(ref, baseDir)
}

func isGitRef(ref string) bool {
	normalized, err := normalizeGitRef(ref)
	if err != nil {
		return false
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ssh" {
		return false
	}
	return u.Host != ""
}

func parseGitRef(ref string) (repoURL, filePath, gitRef string, err error) {
	normalized, err := normalizeGitRef(ref)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid git ref %q: %w", ref, err)
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid git ref %q: %w", ref, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ssh" {
		return "", "", "", fmt.Errorf("unsupported git scheme %q in %s", u.Scheme, ref)
	}
	pathParts := strings.SplitN(u.Path, "//", 2)
	if len(pathParts) != 2 {
		return "", "", "", fmt.Errorf("git ref %q must include //path/to/spec.yaml", ref)
	}
	repoURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, pathParts[0])
	filePath = strings.TrimPrefix(pathParts[1], "/")
	gitRef = u.Query().Get("ref")
	return repoURL, filePath, gitRef, nil
}

func hashRepoKey(repoURL, gitRef string) string {
	sum := sha256.Sum256([]byte(repoURL + "|" + gitRef))
	return hex.EncodeToString(sum[:])
}

func gitClone(repoURL, dir string) error {
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return fmt.Errorf("git clone failed for %s: %w", repoURL, err)
	}
	return nil
}

func gitCheckout(dir, gitRef string) error {
	if strings.TrimSpace(gitRef) == "" {
		gitRef = "HEAD"
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("open git repo %s: %w", dir, err)
	}

	if shouldRefreshGitRef(gitRef) && shouldFetchRepo(dir) {
		if err := repo.Fetch(&git.FetchOptions{Tags: git.AllTags}); err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("git fetch failed in %s: %w", dir, err)
		}
		_ = recordRepoFetch(dir)
	}

	if err := checkoutRevision(repo, gitRef); err == nil {
		return nil
	}
	if err := repo.Fetch(&git.FetchOptions{Tags: git.AllTags}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("git fetch failed in %s: %w", dir, err)
	}
	_ = recordRepoFetch(dir)
	return checkoutRevision(repo, gitRef)
}

func checkoutRevision(repo *git.Repository, ref string) error {
	rev, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Checkout(&git.CheckoutOptions{
		Hash:  *rev,
		Force: true,
	})
}

func shouldRefreshGitRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" || ref == "HEAD" {
		return true
	}
	if len(ref) != 40 {
		return true
	}
	for _, r := range ref {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return true
		}
	}
	return false
}

const gitFetchTTL = 10 * time.Minute

func shouldFetchRepo(dir string) bool {
	path := filepath.Join(dir, ".pltf_fetch")
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > gitFetchTTL
}

func recordRepoFetch(dir string) error {
	path := filepath.Join(dir, ".pltf_fetch")
	return os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0o644)
}
