package cloud

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"pltf/pkg/provider"
)

// GCP is the GCP provider.
type GCP struct{}

// NewGCP creates a new GCP provider.
func NewGCP() *GCP {
	return &GCP{}
}

// RequiredProviders adds the required providers for GCP to the HCL body.
func (g *GCP) RequiredProviders(body *hclwrite.Body, needsK8s bool, needsHelm bool) {
	body.SetAttributeValue("google", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/google"),
		"version": cty.StringVal(provider.GCPProviderVersion),
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

// Backend adds the GCS backend to the HCL body.
func (g *GCP) Backend(body *hclwrite.Body, bucket, key, _, _, _, _ string) {
	backendBlock := body.AppendNewBlock("backend", []string{"gcs"})
	bb := backendBlock.Body()
	bb.SetAttributeValue("bucket", cty.StringVal(bucket))
	bb.SetAttributeValue("prefix", cty.StringVal(key))
}

// Provider adds the GCP provider to the HCL body.
func (g *GCP) Provider(body *hclwrite.Body, region, project string) {
	provBlock := body.AppendNewBlock("provider", []string{"google"})
	provBody := provBlock.Body()
	provBody.SetAttributeValue("project", cty.StringVal(project))
	provBody.SetAttributeValue("region", cty.StringVal(region))
}
