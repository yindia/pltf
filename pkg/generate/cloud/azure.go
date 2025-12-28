package cloud

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"pltf/pkg/provider"
)

// Azure is the Azure provider.
type Azure struct{}

// NewAzure creates a new Azure provider.
func NewAzure() *Azure {
	return &Azure{}
}

// RequiredProviders adds the required providers for Azure to the HCL body.
func (a *Azure) RequiredProviders(body *hclwrite.Body, needsK8s bool, needsHelm bool) {
	body.SetAttributeValue("azurerm", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/azurerm"),
		"version": cty.StringVal(provider.AzureProviderVersion),
	}))
	if needsK8s {
		body.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/kubernetes"),
			"version": cty.StringVal(provider.K8sProviderVersion),
		}))
	}
	if needsHelm {
		body.SetAttributeValue("helm", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/helm"),
			"version": cty.StringVal(provider.HelmProviderVersion),
		}))
	}
}

// Backend adds the Azure RM backend to the HCL body.
func (a *Azure) Backend(body *hclwrite.Body, bucket, key, _, _, container, resourceGroup string) {
	if bucket == "" {
		panic("backend.bucket (storage account name) is required for azure")
	}
	if container == "" {
		container = "tfstate"
	}
	backendBlock := body.AppendNewBlock("backend", []string{"azurerm"})
	bb := backendBlock.Body()
	bb.SetAttributeValue("storage_account_name", cty.StringVal(bucket))
	bb.SetAttributeValue("container_name", cty.StringVal(container))
	bb.SetAttributeValue("key", cty.StringVal(key))
	if resourceGroup != "" {
		bb.SetAttributeValue("resource_group_name", cty.StringVal(resourceGroup))
	}
}

// Provider adds the Azure provider to the HCL body.
func (a *Azure) Provider(body *hclwrite.Body, _, subscriptionID string) {
	provBlock := body.AppendNewBlock("provider", []string{"azurerm"})
	provBody := provBlock.Body()
	provBody.SetAttributeValue("subscription_id", cty.StringVal(subscriptionID))
	provBody.SetAttributeValue("features", cty.ObjectVal(map[string]cty.Value{}))
}
