package generate

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type clusterRef struct {
	host   hcl.Traversal
	caData hcl.Traversal
	token  hcl.Traversal
	auth   *authData
}

type authData struct {
	blockType string
	name      string
	attrs     map[string]hcl.Traversal
}

func (g *Generator) clusterRefs() *clusterRef {
	eksID := g.findFirstModuleByType("aws_eks")
	if eksID != "" {
		return &clusterRef{
			host: hcl.Traversal{
				hcl.TraverseRoot{Name: "module"},
				hcl.TraverseAttr{Name: eksID},
				hcl.TraverseAttr{Name: "k8s_endpoint"},
			},
			caData: hcl.Traversal{
				hcl.TraverseRoot{Name: "module"},
				hcl.TraverseAttr{Name: eksID},
				hcl.TraverseAttr{Name: "k8s_ca_data"},
			},
			token: hcl.Traversal{
				hcl.TraverseRoot{Name: "data"},
				hcl.TraverseAttr{Name: "aws_eks_cluster_auth"},
				hcl.TraverseAttr{Name: "this"},
				hcl.TraverseAttr{Name: "token"},
			},
			auth: &authData{
				blockType: "aws_eks_cluster_auth",
				name:      "this",
				attrs: map[string]hcl.Traversal{
					"name": {
						hcl.TraverseRoot{Name: "module"},
						hcl.TraverseAttr{Name: eksID},
						hcl.TraverseAttr{Name: "k8s_cluster_name"},
					},
				},
			},
		}
	}

	gkeID := g.findFirstModuleByType("gcp_gke")
	if gkeID != "" {
		return &clusterRef{
			host: hcl.Traversal{
				hcl.TraverseRoot{Name: "module"},
				hcl.TraverseAttr{Name: gkeID},
				hcl.TraverseAttr{Name: "endpoint"},
			},
			caData: hcl.Traversal{
				hcl.TraverseRoot{Name: "module"},
				hcl.TraverseAttr{Name: gkeID},
				hcl.TraverseAttr{Name: "cluster_ca_certificate"},
			},
			token: hcl.Traversal{
				hcl.TraverseRoot{Name: "data"},
				hcl.TraverseAttr{Name: "google_client_config"},
				hcl.TraverseAttr{Name: "default"},
				hcl.TraverseAttr{Name: "access_token"},
			},
			auth: &authData{
				blockType: "google_client_config",
				name:      "default",
				attrs:     map[string]hcl.Traversal{},
			},
		}
	}

	return nil
}

func base64DecodeTokens(inner hcl.Traversal) hclwrite.Tokens {
	toks := hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("base64decode")},
		&hclwrite.Token{Type: hclsyntax.TokenOParen, Bytes: []byte("(")},
	}
	toks = append(toks, hclwrite.TokensForTraversal(inner)...)
	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenCParen, Bytes: []byte(")")})
	return toks
}

func writeAuthData(body *hclwrite.Body, auth *authData) {
	if auth == nil {
		return
	}
	block := body.AppendNewBlock("data", []string{auth.blockType, auth.name})
	b := block.Body()
	for k, v := range auth.attrs {
		b.SetAttributeRaw(k, hclwrite.TokensForTraversal(v))
	}
	body.AppendNewline()
}

func helmKubeConfigTokens(cluster *clusterRef) hclwrite.Tokens {
	if cluster == nil {
		return nil
	}

	var toks hclwrite.Tokens
	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}})
	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})

	addAttr := func(name string, val hclwrite.Tokens) {
		toks = append(toks, hclwrite.TokensForIdentifier(name)...)
		toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}})
		toks = append(toks, val...)
		toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}

	addAttr("host", hclwrite.TokensForTraversal(cluster.host))
	addAttr("cluster_ca_certificate", base64DecodeTokens(cluster.caData))
	addAttr("token", hclwrite.TokensForTraversal(cluster.token))

	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}})
	return toks
}

func (g *Generator) hasModuleType(t string) bool {
	for _, m := range g.allModules {
		if m.Type == t {
			return true
		}
	}
	return false
}
