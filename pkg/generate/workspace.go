package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"pltf/pkg/config"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// ExportEnvironmentWorkspace renders a single Terraform root with per-env tfvars files.
func ExportEnvironmentWorkspace(envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, outDir, specDir string, cliVars map[string]string) error {
	envKey := firstEnvKey(envCfg.Environments)
	if envKey == "" {
		return fmt.Errorf("no environments defined for workspace export")
	}
	if err := ensureWorkspaceBackendConsistent(envCfg, nil); err != nil {
		return err
	}
	g, err := NewGenerator(envCfg, nil, embeddedRoot, customRoot, envKey, outDir, specDir, nil)
	if err != nil {
		return err
	}
	workspacePrefix := fmt.Sprintf("env/%s", envCfg.Metadata.Name)
	if err := g.GenerateWorkspace(workspacePrefix); err != nil {
		return err
	}

	if err := writeVariablesTF(outDir, g.mergedVars, g.getSecretNames()); err != nil {
		return err
	}

	envKeys := sortedEnvKeys(envCfg.Environments)
	for _, key := range envKeys {
		values := buildWorkspaceVars(envCfg, nil, key, cliVars)
		if err := writeTFVars(filepath.Join(outDir, fmt.Sprintf("%s.tfvars", key)), values); err != nil {
			return err
		}
	}

	return nil
}

// ExportEnvironmentWorkspaceForEnv renders a single Terraform root and one tfvars file for envKey.
func ExportEnvironmentWorkspaceForEnv(envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, outDir, specDir, envKey string, cliVars map[string]string) error {
	if envKey == "" {
		return fmt.Errorf("environment key is required for workspace export")
	}
	if _, ok := envCfg.Environments[envKey]; !ok {
		return fmt.Errorf("environment %q not found for workspace export", envKey)
	}
	if err := ensureWorkspaceBackendConsistent(envCfg, nil); err != nil {
		return err
	}
	g, err := NewGenerator(envCfg, nil, embeddedRoot, customRoot, envKey, outDir, specDir, nil)
	if err != nil {
		return err
	}
	workspacePrefix := fmt.Sprintf("env/%s", envCfg.Metadata.Name)
	if err := g.GenerateWorkspace(workspacePrefix); err != nil {
		return err
	}
	if err := writeVariablesTF(outDir, g.mergedVars, g.getSecretNames()); err != nil {
		return err
	}
	values := buildWorkspaceVars(envCfg, nil, envKey, cliVars)
	if err := writeTFVars(filepath.Join(outDir, fmt.Sprintf("%s.tfvars", envKey)), values); err != nil {
		return err
	}
	return nil
}

// ExportServiceWorkspace renders a single Terraform root with per-env tfvars files.
func ExportServiceWorkspace(svcCfg *config.ServiceConfig, envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, outDir, specDir string, cliVars map[string]string) error {
	envKey := firstEnvKey(svcCfg.Metadata.EnvRef)
	if envKey == "" {
		return fmt.Errorf("no envRef entries defined for workspace export")
	}
	if err := ensureWorkspaceBackendConsistent(envCfg, svcCfg); err != nil {
		return err
	}
	g, err := NewGenerator(envCfg, svcCfg, embeddedRoot, customRoot, envKey, outDir, specDir, nil)
	if err != nil {
		return err
	}
	workspacePrefix := fmt.Sprintf("service/%s", svcCfg.Metadata.Name)
	if err := g.GenerateWorkspace(workspacePrefix); err != nil {
		return err
	}

	if err := writeVariablesTF(outDir, g.mergedVars, g.getSecretNames()); err != nil {
		return err
	}

	envKeys := sortedServiceEnvKeys(svcCfg.Metadata.EnvRef)
	for _, key := range envKeys {
		values := buildWorkspaceVars(envCfg, svcCfg, key, cliVars)
		if err := writeTFVars(filepath.Join(outDir, fmt.Sprintf("%s.tfvars", key)), values); err != nil {
			return err
		}
	}

	return nil
}

// ExportServiceWorkspaceForEnv renders a single Terraform root and one tfvars file for envKey.
func ExportServiceWorkspaceForEnv(svcCfg *config.ServiceConfig, envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, outDir, specDir, envKey string, cliVars map[string]string) error {
	if envKey == "" {
		return fmt.Errorf("environment key is required for workspace export")
	}
	if _, ok := svcCfg.Metadata.EnvRef[envKey]; !ok {
		return fmt.Errorf("service envRef.%q not found for workspace export", envKey)
	}
	if err := ensureWorkspaceBackendConsistent(envCfg, svcCfg); err != nil {
		return err
	}
	g, err := NewGenerator(envCfg, svcCfg, embeddedRoot, customRoot, envKey, outDir, specDir, nil)
	if err != nil {
		return err
	}
	workspacePrefix := fmt.Sprintf("service/%s", svcCfg.Metadata.Name)
	if err := g.GenerateWorkspace(workspacePrefix); err != nil {
		return err
	}
	if err := writeVariablesTF(outDir, g.mergedVars, g.getSecretNames()); err != nil {
		return err
	}
	values := buildWorkspaceVars(envCfg, svcCfg, envKey, cliVars)
	if err := writeTFVars(filepath.Join(outDir, fmt.Sprintf("%s.tfvars", envKey)), values); err != nil {
		return err
	}
	return nil
}

func writeVariablesTF(outDir string, vars map[string]interface{}, secretNames map[string]bool) error {
	file := hclwrite.NewEmptyFile()
	body := file.Body()
	keys := sortedKeysInterfaceMap(vars)
	for _, k := range keys {
		if secretNames[k] {
			continue
		}
		block := body.AppendNewBlock("variable", []string{k})
		b := block.Body()
		b.SetAttributeRaw("type", hclwrite.TokensForIdentifier("any"))
		body.AppendNewline()
	}
	return os.WriteFile(filepath.Join(outDir, "variables.tf"), file.Bytes(), 0o644)
}

func writeTFVars(path string, values map[string]interface{}) error {
	file := hclwrite.NewEmptyFile()
	body := file.Body()
	keys := sortedKeysInterfaceMap(values)
	for _, k := range keys {
		ctyVal, err := toCtyValue(values[k])
		if err != nil {
			return fmt.Errorf("tfvars %s: cannot convert %q: %w", path, k, err)
		}
		body.SetAttributeValue(k, ctyVal)
	}
	return os.WriteFile(path, file.Bytes(), 0o644)
}

func buildWorkspaceVars(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig, envKey string, cliVars map[string]string) map[string]interface{} {
	entry := envCfg.Environments[envKey]
	labels := envCfg.Metadata.Labels
	if svcCfg != nil {
		labels = mergeLabels(labels, svcCfg.Metadata.Labels)
	}
	out := map[string]interface{}{
		"account_id":  entry.Account,
		"region":      entry.Region,
		"environment": envKey,
		"global_tags": labels,
	}
	for k, v := range entry.Variables {
		out[k] = parseVarValue(v)
	}
	if svcCfg != nil {
		for k, v := range svcCfg.Variables {
			out[k] = parseVarValue(v)
		}
	}
	for k, v := range cliVars {
		out[k] = parseVarValue(v)
	}
	return out
}

func firstEnvKey[T any](m map[string]T) string {
	if len(m) == 0 {
		return ""
	}
	keys := sortedMapKeys(m)
	return keys[0]
}

func sortedEnvKeys(m map[string]config.EnvironmentEntry) []string {
	return sortedMapKeys(m)
}

func sortedServiceEnvKeys(m map[string]config.ServiceEnvRefEntry) []string {
	return sortedMapKeys(m)
}

func sortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ensureWorkspaceBackendConsistent(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}
	var envKeys []string
	if svcCfg == nil {
		envKeys = sortedEnvKeys(envCfg.Environments)
	} else {
		envKeys = sortedServiceEnvKeys(svcCfg.Metadata.EnvRef)
	}
	if len(envKeys) == 0 {
		return fmt.Errorf("no environments defined for workspace export")
	}
	baseEnv := envCfg.Environments[envKeys[0]]
	baseCfg, err := ResolveBackendConfig(envCfg.Metadata.Provider, envCfg, baseEnv)
	if err != nil {
		return err
	}
	for _, key := range envKeys[1:] {
		entry := envCfg.Environments[key]
		cfg, err := ResolveBackendConfig(envCfg.Metadata.Provider, envCfg, entry)
		if err != nil {
			return err
		}
		if cfg.BackendType != baseCfg.BackendType ||
			cfg.Bucket != baseCfg.Bucket ||
			cfg.Region != baseCfg.Region ||
			cfg.Container != baseCfg.Container ||
			cfg.ResourceGroup != baseCfg.ResourceGroup ||
			cfg.Profile != baseCfg.Profile {
			return fmt.Errorf("workspace export requires consistent backend config across environments (backend/region/bucket must match)")
		}
	}
	return nil
}
