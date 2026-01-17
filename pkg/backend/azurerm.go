package backend

import (
	"context"
	"strings"

	"pltf/pkg/config"
)

type azurermBackend struct{}

func (a azurermBackend) Type() string { return "azurerm" }

func (a azurermBackend) Resolve(envCfg *config.EnvironmentConfig, _ config.EnvironmentEntry) (Config, error) {
	bucket := strings.TrimSpace(envCfg.Backend.Bucket)
	if bucket == "" {
		bucket = defaultBackendBucket("azurerm", envCfg.Metadata.Org, envCfg.Metadata.Name)
	}
	region := strings.TrimSpace(envCfg.Backend.Region)
	return Config{
		Type:          "azurerm",
		Bucket:        bucket,
		Region:        region,
		Container:     strings.TrimSpace(envCfg.Backend.Container),
		ResourceGroup: strings.TrimSpace(envCfg.Backend.ResourceGroup),
		Profile:       strings.TrimSpace(envCfg.Backend.Profile),
	}, nil
}

func (a azurermBackend) Ensure(_ context.Context, _ Config) error {
	return nil
}

func init() {
	RegisterProvider(azurermBackend{})
}
