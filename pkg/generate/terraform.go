package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pltf/pkg/generate/cloud"
	"pltf/pkg/provider"
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
	providerType string,
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
	tfBody.SetAttributeValue("required_version", cty.StringVal(provider.RequiredTfVersion))

	p, err := cloud.New(providerType)
	if err != nil {
		return err
	}

	// required_providers block
	rpBlock := tfBody.AppendNewBlock("required_providers", nil)
	p.RequiredProviders(rpBlock.Body(), needsK8s, needsHelm)

	// backend block
	p.Backend(tfBody, backendBucket, backendKey, backendRegion, backendProfile, container, resourceGroup)

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
	providerType string,
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

	p, err := cloud.New(providerType)
	if err != nil {
		return err
	}

	// Core provider
	p.Provider(body, region, account)

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

