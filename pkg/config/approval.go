package config

// ApprovalRequirement defines an approval gate for a spec.
type ApprovalRequirement struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias,omitempty"`
	Count int    `yaml:"count,omitempty"`
}
