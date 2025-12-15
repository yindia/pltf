package cmd

import (
	"github.com/spf13/cobra"
	"strings"
)

var (
	autoGenFile       string
	autoGenEnv        string
	autoGenOut        string
	autoGenModulesDir string
	autoGenVars       []string
)

// generateCmd auto-detects whether the file is an Environment or Service spec and generates accordingly.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Args:  cobra.NoArgs,
	Short: "Generate Terraform from an Environment or Service spec (auto-detects kind)",
	Long: `Read a YAML spec, detect Environment vs Service, and render Terraform with the proper
remote state, providers, locals, secrets, and module wiring. Uses embedded modules by
default; can override modules root and output directory.`,
	Example: `  pltf generate -f env.yaml -e dev
  pltf generate -f service.yaml -e prod -m ./modules -o .pltf/my-env/my-svc/env/prod`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		autoGenFile = defaultString(autoGenFile, "env.yaml")
		autoGenFile = cleanOptionalPath(autoGenFile)
		autoGenEnv = strings.TrimSpace(autoGenEnv)
		autoGenModulesDir = cleanOptionalPath(autoGenModulesDir)
		autoGenOut = cleanOptionalPath(autoGenOut)

		if err := ensureFile(autoGenFile, "spec file"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return autoGenerate(autoGenFile, autoGenEnv, autoGenModulesDir, autoGenOut, autoGenVars)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&autoGenFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	generateCmd.Flags().StringVarP(&autoGenEnv, "env", "e", "", "Environment key to render (dev, prod, etc.); required for both env and service specs")
	generateCmd.Flags().StringVarP(&autoGenModulesDir, "modules", "m", "", "Root directory containing module type folders with module.yaml metadata; defaults to embedded modules bundle")
	generateCmd.Flags().StringVarP(&autoGenOut, "out", "o", "", "Output directory for generated Terraform (defaults based on kind: .pltf/<env_name>/env/<env> or .pltf/<env_name>/<service>/env/<env>)")
	generateCmd.Flags().StringArrayVarP(&autoGenVars, "var", "v", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Can be repeated for multiple overrides.")
}
