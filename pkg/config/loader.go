package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"path/filepath"

	"gopkg.in/yaml.v3"
)

func decodeYAMLStrict(data []byte, target interface{}, path string) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("failed to parse yaml %s: %w", path, err)
	}
	// Ensure there are no trailing documents; decode until EOF.
	for {
		var extra interface{}
		if err := dec.Decode(&extra); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to parse yaml %s: %w", path, err)
		}
		return fmt.Errorf("file %s contains multiple YAML documents; only one is supported", path)
	}
	return nil
}

// DetectKind reads only the "kind" field from a YAML file without enforcing known fields.
// It returns the raw kind value or an error if the file cannot be read/parsed.
func DetectKind(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	var header struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return "", fmt.Errorf("failed to parse yaml %s: %w", path, err)
	}
	return header.Kind, nil
}

// LoadEnvironmentConfig loads, parses, and validates an Environment YAML.
func LoadEnvironmentConfig(path string) (*EnvironmentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %s: %w", path, err)
	}

	var cfg EnvironmentConfig
	if err := decodeYAMLStrict(data, &cfg, path); err != nil {
		return nil, err
	}

	if cfg.Kind != "Environment" {
		return nil, fmt.Errorf("file %s is kind %q, expected 'Environment'", path, cfg.Kind)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("environment validation failed for %s: %w", path, err)
	}

	return &cfg, nil
}

// LoadService loads a Service config AND the referenced Environment
// (following metadata.ref, resolving relative to the service file).
// Validation errors from either file are surfaced with context.
func LoadService(servicePath string) (*ServiceConfig, *EnvironmentConfig, error) {
	svcData, err := os.ReadFile(servicePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read service file %s: %w", servicePath, err)
	}

	var svc ServiceConfig
	if err := decodeYAMLStrict(svcData, &svc, servicePath); err != nil {
		return nil, nil, err
	}
	if svc.Kind != "Service" {
		return nil, nil, fmt.Errorf("file %s is kind %q, expected 'Service'", servicePath, svc.Kind)
	}

	if svc.Metadata.Ref == "" {
		return nil, nil, fmt.Errorf("service %s metadata.ref is empty (no environment reference)", svc.Metadata.Name)
	}

	// Resolve env path relative to service file
	envPath := svc.Metadata.Ref
	if !filepath.IsAbs(envPath) {
		envPath = filepath.Join(filepath.Dir(servicePath), envPath)
	}

	env, err := LoadEnvironmentConfig(envPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load environment for service %s: %w", svc.Metadata.Name, err)
	}

	// Now validate service WITH environment context
	if err := svc.Validate(env); err != nil {
		return nil, nil, fmt.Errorf("service validation failed for %s: %w", servicePath, err)
	}

	return &svc, env, nil
}

// LoadModuleMetadata reads module.yaml from a module directory.
func LoadModuleMetadata(dir string) (*ModuleMetadata, error) {
	path := filepath.Join(dir, "module.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read module metadata %s: %w", path, err)
	}

	var m ModuleMetadata
	if err := decodeYAMLStrict(data, &m, path); err != nil {
		return nil, err
	}

	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("module metadata validation failed for %s: %w", path, err)
	}

	return &m, nil
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// ScanModulesRoot scans a directory of modules (each subdir with module.yaml)
// and returns a map[type]*ModuleMetadata.
func ScanModulesRoot(root string, modules []string) (map[string]*ModuleMetadata, error) {
	if root == "" {
		return nil, fmt.Errorf("modules root is empty")
	}
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("modules root %s is not accessible: %w", root, err)
	}

	required := make(map[string]struct{}, len(modules))
	for _, m := range modules {
		required[m] = struct{}{}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules root %s: %w", root, err)
	}

	result := make(map[string]*ModuleMetadata)

	for _, e := range entries {
		if !contains(modules, e.Name()) {
			continue
		}
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		m, err := LoadModuleMetadata(dir)
		if err != nil {
			// You can decide: skip or fail fast. For now, fail fast.
			return nil, err
		}

		if _, exists := result[m.Type]; exists {
			return nil, fmt.Errorf("duplicate module type %q (dirs: %s and another)", m.Type, dir)
		}
		result[m.Type] = m
	}

	for moduleType := range required {
		if _, ok := result[moduleType]; !ok {
			return nil, fmt.Errorf("module type %q not found under %s", moduleType, root)
		}
	}

	return result, nil
}
