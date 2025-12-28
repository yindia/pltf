package provider

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	RequiredTfVersion    = ">= 1.5.7"
	AWSProviderVersion   = "~> 6.0"
	GCPProviderVersion   = ">= 5.0.0"
	K8sProviderVersion   = ">= 2.30.0"
	HelmProviderVersion  = ">= 2.13.2"
	AzureProviderVersion = ">= 4.0.0"
)

func DefaultTagsTokens() hclwrite.Tokens {
	toks := hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("merge")},
		&hclwrite.Token{Type: hclsyntax.TokenOParen, Bytes: []byte("(")},
		&hclwrite.Token{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("Environment")},
		&hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte("=")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("local")},
		&hclwrite.Token{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("environment")},
		&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("Owner")},
		&hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte("=")},
	}
	toks = append(toks, hclwrite.TokensForValue(cty.StringVal("PlatformTeam"))...)
	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")})
	toks = append(toks,
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("terraform")},
		&hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte("=")},
	)
	toks = append(toks, hclwrite.TokensForValue(cty.StringVal("true"))...)
	toks = append(toks,
		&hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
		&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("local")},
		&hclwrite.Token{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
		&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("global_tags")},
		&hclwrite.Token{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
	)
	return toks
}
