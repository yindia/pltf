package generate

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// writeKubernetesHelmProviders adds kubernetes and helm providers if an EKS module is present.
func (g *Generator) writeKubernetesHelmProviders(body *hclwrite.Body) {
	eksID := g.findFirstModuleByType("aws_eks")
	if eksID == "" {
		return
	}

	// data "aws_eks_cluster_auth" "this" is expected to exist in modules; we reference its token.
	k8sBlock := body.AppendNewBlock("provider", []string{"kubernetes"})
	k8sBody := k8sBlock.Body()
	k8sBody.SetAttributeRaw("host", hclwrite.TokensForValue(cty.StringVal("${module."+eksID+".k8s_endpoint}")))
	k8sBody.SetAttributeRaw("cluster_ca_certificate", hclwrite.TokensForValue(cty.StringVal("${base64decode(module."+eksID+".k8s_ca_data)}")))
	k8sBody.SetAttributeRaw("token", hclwrite.TokensForValue(cty.StringVal("${data.aws_eks_cluster_auth.this.token}")))

	helmBlock := body.AppendNewBlock("provider", []string{"helm"})
	helmBody := helmBlock.Body()
	helmBody.SetAttributeRaw("kubernetes", hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
		&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
	})
	helmBody.SetAttributeRaw("kubernetes.host", hclwrite.TokensForValue(cty.StringVal("${module."+eksID+".k8s_endpoint}")))
	helmBody.SetAttributeRaw("kubernetes.cluster_ca_certificate", hclwrite.TokensForValue(cty.StringVal("${base64decode(module."+eksID+".k8s_ca_data)}")))
	helmBody.SetAttributeRaw("kubernetes.token", hclwrite.TokensForValue(cty.StringVal("${data.aws_eks_cluster_auth.this.token}")))
}
