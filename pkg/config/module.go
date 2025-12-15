package config

import (
	"fmt"
)

// ModuleMetadata describes a Terraform module and its contract to the platform.
type ModuleMetadata struct {
	Name         string       `yaml:"name"`     // e.g. "aws-s3"
	Type         string       `yaml:"type"`     // e.g. "aws-s3" (matches config modules[].type)
	Provider     string       `yaml:"provider"` // "aws", "kubernetes", etc.
	Version      string       `yaml:"version"`  // "1.0.0"
	Description  string       `yaml:"description,omitempty"`
	Capabilities Capabilities `yaml:"capabilities"` // what it provides/accepts
	Inputs       []InputSpec  `yaml:"inputs,omitempty"`
	Outputs      []OutputSpec `yaml:"outputs,omitempty"`
}

// Capabilities this module exposes.
type Capabilities struct {
	Provides []string `yaml:"provides,omitempty"` // e.g. ["storage.s3", "iam.resourceArn"]
	Accepts  []string `yaml:"accepts,omitempty"`  // e.g. ["runtime.kubernetesCluster"]
}

type InputSpec struct {
	Name        string      `yaml:"name"`                  // input variable name
	Type        string      `yaml:"type"`                  // "string", "number", "bool", "list", "map", etc.
	Required    bool        `yaml:"required"`              // if true, must be set in stack
	Default     interface{} `yaml:"default,omitempty"`     // default value (optional)
	Description string      `yaml:"description,omitempty"` // docstring
	Capability  string      `yaml:"capability,omitempty"`
}

type OutputSpec struct {
	Name        string `yaml:"name"`                  // output variable name
	Type        string `yaml:"type"`                  // "string", "number", etc.
	Capability  string `yaml:"capability,omitempty"`  // optional semantic tag: "iam.principal", "iam.resourceArn", etc.
	Description string `yaml:"description,omitempty"` // docstring
}

// Validate checks the ModuleMetadata for structural issues and generic capability sanity.
func (m *ModuleMetadata) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Type == "" {
		return fmt.Errorf("type is required (must match stack modules[].type)")
	}
	if m.Provider == "" {
		return fmt.Errorf("provider is required (e.g. aws, kubernetes)")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}

	// ---------- Inputs ----------
	inputNames := make(map[string]struct{})
	for i, in := range m.Inputs {
		if in.Name == "" {
			return fmt.Errorf("input name is required")
		}
		if in.Type == "" {
			// treat missing type as "any" to accommodate existing module.yaml files
			m.Inputs[i].Type = "any"
		}
		if _, exists := inputNames[in.Name]; exists {
			return fmt.Errorf("duplicate input name %q", in.Name)
		}
		inputNames[in.Name] = struct{}{}

	}

	// ---------- Outputs ----------
	outputNames := make(map[string]struct{})
	for _, out := range m.Outputs {
		if out.Name == "" {
			return fmt.Errorf("output name is required")
		}
		if _, exists := outputNames[out.Name]; exists {
			return fmt.Errorf("duplicate output name %q", out.Name)
		}
		outputNames[out.Name] = struct{}{}

		if out.Type == "" {
			return fmt.Errorf("output %q type is required", out.Name)
		}
	}

	// ---------- Capabilities ----------
	// No duplicates in provides/accepts
	providesSeen := make(map[string]struct{})
	for _, cap := range m.Capabilities.Provides {
		if cap == "" {
			return fmt.Errorf("empty capability in capabilities.provides")
		}
		if _, exists := providesSeen[cap]; exists {
			return fmt.Errorf("duplicate capability %q in capabilities.provides", cap)
		}
		providesSeen[cap] = struct{}{}
	}

	acceptsSeen := make(map[string]struct{})
	for _, cap := range m.Capabilities.Accepts {
		if cap == "" {
			return fmt.Errorf("empty capability in capabilities.accepts")
		}
		if _, exists := acceptsSeen[cap]; exists {
			return fmt.Errorf("duplicate capability %q in capabilities.accepts", cap)
		}
		acceptsSeen[cap] = struct{}{}
	}

	// Generic rule: if an output declares a capability, it must be in provides.
	for _, out := range m.Outputs {
		if out.Capability == "" {
			continue
		}
		if _, ok := providesSeen[out.Capability]; !ok {
			return fmt.Errorf(
				"output %q declares capability %q but it is not listed in capabilities.provides",
				out.Name,
				out.Capability,
			)
		}
	}

	return nil
}
