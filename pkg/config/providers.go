package config

// ProviderRequirements declares which Terraform providers must be configured.
type ProviderRequirements struct {
	Kubernetes bool `yaml:"kubernetes,omitempty"`
	Helm       bool `yaml:"helm,omitempty"`
	Kustomize  bool `yaml:"kustomize,omitempty"`
}

type ProviderExplicit struct {
	KubernetesSet bool
	HelmSet       bool
	KustomizeSet  bool
}
