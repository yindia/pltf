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

func detectProviderExplicit(data []byte) (ProviderExplicit, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return ProviderExplicit{}, err
	}
	return providerExplicitFromNode(&node), nil
}

func providerExplicitFromNode(node *yaml.Node) ProviderExplicit {
	if node == nil {
		return ProviderExplicit{}
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return ProviderExplicit{}
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		if key.Value == "providers" && val.Kind == yaml.MappingNode {
			return providerExplicitFromProviders(val)
		}
	}
	return ProviderExplicit{}
}

func providerExplicitFromProviders(node *yaml.Node) ProviderExplicit {
	var out ProviderExplicit
	if node == nil || node.Kind != yaml.MappingNode {
		return out
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		switch key.Value {
		case "helm":
			out.HelmSet = true
		case "kubernetes":
			out.KubernetesSet = true
		case "kustomize":
			out.KustomizeSet = true
		}
	}
	return out
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

	if err := applyStacksToEnvironment(path, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("environment validation failed for %s: %w", path, err)
	}

	applyEnvironmentGlobals(&cfg)
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

	providerExplicit, err := detectProviderExplicit(svcData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse providers in %s: %w", servicePath, err)
	}

	var svc ServiceConfig
	if err := decodeYAMLStrict(svcData, &svc, servicePath); err != nil {
		return nil, nil, err
	}
	svc.ProvidersExplicit = providerExplicit
	if svc.Kind != "Service" {
		return nil, nil, fmt.Errorf("file %s is kind %q, expected 'Service'", servicePath, svc.Kind)
	}

	if svc.Metadata.Ref == "" {
		return nil, nil, fmt.Errorf("service %s metadata.ref is empty (no environment reference)", svc.Metadata.Name)
	}

	// Resolve env path relative to service file
	envPath, err := ResolveSpecPath(svc.Metadata.Ref, filepath.Dir(servicePath))
	if err != nil {
		return nil, nil, err
	}

	env, err := LoadEnvironmentConfig(envPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load environment for service %s: %w", svc.Metadata.Name, err)
	}

	if err := applyStacksToService(servicePath, &svc); err != nil {
		return nil, nil, err
	}

	// Now validate service WITH environment context
	if err := svc.Validate(env); err != nil {
		return nil, nil, fmt.Errorf("service validation failed for %s: %w", servicePath, err)
	}

	applyServiceGlobals(&svc)
	return &svc, env, nil
}

// LoadStack loads, parses, and validates a Stack YAML.
func LoadStack(path string) (*StackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack file %s: %w", path, err)
	}

	var cfg StackConfig
	if err := decodeYAMLStrict(data, &cfg, path); err != nil {
		return nil, err
	}

	if cfg.Kind != "Stack" {
		return nil, fmt.Errorf("file %s is kind %q, expected 'Stack'", path, cfg.Kind)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("stack validation failed for %s: %w", path, err)
	}

	return &cfg, nil
}

func applyStacksToEnvironment(specPath string, cfg *EnvironmentConfig) error {
	if len(cfg.Metadata.Stacks) == 0 {
		return nil
	}

	stacks, err := loadStacks(specPath, cfg.Metadata.Stacks)
	if err != nil {
		return err
	}

	if err := ensureNoModuleOverrides(cfg.Modules, stacks, specPath); err != nil {
		return err
	}

	labels, variables, modules, providers := mergeStacks(stacks)
	if err := ensureNoVariableOverrides(cfg.Variables, variables, specPath); err != nil {
		return err
	}
	cfg.Metadata.Labels = mergeLabelMaps(labels, cfg.Metadata.Labels)
	cfg.Modules = mergeModules(modules, cfg.Modules)
	cfg.Variables = mergeStringMap(variables, cfg.Variables)
	cfg.Providers = mergeProviders(providers, cfg.Providers)
	return nil
}

func applyStacksToService(specPath string, cfg *ServiceConfig) error {
	if len(cfg.Metadata.Stacks) == 0 {
		return nil
	}

	stacks, err := loadStacks(specPath, cfg.Metadata.Stacks)
	if err != nil {
		return err
	}

	if err := ensureNoModuleOverrides(cfg.Modules, stacks, specPath); err != nil {
		return err
	}

	labels, variables, modules, providers := mergeStacks(stacks)
	if err := ensureNoVariableOverrides(cfg.Variables, variables, specPath); err != nil {
		return err
	}
	cfg.Metadata.Labels = mergeLabelMaps(labels, cfg.Metadata.Labels)
	cfg.Modules = mergeModules(modules, cfg.Modules)
	cfg.Variables = mergeStringMap(variables, cfg.Variables)
	cfg.Providers = mergeProviders(providers, cfg.Providers)
	return nil
}

func loadStacks(specPath string, stackRefs []string) ([]*StackConfig, error) {
	baseDir := filepath.Dir(specPath)
	stacks := make([]*StackConfig, 0, len(stackRefs))
	for _, ref := range stackRefs {
		if ref == "" {
			return nil, fmt.Errorf("stack reference is empty in %s", specPath)
		}
		stackPath, err := ResolveStackRef(ref, baseDir)
		if err != nil {
			return nil, err
		}

		stackCfg, err := LoadStack(stackPath)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, stackCfg)
	}
	return stacks, nil
}

func mergeStacks(stacks []*StackConfig) (map[string]string, map[string]string, []Module, ProviderRequirements) {
	var (
		labels    map[string]string
		variables map[string]string
		modules   []Module
		providers ProviderRequirements
	)
	for _, stack := range stacks {
		labels = mergeLabelMaps(labels, stack.Metadata.Labels)
		variables = mergeStringMap(variables, stack.Variables)
		modules = mergeModules(modules, stack.Modules)
		providers = mergeProviders(providers, stack.Providers)
	}
	return labels, variables, modules, providers
}

func mergeLabelMaps(base, override map[string]string) map[string]string {
	return mergeStringMap(base, override)
}

func mergeStringMap(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

func mergeSecretMap(base, override map[string]SecretRef) map[string]SecretRef {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := map[string]SecretRef{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

func mergeModules(base, override []Module) []Module {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}

	overrideIDs := make(map[string]struct{}, len(override))
	for _, m := range override {
		overrideIDs[m.ID] = struct{}{}
	}

	merged := make([]Module, 0, len(base)+len(override))
	for _, m := range base {
		if _, exists := overrideIDs[m.ID]; !exists {
			merged = append(merged, m)
		}
	}
	merged = append(merged, override...)
	return merged
}

func mergeProviders(base, override ProviderRequirements) ProviderRequirements {
	return ProviderRequirements{
		Kubernetes: base.Kubernetes || override.Kubernetes,
		Helm:       base.Helm || override.Helm,
		Kustomize:  base.Kustomize || override.Kustomize,
	}
}

func ensureNoModuleOverrides(specModules []Module, stacks []*StackConfig, specPath string) error {
	if len(specModules) == 0 || len(stacks) == 0 {
		return nil
	}

	stackIDs := make(map[string]struct{})
	for _, stack := range stacks {
		for _, m := range stack.Modules {
			stackIDs[m.ID] = struct{}{}
		}
	}

	for _, m := range specModules {
		if _, exists := stackIDs[m.ID]; exists {
			return fmt.Errorf("module %q in %s is already defined by a referenced stack; overrides are not allowed", m.ID, specPath)
		}
	}
	return nil
}

func ensureNoVariableOverrides(specVars, stackVars map[string]string, specPath string) error {
	if len(specVars) == 0 || len(stackVars) == 0 {
		return nil
	}
	for name := range specVars {
		if _, exists := stackVars[name]; exists {
			return fmt.Errorf(
				"variable %q in %s is already defined by a referenced stack; overrides are not allowed",
				name,
				specPath,
			)
		}
	}
	return nil
}

func applyEnvironmentGlobals(cfg *EnvironmentConfig) {
	if len(cfg.Variables) == 0 && len(cfg.Secrets) == 0 {
		return
	}
	for envName, envEntry := range cfg.Environments {
		if len(cfg.Variables) > 0 {
			envEntry.Variables = mergeStringMap(cfg.Variables, envEntry.Variables)
		}
		if len(cfg.Secrets) > 0 {
			envEntry.Secrets = mergeSecretMap(cfg.Secrets, envEntry.Secrets)
		}
		cfg.Environments[envName] = envEntry
	}
}

func applyServiceGlobals(cfg *ServiceConfig) {
	if cfg != nil && !cfg.ProvidersExplicit.HelmSet && !cfg.Providers.Helm {
		cfg.Providers.Helm = true
	}
	if len(cfg.Variables) == 0 && len(cfg.Secrets) == 0 {
		return
	}
	for envName, envEntry := range cfg.Metadata.EnvRef {
		if len(cfg.Variables) > 0 {
			envEntry.Variables = mergeStringMap(cfg.Variables, envEntry.Variables)
		}
		if len(cfg.Secrets) > 0 {
			envEntry.Secrets = mergeSecretMap(cfg.Secrets, envEntry.Secrets)
		}
		cfg.Metadata.EnvRef[envName] = envEntry
	}
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
