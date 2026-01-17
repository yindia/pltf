package generate

import (
	"pltf/pkg/backend"
	"pltf/pkg/config"
)

type BackendConfig struct {
	BackendType   string
	Bucket        string
	Region        string
	Container     string
	ResourceGroup string
	Profile       string
}

func ResolveBackendConfig(provider string, envCfg *config.EnvironmentConfig, envEntry config.EnvironmentEntry) (BackendConfig, error) {
	cfg, err := backend.Resolve(provider, envCfg, envEntry)
	if err != nil {
		return BackendConfig{}, err
	}
	return BackendConfig{
		BackendType:   cfg.Type,
		Bucket:        cfg.Bucket,
		Region:        cfg.Region,
		Container:     cfg.Container,
		ResourceGroup: cfg.ResourceGroup,
		Profile:       cfg.Profile,
	}, nil
}
