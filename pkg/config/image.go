package config

// ImageBuild defines a Docker image build configuration.
type ImageBuild struct {
	Name       string            `yaml:"name"`
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Include    []string          `yaml:"include,omitempty"`
	Exclude    []string          `yaml:"exclude,omitempty"`
	Target     string            `yaml:"target,omitempty"`
	Tags       []string          `yaml:"tags,omitempty"`
	BuildArgs  map[string]string `yaml:"buildArgs,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
	Platforms  []string          `yaml:"platforms,omitempty"`
}
