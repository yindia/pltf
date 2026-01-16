package cloud

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"pltf/pkg/provider"
)

// AWS is the AWS provider.
type AWS struct{}

// NewAWS creates a new AWS provider.
func NewAWS() *AWS {
	return &AWS{}
}

// RequiredProviders adds the required providers for AWS to the HCL body.
func (a *AWS) RequiredProviders(body *hclwrite.Body, needsK8s bool, needsHelm bool, needsKustomize bool) {
	body.SetAttributeValue("aws", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/aws"),
		"version": cty.StringVal(provider.ResolveProviderVersion("aws")),
	}))
	if needsK8s {
		body.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/kubernetes"),
			"version": cty.StringVal(provider.ResolveProviderVersion("kubernetes")),
		}))
	}
	if needsHelm {
		body.SetAttributeValue("helm", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/helm"),
			"version": cty.StringVal(provider.ResolveProviderVersion("helm")),
		}))
	}
	if needsKustomize {
		body.SetAttributeValue("kustomization", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("kbst/kustomization"),
			"version": cty.StringVal(provider.ResolveProviderVersion("kustomization")),
		}))
	}
}

// Backend adds the S3 backend to the HCL body.
func (a *AWS) Backend(body *hclwrite.Body, bucket, key, region, profile, _, _, workspaceKeyPrefix string) {
	backendBlock := body.AppendNewBlock("backend", []string{"s3"})
	bb := backendBlock.Body()
	bb.SetAttributeValue("bucket", cty.StringVal(bucket))
	bb.SetAttributeValue("key", cty.StringVal(key))
	bb.SetAttributeValue("region", cty.StringVal(region))
	if strings.TrimSpace(profile) != "" {
		bb.SetAttributeValue("profile", cty.StringVal(profile))
	}
	if strings.TrimSpace(workspaceKeyPrefix) != "" {
		bb.SetAttributeValue("workspace_key_prefix", cty.StringVal(workspaceKeyPrefix))
	}
}

// Provider adds the AWS provider to the HCL body.
func (a *AWS) Provider(body *hclwrite.Body, region, _ string, useVarRefs bool) {
	provBlock := body.AppendNewBlock("provider", []string{"aws"})
	provBody := provBlock.Body()
	if useVarRefs {
		provBody.SetAttributeRaw("region", hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "local"},
			hcl.TraverseAttr{Name: "region"},
		}))
	} else {
		provBody.SetAttributeValue("region", cty.StringVal(region))
	}
	dt := provBody.AppendNewBlock("default_tags", nil)
	tags := dt.Body()
	tags.SetAttributeRaw("tags", provider.DefaultTagsTokens())
}
