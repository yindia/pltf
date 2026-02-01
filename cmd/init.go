package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pltf/pkg/clihelper"
	"pltf/pkg/config"
)

var (
	moduleInitPath      string
	moduleInitName      string
	moduleInitType      string
	moduleInitProvider  string
	moduleInitDesc      string
	moduleInitOut       string
	moduleInitOverwrite bool
	moduleListRoot      string
	moduleListOut       string
)

// module list
var moduleListCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "List available modules (reads module.yaml inventory)",
	Long:  "Scan a modules root for module.yaml files and list the module types, providers, and descriptions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := clihelper.ResolveModulesRoot(moduleListRoot)
		if err != nil {
			return err
		}
		metas, err := scanModules(root)
		if err != nil {
			return err
		}
		return printModules(metas, moduleListOut)
	},
}

// module get
var moduleGetCmd = &cobra.Command{
	Use:   "get <module_type>",
	Args:  cobra.ExactArgs(1),
	Short: "Show details for a module (inputs/outputs)",
	Long:  "Display module metadata from module.yaml including provider, version, inputs, and outputs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, label, err := clihelper.ResolveModulesRootWithLabel(moduleListRoot)
		if err != nil {
			return err
		}
		metas, err := scanModules(root)
		if err != nil {
			return err
		}
		mod, ok := metas[args[0]]
		if !ok {
			suggestions := suggestModuleTypes(args[0], metas)
			if len(suggestions) > 0 {
				return fmt.Errorf("module %q not found in %s (did you mean: %s)", args[0], label, strings.Join(suggestions, ", "))
			}
			return fmt.Errorf("module %q not found in %s (run \"pltf module list\" to see available modules)", args[0], label)
		}
		return printModuleDetail(mod, moduleListOut)
	},
}

const (
	defaultModuleProvider = "aws"
	defaultModuleVersion  = "1.0.0"
)

// Parent command: pltf module
var moduleCmd = &cobra.Command{
	Use:   "module",
	Args:  cobra.NoArgs,
	Short: "Helpers for working with Terraform modules",
	Long:  "Inspect Terraform modules and scaffold module.yaml metadata files used by env/service generation and module discovery.",
}

// -------------------------------------------------------
// YAML structs
// -------------------------------------------------------

type inputSpecYAML struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Capability  string      `yaml:"capability,omitempty"`
	Required    bool        `yaml:"required"` // <-- always output
	Default     interface{} `yaml:"default,omitempty"`
	HasDefault  bool        `yaml:"-"`
}

// Custom YAML so null is kept instead of omitted
func (i inputSpecYAML) MarshalYAML() (interface{}, error) {
	m := map[string]interface{}{
		"name":     i.Name,
		"required": i.Required,
	}

	if i.Type != "" {
		m["type"] = i.Type
	}
	if i.Description != "" {
		m["description"] = i.Description
	}
	if i.Capability != "" {
		m["capability"] = i.Capability
	}

	if i.HasDefault { // default exists even if null
		m["default"] = i.Default
	}

	return m, nil
}

type moduleMetadataYAML struct {
	Name         string              `yaml:"name"`
	Type         string              `yaml:"type"`
	Provider     string              `yaml:"provider"`
	Version      string              `yaml:"version"`
	Description  string              `yaml:"description,omitempty"`
	Cluster      bool                `yaml:"cluster,omitempty"`
	Capabilities config.Capabilities `yaml:"capabilities,omitempty"`
	Resources    []string            `yaml:"resources,omitempty"`
	DataSources  []string            `yaml:"data,omitempty"`
	Inputs       []inputSpecYAML     `yaml:"inputs,omitempty"`
	Outputs      []config.OutputSpec `yaml:"outputs,omitempty"`
}

// -------------------------------------------------------
// Command: module init
// -------------------------------------------------------

var moduleInitCmd = &cobra.Command{
	Use:   "init",
	Args:  cobra.NoArgs,
	Short: "Generate a module.yaml from an existing Terraform module",
	Long: `Scan a Terraform module directory, read variables/outputs, and write a module.yaml
descriptor. If module.yaml already exists at the destination it will be replaced.
Use flags to override metadata such as name, type, description, or output path.
Provider defaults to aws and version to 1.0.0 (override with --provider).`,
	Example: `  # Generate module.yaml inside ./modules/aws_eks
  pltf module init --path ./modules/aws_eks

  # Generate module.yaml for a GCP module
  pltf module init --path ./modules/gcp_gcs --provider gcp

  # Write to a custom location and override name/type
  pltf module init --path ./modules/db --name postgres --type aws_postgres --out ./modules/db/module.yaml`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		moduleInitPath = clihelper.DefaultString(moduleInitPath, ".")
		moduleInitPath = clihelper.CleanOptionalPath(moduleInitPath)
		moduleInitOut = clihelper.CleanOptionalPath(moduleInitOut)
		if err := clihelper.EnsureDir(moduleInitPath, "module path"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(moduleInitPath)
		if err != nil {
			return err
		}

		tfMod, diags := tfconfig.LoadModule(abs)
		if diags.HasErrors() {
			return fmt.Errorf("error loading module: %v", diags)
		}

		provider, err := normalizeModuleProvider(moduleInitProvider)
		if err != nil {
			return err
		}
		meta := buildModuleMetadata(abs, tfMod, provider)
		yamlMeta := buildModuleMetadataYAML(meta, tfMod)

		out, err := yaml.Marshal(yamlMeta)
		if err != nil {
			return err
		}

		outFile := moduleInitOut
		if outFile == "" {
			outFile = filepath.Join(abs, "module.yaml")
		}
		if err := clihelper.BackupIfExists(outFile, moduleInitOverwrite); err != nil {
			return err
		}

		return os.WriteFile(outFile, out, 0o644)
	},
}

func buildModuleMetadata(abs string, tfMod *tfconfig.Module, provider string) *config.ModuleMetadata {
	name := moduleInitName
	if name == "" {
		name = filepath.Base(abs)
	}
	modType := moduleInitType
	if modType == "" {
		modType = name
	}

	inputs, accepts := buildInputs(tfMod)
	outputs, provides := buildOutputs(tfMod)
	iamProvides := detectIamProvides(outputs)
	iamAccepts := detectIamAccepts(inputs)
	provides = dedupeStrings(append(provides, iamProvides...))
	accepts = dedupeStrings(append(accepts, iamAccepts...))
	resources, dataSources := buildResourceLists(tfMod)

	return &config.ModuleMetadata{
		Name:        name,
		Type:        modType,
		Provider:    provider,
		Version:     defaultModuleVersion,
		Description: moduleInitDesc,
		Cluster:     hasClusterOutputs(outputs),
		Capabilities: config.Capabilities{
			Provides: provides,
			Accepts:  accepts,
		},
		Resources:   resources,
		DataSources: dataSources,
		Inputs:      inputs,
		Outputs:     outputs,
	}
}

func normalizeModuleProvider(raw string) (string, error) {
	provider := strings.ToLower(strings.TrimSpace(raw))
	if provider == "" {
		return defaultModuleProvider, nil
	}
	if provider == "google" {
		provider = "gcp"
	}

	allowed := map[string]struct{}{
		"aws":     {},
		"gcp":     {},
		"azure":   {},
		"azurerm": {},
		"helm":    {},
	}
	if _, ok := allowed[provider]; !ok {
		return "", fmt.Errorf("invalid provider %q (expected aws, gcp, azure, azurerm, helm)", raw)
	}
	return provider, nil
}

func buildInputs(tfMod *tfconfig.Module) ([]config.InputSpec, []string) {
	var keys []string
	for k := range tfMod.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var (
		result  []config.InputSpec
		accepts []string
	)
	for _, k := range keys {
		v := tfMod.Variables[k]
		cap := inferCapability(k)
		result = append(result, config.InputSpec{
			Name:        k,
			Type:        strings.TrimSpace(v.Type),
			Description: strings.TrimSpace(v.Description),
			Required:    v.Required,
			Default:     v.Default,
			Capability:  cap,
		})
		if cap != "" {
			accepts = append(accepts, cap)
		}
	}
	return result, dedupeStrings(accepts)
}

func buildOutputs(tfMod *tfconfig.Module) ([]config.OutputSpec, []string) {
	var keys []string
	for k := range tfMod.Outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var (
		result   []config.OutputSpec
		provides []string
	)
	for _, k := range keys {
		v := tfMod.Outputs[k]
		cap := inferCapability(k)
		result = append(result, config.OutputSpec{
			Name:        k,
			Type:        "string",
			Description: strings.TrimSpace(v.Description),
			Capability:  cap,
		})
		if cap != "" {
			provides = append(provides, cap)
		}
	}
	return result, dedupeStrings(provides)
}

func buildResourceLists(tfMod *tfconfig.Module) ([]string, []string) {
	resources := map[string]struct{}{}
	for _, res := range tfMod.ManagedResources {
		resources[res.MapKey()] = struct{}{}
	}
	dataSources := map[string]struct{}{}
	for _, res := range tfMod.DataResources {
		dataSources[res.MapKey()] = struct{}{}
	}

	resList := make([]string, 0, len(resources))
	for name := range resources {
		resList = append(resList, name)
	}
	sort.Strings(resList)

	dataList := make([]string, 0, len(dataSources))
	for name := range dataSources {
		dataList = append(dataList, name)
	}
	sort.Strings(dataList)

	return resList, dataList
}

func detectIamProvides(outputs []config.OutputSpec) []string {
	var caps []string
	for _, out := range outputs {
		switch out.Name {
		case "role_arn":
			caps = append(caps, "iam.role")
		case "user_arn":
			caps = append(caps, "iam.user")
		case "policy_arn":
			caps = append(caps, "iam.policy")
		}
	}
	return dedupeStrings(caps)
}

func detectIamAccepts(inputs []config.InputSpec) []string {
	var caps []string
	for _, in := range inputs {
		switch in.Name {
		case "iam_policy", "policy_arns":
			caps = append(caps, "iam.policy")
		case "kubernetes_trusts":
			caps = append(caps, "iam.trusts")
		}
	}
	return dedupeStrings(caps)
}

func buildModuleMetadataYAML(meta *config.ModuleMetadata, tfMod *tfconfig.Module) moduleMetadataYAML {
	inputs := []inputSpecYAML{}

	for _, in := range meta.Inputs {
		v := tfMod.Variables[in.Name]
		hasDefault := !v.Required // means default is present (even null)

		inputs = append(inputs, inputSpecYAML{
			Name:        in.Name,
			Type:        in.Type,
			Description: in.Description,
			Required:    in.Required, // always written
			Default:     in.Default,
			HasDefault:  hasDefault,
		})
	}

	return moduleMetadataYAML{
		Name:         meta.Name,
		Type:         meta.Type,
		Provider:     meta.Provider,
		Version:      meta.Version,
		Description:  meta.Description,
		Cluster:      meta.Cluster,
		Capabilities: meta.Capabilities,
		Resources:    meta.Resources,
		DataSources:  meta.DataSources,
		Inputs:       inputs,
		Outputs:      meta.Outputs,
	}
}

func inferCapability(name string) string {
	l := strings.ToLower(name)
	keywords := []string{
		"password",
		"secret",
		"token",
		"private_key",
		"client_secret",
		"api_key",
	}
	for _, kw := range keywords {
		if strings.Contains(l, kw) {
			return "secret"
		}
	}
	return ""
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func suggestModuleTypes(query string, metas map[string]*config.ModuleMetadata) []string {
	query = strings.ToLower(query)
	var startsWith []string
	var contains []string
	for name := range metas {
		lower := strings.ToLower(name)
		switch {
		case strings.HasPrefix(lower, query):
			startsWith = append(startsWith, name)
		case strings.Contains(lower, query):
			contains = append(contains, name)
		}
	}
	sort.Strings(startsWith)
	sort.Strings(contains)
	out := append(startsWith, contains...)
	if len(out) > 5 {
		return out[:5]
	}
	return out
}

func hasClusterOutputs(outputs []config.OutputSpec) bool {
	required := map[string]struct{}{
		"k8s_endpoint":     {},
		"k8s_ca_data":      {},
		"k8s_cluster_name": {},
		"plt_cluster_type": {},
	}
	for _, out := range outputs {
		delete(required, out.Name)
	}
	return len(required) == 0
}

// -------------------------------------------------------

func init() {
	rootCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleInitCmd)
	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleGetCmd)

	moduleInitCmd.Flags().StringVar(&moduleInitPath, "path", ".", "Directory containing the Terraform module to inspect; defaults to current directory")
	moduleInitCmd.Flags().StringVar(&moduleInitName, "name", "", "Module name to write into module.yaml (defaults to directory name)")
	moduleInitCmd.Flags().StringVar(&moduleInitType, "type", "", "Logical module type; defaults to the module name when omitted")
	moduleInitCmd.Flags().StringVar(&moduleInitProvider, "provider", "", "Provider for module.yaml (aws|gcp|azure|azurerm|helm; defaults to aws)")
	moduleInitCmd.Flags().StringVar(&moduleInitDesc, "description", "", "Human-readable description for the module; optional")
	moduleInitCmd.Flags().StringVar(&moduleInitOut, "out", "", "Output path for module.yaml (defaults to <path>/module.yaml)")
	moduleInitCmd.Flags().BoolVar(&moduleInitOverwrite, "force", false, "Overwrite an existing module.yaml (backs up to module.yaml.bak-<timestamp> when absent)")

	moduleListCmd.Flags().StringVarP(&moduleListRoot, "modules", "m", "", "Modules root; defaults to embedded modules")
	moduleListCmd.Flags().StringVarP(&moduleListOut, "output", "o", "table", "Output format: table|json|yaml")

	moduleGetCmd.Flags().StringVarP(&moduleListRoot, "modules", "m", "", "Modules root; defaults to embedded modules")
	moduleGetCmd.Flags().StringVarP(&moduleListOut, "output", "o", "table", "Output format: table|json|yaml")
}
