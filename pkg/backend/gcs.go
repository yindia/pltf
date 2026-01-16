package backend

import (
	"context"
	"strings"

	"pltf/pkg/config"
)

type gcsBackend struct{}

func (g gcsBackend) Type() string { return "gcs" }

func (g gcsBackend) Resolve(envCfg *config.EnvironmentConfig, _ config.EnvironmentEntry) (Config, error) {
	bucket := strings.TrimSpace(envCfg.Backend.Bucket)
	if bucket == "" {
		bucket = defaultBackendBucket("gcs", envCfg.Metadata.Org, envCfg.Metadata.Name)
	}
	region := strings.TrimSpace(envCfg.Backend.Region)
	return Config{
		Type:          "gcs",
		Bucket:        bucket,
		Region:        region,
		Container:     strings.TrimSpace(envCfg.Backend.Container),
		ResourceGroup: strings.TrimSpace(envCfg.Backend.ResourceGroup),
		Profile:       strings.TrimSpace(envCfg.Backend.Profile),
	}, nil
}

func (g gcsBackend) Ensure(_ context.Context, _ Config) error {
	return nil
}

func init() {
	RegisterProvider(gcsBackend{})
}
