package backend

import (
	"context"
	"fmt"
	"strings"

	"pltf/pkg/config"
)

// Config describes the resolved backend config for Terraform state.
type Config struct {
	Type          string
	Bucket        string
	Region        string
	Container     string
	ResourceGroup string
	Profile       string
	AccountID     string
}

// Provider defines a backend implementation.
type Provider interface {
	Type() string
	Resolve(envCfg *config.EnvironmentConfig, envEntry config.EnvironmentEntry) (Config, error)
	Ensure(ctx context.Context, cfg Config) error
}

var registry = map[string]Provider{}

// RegisterProvider registers a backend provider by type name.
func RegisterProvider(p Provider) {
	if p == nil {
		return
	}
	registry[strings.ToLower(strings.TrimSpace(p.Type()))] = p
}

// Resolve selects and resolves a backend based on env config and provider defaults.
func Resolve(envProvider string, envCfg *config.EnvironmentConfig, envEntry config.EnvironmentEntry) (Config, error) {
	if envCfg == nil {
		return Config{}, fmt.Errorf("environment config is required")
	}
	provider := canonicalProvider(envProvider)
	if provider == "" {
		return Config{}, fmt.Errorf("unsupported provider %q for backend", envProvider)
	}
	backendType := strings.TrimSpace(envCfg.Backend.Type)
	if backendType == "" {
		var err error
		backendType, err = defaultBackendType(provider)
		if err != nil {
			return Config{}, err
		}
	}
	backendType = strings.ToLower(backendType)
	if err := validateBackendProvider(provider, backendType); err != nil {
		return Config{}, err
	}
	p, ok := registry[backendType]
	if !ok {
		return Config{}, fmt.Errorf("unsupported backend type %q", backendType)
	}
	return p.Resolve(envCfg, envEntry)
}

// Ensure calls provider-specific ensure logic, when available.
func Ensure(ctx context.Context, cfg Config) error {
	p, ok := registry[strings.ToLower(strings.TrimSpace(cfg.Type))]
	if !ok {
		return fmt.Errorf("unsupported backend type %q", cfg.Type)
	}
	return p.Ensure(ctx, cfg)
}

func defaultBackendType(provider string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "aws":
		return "s3", nil
	case "gcp":
		return "gcs", nil
	case "azure":
		return "azurerm", nil
	default:
		return "", fmt.Errorf("unsupported provider %q for backend default", provider)
	}
}

func canonicalProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "aws":
		return "aws"
	case "gcp", "google":
		return "gcp"
	case "azure", "azurerm":
		return "azure"
	default:
		return ""
	}
}

func validateBackendProvider(provider, backendType string) error {
	switch backendType {
	case "s3":
		if provider != "aws" {
			return fmt.Errorf("backend type %q requires provider aws", backendType)
		}
	case "gcs":
		if provider != "gcp" {
			return fmt.Errorf("backend type %q requires provider gcp", backendType)
		}
	case "azurerm":
		if provider != "azure" {
			return fmt.Errorf("backend type %q requires provider azure", backendType)
		}
	default:
		return fmt.Errorf("unsupported backend type %q", backendType)
	}
	return nil
}
