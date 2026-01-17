package config

// StackConfig is the root for kind: Stack
type StackConfig struct {
	APIVersion string               `yaml:"apiVersion"`
	Kind       string               `yaml:"kind"` // should be "Stack"
	Metadata   StackMetadata        `yaml:"metadata"`
	Variables  map[string]string    `yaml:"variables,omitempty"`
	Providers  ProviderRequirements `yaml:"providers,omitempty"`
	Modules    []Module             `yaml:"modules"`
}

type StackMetadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels,omitempty"`
}
