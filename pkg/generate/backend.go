package generate

import (
	"fmt"
	"strings"

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
	if envCfg == nil {
		return BackendConfig{}, fmt.Errorf("environment config is required")
	}

	backendType := strings.TrimSpace(envCfg.Backend.Type)
	if backendType == "" {
		var err error
		backendType, err = defaultBackendType(provider)
		if err != nil {
			return BackendConfig{}, err
		}
	}
	backendType = strings.ToLower(backendType)

	switch backendType {
	case "s3", "gcs", "azurerm":
	default:
		return BackendConfig{}, fmt.Errorf("unsupported backend type %q", backendType)
	}

	bucket := strings.TrimSpace(envCfg.Backend.Bucket)
	if bucket == "" {
		bucket = defaultBackendBucket(backendType, envCfg.Metadata.Org, envCfg.Metadata.Name)
	}

	region := strings.TrimSpace(envCfg.Backend.Region)
	if region == "" && backendType == "s3" {
		region = envEntry.Region
	}

	return BackendConfig{
		BackendType:   backendType,
		Bucket:        bucket,
		Region:        region,
		Container:     strings.TrimSpace(envCfg.Backend.Container),
		ResourceGroup: strings.TrimSpace(envCfg.Backend.ResourceGroup),
		Profile:       strings.TrimSpace(envCfg.Backend.Profile),
	}, nil
}

func defaultBackendType(provider string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "aws":
		return "s3", nil
	case "gcp", "google":
		return "gcs", nil
	case "azure", "azurerm":
		return "azurerm", nil
	default:
		return "", fmt.Errorf("unsupported provider %q for backend default", provider)
	}
}

func defaultBackendBucket(backendType, org, envName string) string {
	var parts []string
	if strings.TrimSpace(org) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(org)))
	} else {
		parts = append(parts, "pltf")
	}
	if strings.TrimSpace(envName) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(envName)))
	}
	parts = append(parts, "tfstate")
	base := strings.Join(parts, "-")

	switch backendType {
	case "azurerm":
		return sanitizeStorageAccountName(base)
	default:
		return sanitizeBucketName(base)
	}
}

func sanitizeBucketName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "pltf-tfstate"
	}
	if len(out) < 3 {
		out = out + strings.Repeat("0", 3-len(out))
	}
	if len(out) > 63 {
		out = strings.Trim(out[:63], "-")
	}
	if len(out) < 3 {
		out = "pltf-tfstate"
		if len(out) > 63 {
			out = out[:63]
		}
	}
	return out
}

func sanitizeStorageAccountName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}

	out := b.String()
	if out == "" {
		out = "pltftfstate"
	}
	if len(out) < 3 {
		out = out + strings.Repeat("0", 3-len(out))
	}
	if len(out) > 24 {
		out = out[:24]
	}
	return out
}
