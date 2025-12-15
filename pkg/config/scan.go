package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ModuleRecord includes metadata and its source root.
type ModuleRecord struct {
	Meta *ModuleMetadata
	Root string
}

// ScanModuleMetas scans a single modules root (directories containing module.yaml) and returns type -> metadata.
func ScanModuleMetas(root string) (map[string]*ModuleMetadata, error) {
	records, err := ScanModuleRoots([]string{root}, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*ModuleMetadata, len(records))
	for k, v := range records {
		out[k] = v.Meta
	}
	return out, nil
}

// ScanModuleRoots scans multiple roots (in priority order) and returns a map of type -> ModuleRecord.
// If moduleTypes is non-nil, only those types are required; missing required types produce an error.
func ScanModuleRoots(roots []string, moduleTypes []string) (map[string]ModuleRecord, error) {
	if len(roots) == 0 {
		return nil, fmt.Errorf("modules roots are empty")
	}
	required := make(map[string]struct{})
	for _, m := range moduleTypes {
		required[m] = struct{}{}
	}

	result := map[string]ModuleRecord{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, fmt.Errorf("failed to read modules root %s: %w", root, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(root, e.Name())
			meta, err := LoadModuleMetadata(dir)
			if err != nil {
				continue
			}
			// Only set if not already found (earlier roots have precedence)
			if _, exists := result[meta.Type]; !exists {
				result[meta.Type] = ModuleRecord{Meta: meta, Root: root}
			}
		}
	}

	for req := range required {
		if _, ok := result[req]; !ok {
			return nil, fmt.Errorf("module type %q not found in roots %v", req, roots)
		}
	}

	return result, nil
}
