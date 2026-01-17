package provider

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	RequiredTfVersion        = ">= 1.6.6"
	AWSProviderVersion       = "~> 6.0"
	GCPProviderVersion       = ">= 5.0.0"
	K8sProviderVersion       = ">= 2.30.0"
	HelmProviderVersion      = ">= 2.13.2"
	KustomizeProviderVersion = ">= 0.9.0"
	AzureProviderVersion     = ">= 4.0.0"
)

// ResolveProviderVersion allows overriding provider versions via env vars.
// Example: PLTF_PROVIDER_AWS_VERSION="~> 6.10"
func ResolveProviderVersion(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "aws":
		return envOrDefault("PLTF_PROVIDER_AWS_VERSION", AWSProviderVersion)
	case "gcp", "google":
		return envOrDefault("PLTF_PROVIDER_GCP_VERSION", GCPProviderVersion)
	case "azurerm", "azure":
		return envOrDefault("PLTF_PROVIDER_AZURE_VERSION", AzureProviderVersion)
	case "kubernetes", "k8s":
		return envOrDefault("PLTF_PROVIDER_K8S_VERSION", K8sProviderVersion)
	case "helm":
		return envOrDefault("PLTF_PROVIDER_HELM_VERSION", HelmProviderVersion)
	case "kustomize", "kustomization":
		return envOrDefault("PLTF_PROVIDER_KUSTOMIZE_VERSION", KustomizeProviderVersion)
	default:
		return ""
	}
}

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

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
