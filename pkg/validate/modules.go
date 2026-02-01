package validate

import (
	"fmt"
	"strings"

	"pltf/modules"
	"pltf/pkg/config"
)

type ModuleValidator struct {
	embeddedRoot   string
	customRoot     string
	embeddedMetas  map[string]*config.ModuleMetadata
	customMetas    map[string]*config.ModuleMetadata
	embeddedLoaded bool
	customLoaded   bool
	embeddedErr    error
	customErr      error
}

func NewModuleValidator(embeddedRoot, customRoot string) *ModuleValidator {
	return &ModuleValidator{
		embeddedRoot: strings.TrimSpace(embeddedRoot),
		customRoot:   strings.TrimSpace(customRoot),
	}
}

func (v *ModuleValidator) ValidateCustomModules(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}

	var customModules []config.Module
	for _, m := range envCfg.Modules {
		if config.IsCustomRootSource(m.Source) {
			customModules = append(customModules, m)
		}
	}
	if svcCfg != nil {
		for _, m := range svcCfg.Modules {
			if config.IsCustomRootSource(m.Source) {
				customModules = append(customModules, m)
			}
		}
	}

	if len(customModules) == 0 {
		return nil
	}
	if strings.TrimSpace(v.customRoot) == "" {
		return fmt.Errorf("spec uses source=custom but no custom modules root provided (--modules or profile.modules_root)")
	}

	customMetas, err := v.loadCustom()
	if err != nil {
		return err
	}

	for _, m := range customModules {
		if err := modules.ValidateModuleName(m.Type); err != nil {
			return fmt.Errorf("module %q type %q invalid: %w", m.ID, m.Type, err)
		}
		if _, ok := customMetas[m.Type]; !ok {
			return fmt.Errorf("module %q (type=%s) marked source=custom but not found under %s", m.ID, m.Type, v.customRoot)
		}
	}
	return nil
}

func (v *ModuleValidator) ValidateDuplicateOutputs(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}
	modules := append([]config.Module{}, envCfg.Modules...)
	if svcCfg != nil {
		modules = append(modules, svcCfg.Modules...)
	}
	return v.ValidateDuplicateOutputsForModules(modules)
}

func (v *ModuleValidator) ValidateDuplicateOutputsForModules(modules []config.Module) error {
	if len(modules) == 0 {
		return nil
	}
	embeddedMetas, err := v.loadEmbedded()
	if err != nil {
		return err
	}
	customMetas, err := v.loadCustom()
	if err != nil {
		return err
	}

	outputProviders := map[string][]string{}
	for _, m := range modules {
		meta, err := v.selectModuleMeta(m, embeddedMetas, customMetas)
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

func (v *ModuleValidator) ValidateClusterModules(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
	if envCfg == nil {
		return fmt.Errorf("environment config is required")
	}
	modules := append([]config.Module{}, envCfg.Modules...)
	if svcCfg != nil {
		modules = append(modules, svcCfg.Modules...)
	}
	return v.ValidateClusterModulesForModules(modules)
}

func (v *ModuleValidator) ValidateClusterModulesForModules(modules []config.Module) error {
	if len(modules) == 0 {
		return nil
	}
	embeddedMetas, err := v.loadEmbedded()
	if err != nil {
		return err
	}
	customMetas, err := v.loadCustom()
	if err != nil {
		return err
	}

	var clusterModules []string
	for _, m := range modules {
		meta, err := v.selectModuleMeta(m, embeddedMetas, customMetas)
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

func ValidateProviderSupport(envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) error {
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

func (v *ModuleValidator) loadEmbedded() (map[string]*config.ModuleMetadata, error) {
	if v.embeddedLoaded {
		return v.embeddedMetas, v.embeddedErr
	}
	v.embeddedLoaded = true
	v.embeddedMetas, v.embeddedErr = config.ScanModuleMetas(v.embeddedRoot)
	return v.embeddedMetas, v.embeddedErr
}

func (v *ModuleValidator) loadCustom() (map[string]*config.ModuleMetadata, error) {
	if v.customLoaded {
		return v.customMetas, v.customErr
	}
	v.customLoaded = true
	if strings.TrimSpace(v.customRoot) == "" {
		return nil, nil
	}
	v.customMetas, v.customErr = config.ScanModuleMetas(v.customRoot)
	return v.customMetas, v.customErr
}

func (v *ModuleValidator) selectModuleMeta(module config.Module, embeddedMetas, customMetas map[string]*config.ModuleMetadata) (*config.ModuleMetadata, error) {
	if config.IsCustomRootSource(module.Source) {
		if len(customMetas) == 0 {
			return nil, fmt.Errorf("module %q (type=%s) marked source=custom but no custom modules root provided", module.ID, module.Type)
		}
		meta, ok := customMetas[module.Type]
		if !ok {
			return nil, fmt.Errorf("module %q (type=%s) marked source=custom but metadata not found", module.ID, module.Type)
		}
		return meta, nil
	}

	meta, ok := embeddedMetas[module.Type]
	if !ok {
		if len(customMetas) > 0 {
			if alt, ok := customMetas[module.Type]; ok {
				return alt, nil
			}
		}
		return nil, fmt.Errorf("module %q type %q not found in embedded modules (%s); use source: custom with --modules", module.ID, module.Type, v.embeddedRoot)
	}
	return meta, nil
}
