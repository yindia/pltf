package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// =====================
// shared writers
// =====================

func writeVersionsTF(
	outDir string,
	backendBucket string,
	backendKey string,
	backendRegion string,
	provider string,
	backendType string,
	locals map[string]interface{},
	needsK8s bool,
	needsHelm bool,
	container string,
	resourceGroup string,
	backendProfile string,
) error {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	tfBlock := body.AppendNewBlock("terraform", nil)
	tfBody := tfBlock.Body()
	tfBody.SetAttributeValue("required_version", cty.StringVal(RequiredTfVersion))

	// required_providers block
	rpBlock := tfBody.AppendNewBlock("required_providers", nil)
	rpBody := rpBlock.Body()

	switch provider {
	case "aws", "":
		rpBody.SetAttributeValue("aws", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/aws"),
			"version": cty.StringVal(AWSProviderVersion),
		}))
		if needsK8s {
			rpBody.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/kubernetes"),
				"version": cty.StringVal(K8sProviderVersion),
			}))
		}
		if needsHelm {
			rpBody.SetAttributeValue("helm", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/helm"),
				"version": cty.StringVal(HelmProviderVersion),
			}))
		}
	case "azure", "azurerm":
		rpBody.SetAttributeValue("azurerm", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/azurerm"),
			"version": cty.StringVal(AzureProviderVersion),
		}))
		if needsK8s {
			rpBody.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/kubernetes"),
				"version": cty.StringVal(K8sProviderVersion),
			}))
		}
		if needsHelm {
			rpBody.SetAttributeValue("helm", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/helm"),
				"version": cty.StringVal(HelmProviderVersion),
			}))
		}
	case "gcp", "google":
		rpBody.SetAttributeValue("google", cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/google"),
			"version": cty.StringVal(GCPProviderVersion),
		}))
		if needsK8s {
			rpBody.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/kubernetes"),
				"version": cty.StringVal(K8sProviderVersion),
			}))
		}
		if needsHelm {
			rpBody.SetAttributeValue("helm", cty.ObjectVal(map[string]cty.Value{
				"source":  cty.StringVal("hashicorp/helm"),
				"version": cty.StringVal(HelmProviderVersion),
			}))
		}
	default:
		return fmt.Errorf("unsupported provider %q in writeVersionsTF", provider)
	}

	// backend block
	switch backendType {
	case "aws", "s3", "":
		backendBlock := tfBody.AppendNewBlock("backend", []string{"s3"})
		bb := backendBlock.Body()
		bb.SetAttributeValue("bucket", cty.StringVal(backendBucket))
		bb.SetAttributeValue("key", cty.StringVal(backendKey))
		bb.SetAttributeValue("region", cty.StringVal(backendRegion))
		// bb.SetAttributeValue("use_lockfile", cty.BoolVal(true))
		if strings.TrimSpace(backendProfile) != "" {
			bb.SetAttributeValue("profile", cty.StringVal(backendProfile))
		}
	case "gcp", "google", "gcs":
		backendBlock := tfBody.AppendNewBlock("backend", []string{"gcs"})
		bb := backendBlock.Body()
		bb.SetAttributeValue("bucket", cty.StringVal(backendBucket))
		// gcs backend uses "prefix" instead of "key"
		bb.SetAttributeValue("prefix", cty.StringVal(backendKey))
	case "azure", "azurerm":
		if backendBucket == "" {
			return fmt.Errorf("backend.bucket (storage account name) is required for azure")
		}
		if container == "" {
			container = "tfstate"
		}
		backendBlock := tfBody.AppendNewBlock("backend", []string{"azurerm"})
		bb := backendBlock.Body()
		bb.SetAttributeValue("storage_account_name", cty.StringVal(backendBucket))
		bb.SetAttributeValue("container_name", cty.StringVal(container))
		bb.SetAttributeValue("key", cty.StringVal(backendKey))
		if resourceGroup != "" {
			bb.SetAttributeValue("resource_group_name", cty.StringVal(resourceGroup))
		}
	default:
		// already handled above, but keep for safety
		return fmt.Errorf("unsupported provider %q for backend", provider)
	}

	body.AppendNewline()

	// locals
	localsBlock := body.AppendNewBlock("locals", nil)
	localsBody := localsBlock.Body()
	for _, k := range sortedKeysInterfaceMap(locals) {
		ctyVal, err := toCtyValue(locals[k])
		if err != nil {
			return fmt.Errorf("cannot convert local %s to cty: %w", k, err)
		}
		localsBody.SetAttributeValue(k, ctyVal)
	}

	return os.WriteFile(filepath.Join(outDir, "versions.tf"), file.Bytes(), 0o644)
}

func writeProvidersTF(
	outDir string,
	provider string,
	region string,
	account string,
	needsK8s bool,
	needsHelm bool,
	cluster *clusterRef,
) error {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	if (needsK8s || needsHelm) && cluster == nil {
		return fmt.Errorf("kubernetes or helm provider requested but no cluster module found")
	}

	// Core provider
	switch provider {
	case "aws", "":
		provBlock := body.AppendNewBlock("provider", []string{"aws"})
		provBody := provBlock.Body()
		provBody.SetAttributeValue("region", cty.StringVal(region))
		dt := provBody.AppendNewBlock("default_tags", nil)
		tags := dt.Body()
		tags.SetAttributeRaw("tags", defaultTagsTokens())
	case "gcp", "google":
		provBlock := body.AppendNewBlock("provider", []string{"google"})
		provBody := provBlock.Body()
		// assuming envEntry.Account is GCP project ID
		provBody.SetAttributeValue("project", cty.StringVal(account))
		provBody.SetAttributeValue("region", cty.StringVal(region))
	case "azure", "azurerm":
		provBlock := body.AppendNewBlock("provider", []string{"azurerm"})
		provBody := provBlock.Body()
		provBody.SetAttributeValue("subscription_id", cty.StringVal(account))
		provBody.SetAttributeValue("features", cty.ObjectVal(map[string]cty.Value{}))
	default:
		return fmt.Errorf("unsupported provider %q in writeProvidersTF", provider)
	}

	if (needsK8s || needsHelm) && cluster != nil && cluster.auth != nil {
		body.AppendNewline()
		writeAuthData(body, cluster.auth)
	}

	if needsK8s && cluster != nil {
		body.AppendNewline()
		k8s := body.AppendNewBlock("provider", []string{"kubernetes"})
		k8sBody := k8s.Body()
		k8sBody.SetAttributeRaw("host", hclwrite.TokensForTraversal(cluster.host))
		k8sBody.SetAttributeRaw("cluster_ca_certificate", base64DecodeTokens(cluster.caData))
		k8sBody.SetAttributeRaw("token", hclwrite.TokensForTraversal(cluster.token))
	}

	if needsHelm && cluster != nil {
		body.AppendNewline()
		helm := body.AppendNewBlock("provider", []string{"helm"})
		helmBody := helm.Body()
		helmBody.SetAttributeRaw("kubernetes", helmKubeConfigTokens(cluster))
	}

	return os.WriteFile(filepath.Join(outDir, "providers.tf"), file.Bytes(), 0o644)
}

func writeRemoteStateTF(outDir string, backendType string, bucket string, key string, region string, container string, resourceGroup string, backendProfile string) error {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	rsBlock := body.AppendNewBlock("data", []string{"terraform_remote_state", "env"})
	rsBody := rsBlock.Body()

	switch backendType {
	case "aws", "s3", "":
		cfg := map[string]cty.Value{
			"bucket": cty.StringVal(bucket),
			"key":    cty.StringVal(key),
			"region": cty.StringVal(region),
		}
		if strings.TrimSpace(backendProfile) != "" {
			cfg["profile"] = cty.StringVal(backendProfile)
		}
		rsBody.SetAttributeValue("backend", cty.StringVal("s3"))
		rsBody.SetAttributeValue("config", cty.ObjectVal(cfg))
	case "gcp", "google", "gcs":
		rsBody.SetAttributeValue("backend", cty.StringVal("gcs"))
		rsBody.SetAttributeValue("config", cty.ObjectVal(map[string]cty.Value{
			"bucket": cty.StringVal(bucket),
			"prefix": cty.StringVal(key),
		}))
	case "azure", "azurerm":
		if bucket == "" {
			return fmt.Errorf("backend.bucket (storage account name) is required for azure")
		}
		if container == "" {
			container = "tfstate"
		}
		cfg := map[string]cty.Value{
			"storage_account_name": cty.StringVal(bucket),
			"container_name":       cty.StringVal(container),
			"key":                  cty.StringVal(key),
		}
		if resourceGroup != "" {
			cfg["resource_group_name"] = cty.StringVal(resourceGroup)
		}
		rsBody.SetAttributeValue("backend", cty.StringVal("azurerm"))
		rsBody.SetAttributeValue("config", cty.ObjectVal(cfg))
	default:
		return fmt.Errorf("unsupported backend %q in writeRemoteStateTF", backendType)
	}

	return os.WriteFile(filepath.Join(outDir, "state.tf"), file.Bytes(), 0o644)
}

func writeSecretsTF(outDir string, secretNames map[string]bool) error {
	if len(secretNames) == 0 {
		return nil
	}

	file := hclwrite.NewEmptyFile()
	body := file.Body()

	// deterministic order
	names := make([]string, 0, len(secretNames))
	for name := range secretNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		block := body.AppendNewBlock("variable", []string{name})
		b := block.Body()
		// we don't set type: defaults to any, that's OK.
		b.SetAttributeValue("sensitive", cty.BoolVal(true))
		body.AppendNewline()
	}

	return os.WriteFile(filepath.Join(outDir, "secrets.tf"), file.Bytes(), 0o644)
}

// module.<moduleID>.<outputName>
func setAttrModuleOutputRef(body *hclwrite.Body, name, moduleID, outputName string) {
	tokens := hclwrite.Tokens{
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("module"),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(moduleID),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(outputName),
		},
	}
	body.SetAttributeRaw(name, tokens)
}

// data.terraform_remote_state.env.outputs.<outputName>
func setAttrParentOutputRef(body *hclwrite.Body, name, outputName string) {
	tokens := hclwrite.Tokens{
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("data"),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("terraform_remote_state"),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("env"),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("outputs"),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(outputName),
		},
	}
	body.SetAttributeRaw(name, tokens)
}

func defaultTagsTokens() hclwrite.Tokens {
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
