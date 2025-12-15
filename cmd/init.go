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

	"pltf/pkg/config"
)

var (
	moduleInitPath      string
	moduleInitName      string
	moduleInitType      string
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
		root, err := resolveModulesRoot(moduleListRoot)
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
		root, err := resolveModulesRoot(moduleListRoot)
		if err != nil {
			return err
		}
		metas, err := scanModules(root)
		if err != nil {
			return err
		}
		mod, ok := metas[args[0]]
		if !ok {
			return fmt.Errorf("module %q not found under %s", args[0], root)
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
	Capabilities config.Capabilities `yaml:"capabilities,omitempty"`
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
Provider defaults to aws and version to 1.0.0.`,
	Example: `  # Generate module.yaml inside ./modules/aws_eks
  pltf module init --path ./modules/aws_eks

  # Write to a custom location and override name/type
  pltf module init --path ./modules/db --name postgres --type aws_postgres --out ./modules/db/module.yaml`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		moduleInitPath = defaultString(moduleInitPath, ".")
		moduleInitPath = cleanOptionalPath(moduleInitPath)
		moduleInitOut = cleanOptionalPath(moduleInitOut)
		if err := ensureDir(moduleInitPath, "module path"); err != nil {
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

		meta := buildModuleMetadata(abs, tfMod)
		yamlMeta := buildModuleMetadataYAML(meta, tfMod)

		out, err := yaml.Marshal(yamlMeta)
		if err != nil {
			return err
		}

		outFile := moduleInitOut
		if outFile == "" {
			outFile = filepath.Join(abs, "module.yaml")
		}
		if err := backupIfExists(outFile, moduleInitOverwrite); err != nil {
			return err
		}

		return os.WriteFile(outFile, out, 0o644)
	},
}

func buildModuleMetadata(abs string, tfMod *tfconfig.Module) *config.ModuleMetadata {
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

	return &config.ModuleMetadata{
		Name:        name,
		Type:        modType,
		Provider:    defaultModuleProvider,
		Version:     defaultModuleVersion,
		Description: moduleInitDesc,
		Capabilities: config.Capabilities{
			Provides: provides,
			Accepts:  accepts,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}
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
		Capabilities: meta.Capabilities,
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

// -------------------------------------------------------

func init() {
	rootCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleInitCmd)
	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleGetCmd)

	moduleInitCmd.Flags().StringVar(&moduleInitPath, "path", ".", "Directory containing the Terraform module to inspect; defaults to current directory")
	moduleInitCmd.Flags().StringVar(&moduleInitName, "name", "", "Module name to write into module.yaml (defaults to directory name)")
	moduleInitCmd.Flags().StringVar(&moduleInitType, "type", "", "Logical module type; defaults to the module name when omitted")
	moduleInitCmd.Flags().StringVar(&moduleInitDesc, "description", "", "Human-readable description for the module; optional")
	moduleInitCmd.Flags().StringVar(&moduleInitOut, "out", "", "Output path for module.yaml (defaults to <path>/module.yaml)")
	moduleInitCmd.Flags().BoolVar(&moduleInitOverwrite, "force", false, "Overwrite an existing module.yaml (backs up to module.yaml.bak-<timestamp> when absent)")

	moduleListCmd.Flags().StringVarP(&moduleListRoot, "modules", "m", "", "Modules root; defaults to embedded modules")
	moduleListCmd.Flags().StringVarP(&moduleListOut, "output", "o", "table", "Output format: table|json|yaml")

	moduleGetCmd.Flags().StringVarP(&moduleListRoot, "modules", "m", "", "Modules root; defaults to embedded modules")
	moduleGetCmd.Flags().StringVarP(&moduleListOut, "output", "o", "table", "Output format: table|json|yaml")
}
