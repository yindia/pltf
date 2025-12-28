package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pltf/modules"
	"pltf/pkg/config"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestGeneratorAutoWiresLocalsAndOutputs(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {
				Account: "111111111111",
				Region:  "us-east-1",
				Variables: map[string]string{
					"cluster_name": "dev-cluster",
				},
			},
		},
		Modules: []config.Module{
			{ID: "base", Type: "aws_base"},
			{ID: "eks", Type: "aws_eks"},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", outDir, "", map[string]string{
		"enable_metrics": "true",
		"cluster_name":   "dev-cluster",
	})
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}

	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	assertFiles(t, outDir, "versions.tf", "providers.tf", filepath.Join("base.tf"), filepath.Join("eks.tf"))
}

func TestGeneratorServiceUsesParentOutputs(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "base", Type: "aws_base"},
		},
	}

	svcCfg := &config.ServiceConfig{
		Metadata: config.ServiceMetadata{
			Name: "payments",
			EnvRef: map[string]config.ServiceEnvRefEntry{
				"dev": {},
			},
		},
		Modules: []config.Module{
			{
				ID:   "eks",
				Type: "aws_eks",
				Inputs: map[string]interface{}{
					"vpc_id": "parent.vpc_id",
					// satisfy required input
					"cluster_name":   "svc-dev",
					"enable_metrics": true,
				},
			},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, svcCfg, modRoot, "", "dev", outDir, "", nil)
	if err != nil {
		t.Fatalf("NewGenerator(service) error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate(service) error: %v", err)
	}

	// Ensure service module file generated and modules copied.
	assertFiles(t, outDir, "versions.tf", "providers.tf", "state.tf", filepath.Join("eks.tf"))
}

func TestGeneratorServiceSkipsEnvDependsOn(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "base", Type: "aws_base"},
		},
	}

	svcCfg := &config.ServiceConfig{
		Metadata: config.ServiceMetadata{
			Name: "payments",
			EnvRef: map[string]config.ServiceEnvRefEntry{
				"dev": {},
			},
		},
		Modules: []config.Module{
			{
				ID:   "eks",
				Type: "aws_eks",
				Inputs: map[string]interface{}{
					"vpc_id": "module.base.vpc_id",
					// satisfy required input
					"cluster_name":   "svc-dev",
					"enable_metrics": true,
				},
			},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, svcCfg, modRoot, "", "dev", outDir, "", nil)
	if err != nil {
		t.Fatalf("NewGenerator(service) error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate(service) error: %v", err)
	}

	eksTf := filepath.Join(outDir, "eks.tf")
	data, err := os.ReadFile(eksTf)
	if err != nil {
		t.Fatalf("read generated eks.tf: %v", err)
	}
	if strings.Contains(string(data), "depends_on") {
		t.Fatalf("expected no depends_on for env-only dependencies, got:\n%s", string(data))
	}
}

func TestGeneratorIgnoresEmptyCustomRoot(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "base", Type: "aws_base"},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", t.TempDir(), "", nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if g.customModulesRoot != "" {
		t.Fatalf("expected empty customModulesRoot when not provided, got %q", g.customModulesRoot)
	}
	for typ, root := range g.moduleRootByType {
		if filepath.Clean(root) == "." {
			t.Fatalf("module type %s unexpectedly resolved to current dir root %q", typ, root)
		}
	}
}

func TestGeneratorRejectsUnsafeOutDir(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "base", Type: "aws_base"},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", ".", "", nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if err := g.Generate(); err == nil {
		t.Fatalf("expected Generate to fail for unsafe outDir")
	}
}

func TestReplaceIntrinsicPlaceholdersInValueRecurses(t *testing.T) {
	g := &Generator{
		envName: "platform",
		envEntry: config.EnvironmentEntry{
			Account: "123456789012",
			Region:  "us-west-2",
		},
		isService: true,
		svcCfg: &config.ServiceConfig{
			Metadata: config.ServiceMetadata{Name: "payments"},
		},
	}

	input := map[string]interface{}{
		"list": []interface{}{
			"${account_id}",
			map[string]interface{}{
				"region": "${region}",
			},
			[]interface{}{"${layer_name}", "${env_name}", "${parent_name}"},
		},
		"map": map[string]interface{}{
			"project": "${project_id}",
		},
	}

	// Ensure input is not mutated.
	origFirst := input["list"].([]interface{})[0]

	got := g.replaceIntrinsicPlaceholdersInValue(input).(map[string]interface{})

	if origFirst != "${account_id}" {
		t.Fatalf("original input mutated: %v", origFirst)
	}

	list := got["list"].([]interface{})
	if list[0] != "123456789012" {
		t.Fatalf("account_id not replaced in list: %v", list[0])
	}
	innerMap := list[1].(map[string]interface{})
	if innerMap["region"] != "us-west-2" {
		t.Fatalf("region not replaced in object: %v", innerMap["region"])
	}
	nestedSlice := list[2].([]interface{})
	expectedLayer := "payments" // service name becomes layer for services
	if nestedSlice[0] != expectedLayer || nestedSlice[1] != "platform" || nestedSlice[2] != expectedLayer {
		t.Fatalf("layer/env/parent placeholders not replaced: %v", nestedSlice)
	}

	m := got["map"].(map[string]interface{})
	if m["project"] != "123456789012" {
		t.Fatalf("project_id not replaced in map: %v", m["project"])
	}
}

func TestStringToTokensHandlesCurlyRefs(t *testing.T) {
	g := &Generator{
		envName: "example-aws",
		envEntry: config.EnvironmentEntry{
			Account: "111111111111",
			Region:  "us-east-1",
		},
	}

	file := hclwrite.NewEmptyFile()
	body := file.Body()
	body.SetAttributeRaw("region_val", g.stringToTokens("{region}"))
	body.SetAttributeRaw("module_ref", g.stringToTokens("{module.dns.domain}"))
	body.SetAttributeRaw("double_curly", g.stringToTokens("{{region}}"))
	body.SetAttributeRaw("dollar_curly", g.stringToTokens("${region}"))
	body.SetAttributeRaw("dollar_double_curly", g.stringToTokens("${{region}}"))
	body.SetAttributeRaw("dollar_double_module", g.stringToTokens("${{module.dns.domain}}"))

	out := string(file.Bytes())
	if !strings.Contains(out, `region_val`) || strings.Contains(out, `{region}`) || !strings.Contains(out, `"us-east-1"`) {
		t.Fatalf("expected region placeholder to be replaced, got:\n%s", out)
	}
	if !strings.Contains(out, "module_ref") || strings.Contains(out, "{module.dns.domain}") || !strings.Contains(out, "module.dns.domain") {
		t.Fatalf("expected module traversal without quotes, got:\n%s", out)
	}
	if !strings.Contains(out, `double_curly`) || strings.Contains(out, "{{region}}") || !strings.Contains(out, `"us-east-1"`) {
		t.Fatalf("expected {{region}} placeholder to be replaced, got:\n%s", out)
	}
	if !strings.Contains(out, `dollar_curly`) || strings.Contains(out, "${region}") || !strings.Contains(out, `"us-east-1"`) {
		t.Fatalf("expected ${region} placeholder to be replaced, got:\n%s", out)
	}
	if !strings.Contains(out, `dollar_double_curly`) || strings.Contains(out, "${{region}}") || !strings.Contains(out, `"us-east-1"`) {
		t.Fatalf("expected ${{region}} placeholder to be replaced, got:\n%s", out)
	}
	if strings.Contains(out, "$${module.dns.domain") || strings.Contains(out, "${{module.dns.domain}}") {
		t.Fatalf("expected ${{module.dns.domain}} to render as traversal, got:\n%s", out)
	}
}

func TestMapKeysWithDotsAreQuoted(t *testing.T) {
	g := &Generator{}
	file := hclwrite.NewEmptyFile()
	body := file.Body()
	value := map[string]interface{}{
		"pass.txt":                             "secret",
		"kubernetes.io/ingress.class":          "nginx",
		"nginx.ingress.kubernetes.io/app-root": "/console",
		"simple":                               "ok",
	}
	if err := g.setAttribute(body, "values", value); err != nil {
		t.Fatalf("setAttribute error: %v", err)
	}
	out := string(file.Bytes())
	if !strings.Contains(out, `"pass.txt"`) || !strings.Contains(out, `= "secret"`) {
		t.Fatalf("expected pass.txt key to be quoted, got:\n%s", out)
	}
	if !strings.Contains(out, `"kubernetes.io/ingress.class"`) || !strings.Contains(out, `"nginx"`) {
		t.Fatalf("expected ingress key to be quoted, got:\n%s", out)
	}
	if !strings.Contains(out, `"nginx.ingress.kubernetes.io/app-root"`) || !strings.Contains(out, `"/console"`) {
		t.Fatalf("expected nginx ingress key to be quoted, got:\n%s", out)
	}
	if strings.Contains(out, "kubernetes.io/ingress.class =") || strings.Contains(out, "pass.txt =") {
		t.Fatalf("found unquoted dotted keys:\n%s", out)
	}
	if !strings.Contains(out, "simple") || !strings.Contains(out, `= "ok"`) {
		t.Fatalf("expected simple identifier to remain bare, got:\n%s", out)
	}
}

func TestOutputsFileIncludesModuleOutputs(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "bucket", Type: "aws_s3", Inputs: map[string]interface{}{
				"bucket_name": "example-dev",
			}},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", outDir, "", nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "outputs.tf"))
	if err != nil {
		t.Fatalf("outputs.tf missing: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, `output "bucket_arn"`) {
		t.Fatalf("expected bucket_arn output, got:\n%s", out)
	}
	if !strings.Contains(out, "module.bucket.bucket_arn") {
		t.Fatalf("expected module reference for bucket_arn, got:\n%s", out)
	}
}

func TestOutputsDeduplicateWithModulePrefix(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "topic", Type: "aws_sns"},
			{ID: "queue", Type: "aws_sqs"},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", outDir, "", nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "outputs.tf"))
	if err != nil {
		t.Fatalf("outputs.tf missing: %v", err)
	}
	out := string(data)
	if strings.Contains(out, `output "kms_arn"`) {
		t.Fatalf("expected no bare kms_arn output when duplicated, got:\n%s", out)
	}
	if !strings.Contains(out, `output "topic_kms_arn"`) || !strings.Contains(out, `output "queue_kms_arn"`) {
		t.Fatalf("expected module-prefixed kms_arn outputs, got:\n%s", out)
	}
}

func TestMaterializeFileInputsCopiesFile(t *testing.T) {
	specDir := t.TempDir()
	policyPath := filepath.Join(specDir, "policy.json")
	if err := os.WriteFile(policyPath, []byte(`{"Version":"2012-10-17","Statement":[]}`), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "policy", Type: "aws_iam_policy", Inputs: map[string]interface{}{
				"file": "./policy.json",
			}},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}
	outDir := t.TempDir()

	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", outDir, specDir, nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	copied := filepath.Join(outDir, "policy.json")
	data, err := os.ReadFile(copied)
	if err != nil {
		t.Fatalf("expected copied file %s: %v", copied, err)
	}
	if !strings.Contains(string(data), `"Statement"`) {
		t.Fatalf("copied file contents unexpected: %s", string(data))
	}

	tfData, err := os.ReadFile(filepath.Join(outDir, "policy.tf"))
	if err != nil {
		t.Fatalf("read generated policy.tf: %v", err)
	}
	if !strings.Contains(string(tfData), `policy.json`) {
		t.Fatalf("expected module input to point to copied file, got:\n%s", string(tfData))
	}
}

func TestCollectDepsFromNestedValues(t *testing.T) {
	envCfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "testorg",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Account: "111111111111", Region: "us-east-1"},
		},
		Modules: []config.Module{
			{ID: "bucket", Type: "aws_s3", Inputs: map[string]interface{}{
				"bucket_name": "example-dev",
			}},
			{ID: "eks", Type: "aws_eks", Inputs: map[string]interface{}{
				"cluster_name":                  "dev",
				"enable_metrics":                true,
				"kms_account_key_arn":           "arn:aws:kms:region:acct:key/123",
				"private_subnet_ids":            []interface{}{"subnet-1", "subnet-2"},
				"vpc_id":                        "vpc-123",
				"module_name":                   "eks",
				"layer_name":                    "example",
				"env_name":                      "example",
				"node_instance_type":            "t3.medium",
				"node_disk_size":                20,
				"spot_instances":                false,
				"control_plane_security_groups": []interface{}{},
				"node_launch_template":          map[string]interface{}{},
			}},
			{ID: "helm", Type: "helm_chart", Inputs: map[string]interface{}{
				"chart": "test",
				"values": map[string]interface{}{
					"nested": map[string]interface{}{
						"ref": "${module.bucket.bucket_arn}",
					},
				},
			}},
		},
	}

	modRoot, err := modules.Materialize()
	if err != nil {
		t.Fatalf("materialize embedded modules: %v", err)
	}

	outDir := t.TempDir()
	g, err := NewGenerator(envCfg, nil, modRoot, "", "dev", outDir, "", nil)
	if err != nil {
		t.Fatalf("NewGenerator error: %v", err)
	}
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	helmTf := filepath.Join(outDir, "helm.tf")
	data, err := os.ReadFile(helmTf)
	if err != nil {
		t.Fatalf("read generated helm.tf: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "depends_on") || !strings.Contains(out, "module.bucket") {
		t.Fatalf("expected depends_on to include bucket module, got:\n%s", out)
	}
}

func assertFiles(t *testing.T, root string, files ...string) {
	t.Helper()
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Fatalf("expected generated file %s: %v", f, err)
		}
	}
}
