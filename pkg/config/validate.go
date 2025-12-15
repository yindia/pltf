package config

import (
	"fmt"
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
	}

	if _, err := validateModules(e.Modules, "environment"); err != nil {
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

	if _, err := validateModules(s.Modules, "service"); err != nil {
		return err
	}

	return nil
}

// validateModules enforces module id/type presence, uniqueness, and link targets.
func validateModules(mods []Module, context string) (map[string]struct{}, error) {
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
			return nil, fmt.Errorf("module %q type is required%s", m.ID, contextSuffix(context))
		}
		if _, exists := ids[m.ID]; exists {
			return nil, fmt.Errorf("duplicate module id %q%s", m.ID, contextSuffix(context))
		}
		ids[m.ID] = struct{}{}
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

func contextSuffix(context string) string {
	if context == "" {
		return ""
	}
	return fmt.Sprintf(" in %s", context)
}
