package config

// ServiceConfig is the root for kind: Service
type ServiceConfig struct {
	APIVersion        string               `yaml:"apiVersion"`
	Kind              string               `yaml:"kind"` // should be "Service"
	Backend           Backend              `yaml:"backend"`
	Variables         map[string]string    `yaml:"variables,omitempty"`
	Secrets           map[string]SecretRef `yaml:"secrets,omitempty"`
	Providers         ProviderRequirements `yaml:"providers,omitempty"`
	ProvidersExplicit ProviderExplicit     `yaml:"-"`
	Images            []ImageBuild         `yaml:"images,omitempty"`
	Metadata          ServiceMetadata      `yaml:"metadata"`
	Modules           []Module             `yaml:"modules"`
	GitProvider       GitProvider          `yaml:"gitProvider,omitempty"`
}
type ServiceMetadata struct {
	Name   string                        `yaml:"name"`
	Ref    string                        `yaml:"ref"`    // path to env.yaml
	EnvRef map[string]ServiceEnvRefEntry `yaml:"envRef"` // dev, prod, ...
	Labels map[string]string             `yaml:"labels,omitempty"`
	Stacks []string                      `yaml:"stacks,omitempty"`
}

type ServiceEnvRefEntry struct {
	Variables map[string]string    `yaml:"variables,omitempty"`
	Secrets   map[string]SecretRef `yaml:"secrets,omitempty"`
}
