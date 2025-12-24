package config

// ServiceConfig is the root for kind: Service
type ServiceConfig struct {
	APIVersion string          `yaml:"apiVersion"`
	Source     string          `yaml:"source"`
	Directory  string          `yaml:"directory"`
	Version    string          `yaml:"version"`
	Kind       string          `yaml:"kind"` // should be "Service"
	Backend    Backend         `yaml:"backend"`
	Metadata   ServiceMetadata `yaml:"metadata"`
	Modules    []Module        `yaml:"modules"`
}
type ServiceMetadata struct {
	Name    string                        `yaml:"name"`
	Ref     string                        `yaml:"ref"`    // path to env.yaml
	EnvRef  map[string]ServiceEnvRefEntry `yaml:"envRef"` // dev, prod, ...
	Labels  map[string]string             `yaml:"labels,omitempty"`
	// Approve []ApprovalRequirement         `yaml:"approve,omitempty"`
}

type ServiceEnvRefEntry struct {
	Variables map[string]string    `yaml:"variables,omitempty"`
	Secrets   map[string]SecretRef `yaml:"secrets,omitempty"`
}
