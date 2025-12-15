package config

// EnvironmentConfig is the root for kind: Environment
type EnvironmentConfig struct {
	APIVersion   string                      `yaml:"apiVersion"`
	Kind         string                      `yaml:"kind"` // should be "Environment"
	Metadata     EnvironmentMetadata         `yaml:"metadata"`
	Backend      Backend                     `yaml:"backend"`
	Environments map[string]EnvironmentEntry `yaml:"environments"` // dev, prod, ...
	Modules      []Module                    `yaml:"modules"`
}

type EnvironmentMetadata struct {
	Name     string            `yaml:"name"`
	Org      string            `yaml:"org"`
	Provider string            `yaml:"provider"` // "aws", etc.
	Labels   map[string]string `yaml:"labels"`
}

type EnvironmentEntry struct {
	Account   string               `yaml:"account"`             // "111111111111"
	Region    string               `yaml:"region"`              // provider region per environment
	Variables map[string]string    `yaml:"variables,omitempty"` // cluster_name, base_domain, ...
	Secrets   map[string]SecretRef `yaml:"secrets,omitempty"`
}
