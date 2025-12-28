package cloud

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Provider is the interface for cloud providers.
type Provider interface {
	// RequiredProviders returns the HCL block for the `required_providers` section in `versions.tf`.
	RequiredProviders(body *hclwrite.Body, needsK8s bool, needsHelm bool)
	// Backend returns the HCL block for the `backend` section in `versions.tf`.
	Backend(body *hclwrite.Body, bucket, key, region, profile, container, resourceGroup string)
	// Provider returns the HCL block for the `provider` section in `providers.tf`.
	Provider(body *hclwrite.Body, region, account string)
}

// New returns a new cloud provider based on the given provider name.
func New(provider string) (Provider, error) {
	switch provider {
	case "aws", "":
		return NewAWS(), nil
	case "gcp", "google":
		return NewGCP(), nil
	case "azure", "azurerm":
		return NewAzure(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
