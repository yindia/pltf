package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	autoValFile string
	autoValEnv  string
	autoValScan bool
	autoValMods string
)

// validateCmd auto-detects Environment vs Service and validates accordingly.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Args:  cobra.NoArgs,
	Short: "Validate an Environment or Service spec (auto-detects kind)",
	Long: `Parse a YAML spec, detect Environment vs Service, and run structural validation.
Optionally assert that a specific environment key exists in both the environment file
and the service envRef (for services). Lint suggestions are run alongside validation.`,
	Example: `  pltf validate -f env.yaml
  pltf validate -f service.yaml -e dev`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		autoValFile = defaultString(autoValFile, "env.yaml")
		autoValFile = cleanOptionalPath(autoValFile)
		autoValEnv = strings.TrimSpace(autoValEnv)
		return ensureFile(autoValFile, "spec file")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if autoValScan {
			return autoValidateWithScan(os.Stdout, autoValFile, autoValEnv, autoValMods)
		}
		return autoValidate(autoValFile, autoValEnv)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&autoValFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	validateCmd.Flags().StringVarP(&autoValEnv, "env", "e", "", "Environment key to assert exists (dev, prod, etc.)")
	validateCmd.Flags().BoolVar(&autoValScan, "scan", false, "Run tfsec security scan against generated Terraform")
	validateCmd.Flags().StringVarP(&autoValMods, "modules", "m", "", "Override modules root; defaults to embedded modules")
}
