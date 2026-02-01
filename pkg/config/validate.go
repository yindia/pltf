package config

import (
	"fmt"
	"strings"
)

// Validate checks the EnvironmentConfig for structural issues.
func (e *EnvironmentConfig) Validate() error {
	if e.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if e.Kind != "Environment" {
		return fmt.Errorf("kind must be 'Environment', got %q", e.Kind)
	}
	if e.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if e.Metadata.Org == "" {
		return fmt.Errorf("metadata.org is required")
	}
	if e.Metadata.Provider == "" {
		return fmt.Errorf("metadata.provider is required")
	}
	if e.GitProvider == "" {
		e.GitProvider = GitProviderGitHub
	}
	if err := validateGitProvider("metadata.gitProvider", string(e.GitProvider)); err != nil {
		return err
	}

	if len(e.Environments) == 0 {
		return fmt.Errorf("at least one environment entry is required")
	}

	for envName, envEntry := range e.Environments {
		if envEntry.Account == "" {
			return fmt.Errorf("environments.%s.account is required", envName)
		}
		if envEntry.Region == "" {
			return fmt.Errorf("environments.%s.region is required", envName)
		}
		if len(envEntry.Variables) > 0 {
			return fmt.Errorf("environments.%s.variables is not supported; use top-level variables instead", envName)
		}
		if len(envEntry.Secrets) > 0 {
			return fmt.Errorf("environments.%s.secrets is not supported; use top-level secrets instead", envName)
		}
	}

	if _, err := validateModules(e.Modules, "environment"); err != nil {
		return err
	}
	if err := validateImages(e.Images, "environment"); err != nil {
		return err
	}

	return nil
}

// Validate checks the ServiceConfig for structural issues.
// If env is non-nil, it will also validate envRef entries against Environment.
func (s *ServiceConfig) Validate(env *EnvironmentConfig) error {
	if s.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if s.Kind != "Service" {
		return fmt.Errorf("kind must be 'Service', got %q", s.Kind)
	}
	if s.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if s.Metadata.Ref == "" {
		return fmt.Errorf("metadata.ref (path to environment) is required")
	}
	if s.GitProvider == "" {
		s.GitProvider = GitProviderGitHub
	}
	if err := validateGitProvider("metadata.gitProvider", string(s.GitProvider)); err != nil {
		return err
	}

	if len(s.Metadata.EnvRef) == 0 {
		return fmt.Errorf("metadata.envRef must define at least one environment (dev/prod, etc.)")
	}

	if env != nil {
		for envName := range s.Metadata.EnvRef {
			if _, ok := env.Environments[envName]; !ok {
				return fmt.Errorf("service envRef.%s has no matching environment in %s", envName, env.Metadata.Name)
			}
		}
	}
	if env != nil {
		envModuleIDs := make(map[string]struct{})
		for _, m := range env.Modules {
			envModuleIDs[m.ID] = struct{}{}
		}
		for _, m := range s.Modules {
			if _, exists := envModuleIDs[m.ID]; exists {
				return fmt.Errorf("service module %q duplicates an environment module id", m.ID)
			}
		}
	}
	for envName, envEntry := range s.Metadata.EnvRef {
		if len(envEntry.Variables) > 0 {
			return fmt.Errorf("metadata.envRef.%s.variables is not supported; use top-level variables instead", envName)
		}
		if len(envEntry.Secrets) > 0 {
			return fmt.Errorf("metadata.envRef.%s.secrets is not supported; use top-level secrets instead", envName)
		}
	}

	if _, err := validateModules(s.Modules, "service"); err != nil {
		return err
	}
	if err := validateImages(s.Images, "service"); err != nil {
		return err
	}

	return nil
}

// Validate checks the StackConfig for structural issues.
func (s *StackConfig) Validate() error {
	if s.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if s.Kind != "Stack" {
		return fmt.Errorf("kind must be 'Stack', got %q", s.Kind)
	}
	if s.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if _, err := validateModulesNoLinks(s.Modules, "stack"); err != nil {
		return err
	}
	return nil
}

// validateModules enforces module id/type presence, uniqueness, and link targets.
func validateModules(mods []Module, context string) (map[string]struct{}, error) {
	ids, err := validateModulesNoLinks(mods, context)
	if err != nil {
		return nil, err
	}

	for _, m := range mods {
		for access, targets := range m.Links {
			if len(targets) == 0 {
				return nil, fmt.Errorf("module %q links.%s has no targets%s", m.ID, access, contextSuffix(context))
			}
			for _, t := range targets {
				if _, ok := ids[t]; !ok {
					return nil, fmt.Errorf("module %q links.%s refers to unknown module %q%s", m.ID, access, t, contextSuffix(context))
				}
			}
		}
	}

	return ids, nil
}

func validateModulesNoLinks(mods []Module, context string) (map[string]struct{}, error) {
	if len(mods) == 0 {
		if context == "" {
			return nil, fmt.Errorf("at least one module is required")
		}
		return nil, fmt.Errorf("at least one module is required in %s", context)
	}

	ids := make(map[string]struct{})
	for _, m := range mods {
		if m.ID == "" {
			return nil, fmt.Errorf("module id is required%s", contextSuffix(context))
		}
		if m.Type == "" {
			if IsGitSource(m.Source) || IsLocalSource(m.Source) {
				continue
			}
			return nil, fmt.Errorf("module %q type is required%s", m.ID, contextSuffix(context))
		}
		if _, exists := ids[m.ID]; exists {
			return nil, fmt.Errorf("duplicate module id %q%s", m.ID, contextSuffix(context))
		}
		ids[m.ID] = struct{}{}
	}

	return ids, nil
}

func contextSuffix(context string) string {
	if context == "" {
		return ""
	}
	return fmt.Sprintf(" in %s", context)
}

func validateGitProvider(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	switch normalizeGitProvider(value) {
	case string(GitProviderGitHub), string(GitProviderGitLab), string(GitProviderBitbucket):
		return nil
	default:
		return fmt.Errorf("%s must be one of: github, gitlab, bitbucket", field)
	}
}

func validateImages(images []ImageBuild, context string) error {
	if len(images) == 0 {
		return nil
	}
	names := make(map[string]struct{})
	for i, img := range images {
		index := fmt.Sprintf("%s images[%d]", context, i)
		name := strings.TrimSpace(img.Name)
		if name == "" {
			return fmt.Errorf("%s.name is required", index)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("duplicate image name %q in %s", name, context)
		}
		names[name] = struct{}{}

		if strings.TrimSpace(img.Context) == "" {
			return fmt.Errorf("%s.context is required", index)
		}
		if len(img.Tags) > 0 {
			tagSeen := make(map[string]struct{})
			for _, tag := range img.Tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					return fmt.Errorf("%s.tags contains an empty tag", index)
				}
				if _, ok := tagSeen[tag]; ok {
					return fmt.Errorf("%s.tags has duplicate tag %q", index, tag)
				}
				tagSeen[tag] = struct{}{}
			}
		}
		for key, val := range img.BuildArgs {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s.buildArgs contains an empty key", index)
			}
			if strings.TrimSpace(val) == "" {
				return fmt.Errorf("%s.buildArgs.%s is empty", index, key)
			}
		}
		for key, val := range img.Labels {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s.labels contains an empty key", index)
			}
			if strings.TrimSpace(val) == "" {
				return fmt.Errorf("%s.labels.%s is empty", index, key)
			}
		}
		for _, pattern := range img.Include {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("%s.include contains an empty pattern", index)
			}
		}
		for _, pattern := range img.Exclude {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("%s.exclude contains an empty pattern", index)
			}
		}
	}
	return nil
}
