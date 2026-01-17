package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"pltf/pkg/clihelper"
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
	Short: "Generate workspace-ready Terraform from an Environment or Service spec",
	Long: `Read a YAML spec, detect Environment vs Service, and render a workspace-ready Terraform
root with variables.tf, secrets.tf, and a single <env>.tfvars file. Uses embedded modules by
default; can override modules root and output directory.`,
	Example: `  pltf generate -f env.yaml -e dev
  pltf generate -f service.yaml -e prod -m ./modules -o .pltf/my-env/my-svc/workspace`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		autoGenFile = clihelper.DefaultString(autoGenFile, "env.yaml")
		autoGenFile = clihelper.CleanOptionalPath(autoGenFile)
		autoGenEnv = strings.TrimSpace(autoGenEnv)
		autoGenModulesDir = clihelper.CleanOptionalPath(autoGenModulesDir)
		autoGenOut = clihelper.CleanOptionalPath(autoGenOut)

		if err := clihelper.EnsureFile(autoGenFile, "spec file"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(autoGenEnv) == "" {
			return clihelper.AutoGenerateAll(autoGenFile, autoGenModulesDir, autoGenOut, autoGenVars)
		}
		return clihelper.AutoGenerate(autoGenFile, autoGenEnv, autoGenModulesDir, autoGenOut, autoGenVars)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&autoGenFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	generateCmd.Flags().StringVarP(&autoGenEnv, "env", "e", "", "Environment key to render (dev, prod, etc.); required for both env and service specs")
	generateCmd.Flags().StringVarP(&autoGenModulesDir, "modules", "m", "", "Root directory containing module type folders with module.yaml metadata; defaults to embedded modules bundle")
	generateCmd.Flags().StringVarP(&autoGenOut, "out", "o", "", "Output directory for generated Terraform (defaults to .pltf/<env_name>/workspace or .pltf/<env_name>/<service>/workspace)")
	generateCmd.Flags().StringArrayVarP(&autoGenVars, "var", "v", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Can be repeated for multiple overrides.")
}
