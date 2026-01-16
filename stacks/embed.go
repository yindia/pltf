package stacks

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

//go:embed *.yaml
var embeddedStacks embed.FS

var (
	once     sync.Once
	rootPath string
	rootErr  error
)

// Materialize copies embedded stacks to a temp directory and returns the root.
func Materialize() (string, error) {
	once.Do(func() {
		tmp, err := os.MkdirTemp("", "pltf-stacks-*")
		if err != nil {
			rootErr = fmt.Errorf("failed to create temp stacks dir: %w", err)
			return
		}
		if err := copyEmbedded(tmp); err != nil {
			rootErr = err
			return
		}
		rootPath = tmp
	})
	return rootPath, rootErr
}

// EmbeddedStackNames returns embedded stack names without extensions.
func EmbeddedStackNames() ([]string, error) {
	entries, err := fs.ReadDir(embeddedStacks, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded stacks: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			names = append(names, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// ResolveEmbeddedStack returns a local path to an embedded stack file by name.
func ResolveEmbeddedStack(name string) (string, error) {
	candidates := []string{name}
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		candidates = []string{name + ".yaml", name + ".yml"}
	}
	for _, candidate := range candidates {
		if _, err := embeddedStacks.Open(candidate); err == nil {
			root, err := Materialize()
			if err != nil {
				return "", err
			}
			return filepath.Join(root, candidate), nil
		}
	}
	return "", fmt.Errorf("embedded stack %q not found", name)
}

func copyEmbedded(dest string) error {
	return fs.WalkDir(embeddedStacks, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		target := filepath.Join(dest, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := embeddedStacks.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, f); err != nil {
			return err
		}
		return nil
	})
}
