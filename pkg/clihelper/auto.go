package clihelper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pltf/modules"
	"pltf/pkg/config"
	"pltf/pkg/generate"
)

func AutoValidate(file, env, modulesRoot string) error {
	return AutoValidateWithOutput(os.Stdout, file, env, modulesRoot)
}

func AutoValidateWithOutput(out io.Writer, file, env, modulesRoot string) error {
	var (
		envCfg *config.EnvironmentConfig
		svcCfg *config.ServiceConfig
		// envName string
	)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		cfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		_, err = SelectEnvName(kind, env, cfg, nil)
		if err != nil {
			return err
		}
		envCfg = cfg
		fmt.Fprintf(out, "Environment %q is valid (provider=%s, org=%s)\n",
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
			envCfg.Metadata.Org,
		)

	case "Service":
		svc, envConfig, err := config.LoadService(file)
		if err != nil {
			return err
		}
		_, err = SelectEnvName(kind, env, envConfig, svc)
		if err != nil {
			return err
		}
		svcCfg = svc
		envCfg = envConfig
		fmt.Fprintf(out, "Service %q is valid and uses Environment %q (provider=%s)\n",
			svcCfg.Metadata.Name,
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
		)

	case "Stack":
		stackCfg, err := config.LoadStack(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Stack %q is valid\n", stackCfg.Metadata.Name)
		embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
		if err != nil {
			return err
		}
		if err := validateDuplicateOutputsForModules(stackCfg.Modules, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateClusterModulesForModules(stackCfg.Modules, embeddedRoot, customRoot); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}
	if err := validateCustomModules(envCfg, svcCfg, customRoot); err != nil {
		return err
	}
	if err := validateProviderSupport(envCfg, svcCfg); err != nil {
		return err
	}
	if err := validateDuplicateOutputs(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
		return err
	}
	if err := validateClusterModules(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
		return err
	}

	// Run lint suggestions alongside validation.
	return nil
}

func AutoValidateWithScan(out io.Writer, file, env, modules string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	embeddedRoot, customRoot, err := ResolveModuleRoots(modules)
	if err != nil {
		return err
	}

	outDir, err := os.MkdirTemp("", "pltf-validate-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(outDir)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, nil, customRoot); err != nil {
			return err
		}
		if err := validateClusterModules(envCfg, nil, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateProviderSupport(envCfg, nil); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Environment %q is valid (provider=%s, org=%s)\n",
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
			envCfg.Metadata.Org,
		)
		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, outDir, specDir, envName, nil); err != nil {
			return err
		}

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, svcCfg, customRoot); err != nil {
			return err
		}
		if err := validateClusterModules(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateProviderSupport(envCfg, svcCfg); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Service %q is valid and uses Environment %q (provider=%s)\n",
			svcCfg.Metadata.Name,
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
		)
		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, outDir, specDir, envName, nil); err != nil {
			return err
		}

	case "Stack":
		stackCfg, err := config.LoadStack(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Stack %q is valid (scan not supported for Stack specs)\n", stackCfg.Metadata.Name)
		if err := validateDuplicateOutputsForModules(stackCfg.Modules, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateClusterModulesForModules(stackCfg.Modules, embeddedRoot, customRoot); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}

	if _, err := RunTfsecScan(outDir); err != nil {
		return fmt.Errorf("tfsec scan failed: %w", err)
	}
	return nil
}

func AutoGenerate(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, nil, customRoot); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Environment workspace for %q (env=%s) into %s\n",
			envCfg.Metadata.Name, envName, out)
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, svcCfg, customRoot); err != nil {
			return err
		}
		if err := validateClusterModules(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateDuplicateOutputs(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Service workspace for %q (env=%s) into %s\n",
			svcCfg.Metadata.Name, envName, out)
		return nil

	case "Stack":
		return fmt.Errorf("stack specs cannot be generated directly; reference them from Environment or Service specs")

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}
}

func AutoGenerateAll(file, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, nil, customRoot); err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspace(envCfg, embeddedRoot, customRoot, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Environment workspace for %q into %s\n", envCfg.Metadata.Name, out)
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, svcCfg, customRoot); err != nil {
			return err
		}
		if err := validateClusterModules(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
			return err
		}
		if err := validateDuplicateOutputs(envCfg, svcCfg, embeddedRoot, customRoot); err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspace(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Service workspace for %q into %s\n", svcCfg.Metadata.Name, out)
		return nil

	case "Stack":
		return fmt.Errorf("stack specs cannot be generated directly; reference them from Environment or Service specs")

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}
}

// autoGenerateQuiet renders Terraform without printing status messages. Used by graph command to keep DOT output clean.
func AutoGenerateQuiet(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, nil, customRoot); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validateCustomModules(envCfg, svcCfg, customRoot); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}
}

func validateCustomModules(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig, customRoot string) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}

	var customModules []config.Module
	for _, m := range envCfg.Modules {
		if strings.EqualFold(m.Source, "custom") {
			customModules = append(customModules, m)
		}
	}
	if svcCfg != nil {
		for _, m := range svcCfg.Modules {
			if strings.EqualFold(m.Source, "custom") {
				customModules = append(customModules, m)
			}
		}
	}

	if len(customModules) == 0 {
		return nil
	}
	if strings.TrimSpace(customRoot) == "" {
		return fmt.Errorf("spec uses source=custom but no custom modules root provided (--modules or profile.modules_root)")
	}

	metas, err := config.ScanModuleMetas(customRoot)
	if err != nil {
		return err
	}

	for _, m := range customModules {
		if err := modules.ValidateModuleName(m.Type); err != nil {
			return fmt.Errorf("module %q type %q invalid: %w", m.ID, m.Type, err)
		}
		if _, ok := metas[m.Type]; !ok {
			return fmt.Errorf("module %q (type=%s) marked source=custom but not found under %s", m.ID, m.Type, customRoot)
		}
	}
	return nil
}

func validateDuplicateOutputs(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig, embeddedRoot, customRoot string) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}
	modules := append([]config.Module{}, envCfg.Modules...)
	if svcCfg != nil {
		modules = append(modules, svcCfg.Modules...)
	}
	return validateDuplicateOutputsForModules(modules, embeddedRoot, customRoot)
}

func validateDuplicateOutputsForModules(modules []config.Module, embeddedRoot, customRoot string) error {
	if len(modules) == 0 {
		return nil
	}
	embeddedMetas, err := config.ScanModuleMetas(embeddedRoot)
	if err != nil {
		return err
	}
	var customMetas map[string]*config.ModuleMetadata
	if strings.TrimSpace(customRoot) != "" {
		customMetas, err = config.ScanModuleMetas(customRoot)
		if err != nil {
			return err
		}
	}

	outputProviders := map[string][]string{}
	for _, m := range modules {
		meta, err := SelectModuleMeta(m, embeddedMetas, customMetas, embeddedRoot)
		if err != nil {
			return err
		}
		for _, out := range meta.Outputs {
			outputProviders[out.Name] = append(outputProviders[out.Name], m.ID)
		}
	}

	for name, providers := range outputProviders {
		if len(providers) > 1 {
			return fmt.Errorf("output %q is provided by multiple modules: %s; auto-wiring requires unique output names", name, strings.Join(providers, ", "))
		}
	}
	return nil
}

func validateProviderSupport(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
	if envCfg == nil {
		return nil
	}
	provider := strings.ToLower(strings.TrimSpace(envCfg.Metadata.Provider))
	requiresK8s := envCfg.Providers.Kubernetes || envCfg.Providers.Helm || envCfg.Providers.Kustomize
	if svcCfg != nil {
		requiresK8s = requiresK8s || svcCfg.Providers.Kubernetes || svcCfg.Providers.Helm || svcCfg.Providers.Kustomize
	}
	if requiresK8s && provider != "aws" && provider != "gcp" && provider != "google" {
		return fmt.Errorf("kubernetes/helm/kustomize providers are only supported for aws and gcp right now; support for azure is coming soon")
	}
	return nil
}

func validateClusterModules(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig, embeddedRoot, customRoot string) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}
	modules := append([]config.Module{}, envCfg.Modules...)
	if svcCfg != nil {
		modules = append(modules, svcCfg.Modules...)
	}
	return validateClusterModulesForModules(modules, embeddedRoot, customRoot)
}

func validateClusterModulesForModules(modules []config.Module, embeddedRoot, customRoot string) error {
	if len(modules) == 0 {
		return nil
	}
	embeddedMetas, err := config.ScanModuleMetas(embeddedRoot)
	if err != nil {
		return err
	}
	var customMetas map[string]*config.ModuleMetadata
	if strings.TrimSpace(customRoot) != "" {
		customMetas, err = config.ScanModuleMetas(customRoot)
		if err != nil {
			return err
		}
	}

	var clusterModules []string
	for _, m := range modules {
		meta, err := SelectModuleMeta(m, embeddedMetas, customMetas, embeddedRoot)
		if err != nil {
			return err
		}
		if meta.Cluster {
			clusterModules = append(clusterModules, m.ID)
		}
	}
	if len(clusterModules) > 1 {
		return fmt.Errorf("multiple modules are marked cluster providers: %s", strings.Join(clusterModules, ", "))
	}
	return nil
}
