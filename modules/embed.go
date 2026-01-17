package modules

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

//go:embed * */*
var embeddedModules embed.FS

var (
	once     sync.Once
	rootPath string
	rootErr  error
)

var providerPrefixes = []string{"aws_", "gcp_", "azure_", "azurerm_"}

// Materialize copies the embedded modules directory to a temp directory and returns
// the path to the modules root. It runs once per process.
func Materialize() (string, error) {
	once.Do(func() {
		tmp, err := os.MkdirTemp("", "pltf-modules-*")
		if err != nil {
			rootErr = fmt.Errorf("failed to create temp modules dir: %w", err)
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

// EmbeddedModuleNames returns the list of embedded module directory names.
func EmbeddedModuleNames() ([]string, error) {
	entries, err := fs.ReadDir(embeddedModules, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded modules: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// IsEmbeddedModule returns true if name matches an embedded module directory.
func IsEmbeddedModule(name string) (bool, error) {
	names, err := EmbeddedModuleNames()
	if err != nil {
		return false, err
	}
	for _, n := range names {
		if n == name {
			return true, nil
		}
	}
	return false, nil
}

// ValidateModuleName checks naming conventions for module types.
func ValidateModuleName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("module name is empty")
	}
	if trimmed != name {
		return fmt.Errorf("module name %q has leading or trailing whitespace", name)
	}
	if name != strings.ToLower(name) {
		return fmt.Errorf("module name %q must be lowercase", name)
	}
	if !strings.Contains(name, "_") {
		return fmt.Errorf("module name %q must include a provider prefix like aws_*", name)
	}
	if !hasProviderPrefix(name) {
		return fmt.Errorf("module name %q must start with a provider prefix (%s)", name, strings.Join(providerPrefixes, ", "))
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fmt.Errorf("module name %q contains invalid character %q", name, r)
	}
	return nil
}

// ValidateEmbeddedModuleName verifies naming conventions and availability in the embedded catalog.
func ValidateEmbeddedModuleName(name string) error {
	if err := ValidateModuleName(name); err != nil {
		return err
	}
	ok, err := IsEmbeddedModule(name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("module %q not found in embedded modules", name)
	}
	return nil
}

func hasProviderPrefix(name string) bool {
	for _, prefix := range providerPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func copyEmbedded(dest string) error {
	return fs.WalkDir(embeddedModules, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".go") {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			return nil
		}

		target := filepath.Join(dest, path)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		f, err := embeddedModules.Open(path)
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
