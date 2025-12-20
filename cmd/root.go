package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global flags (if you want them later)
var (
	flagVerbose   bool
	flagTelemetry bool
)

var RootCmd  *cobra.Command

func init(){
	RootCmd = rootCmd
}
// rootCmd is the base command for the CLI.
// Subcommands like `env` and `service` are added to this in their own init() funcs.
var rootCmd = &cobra.Command{
	Use:   "pltf",
	Args:  cobra.NoArgs,
	Short: "Platform toolkit for validating and generating Terraform stacks",
	Long: `pltf consumes YAML definitions for platform "environments" and application "services"
and produces validated Terraform. Environment files capture accounts, regions, shared
modules, and defaults; Service files reference an environment and declare app-specific
modules. pltf checks structure and wiring, then renders Terraform with remote state,
providers, locals, secrets, and module connections. Modules are discovered from a
modules root where each module type exposes a module.yaml (generated via pltf module init).`,
	Example: `# Validate configs
pltf env validate --file env.yaml
pltf service validate --file service.yaml

# Generate Terraform for dev
pltf env generate --file env.yaml --env dev
pltf service generate --file service.yaml --env dev

# Scaffold module metadata for an existing TF module
pltf module init --path ./modules/aws_eks`,
	// Run without subcommand: just show help
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is called from main.main(). It runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error, but we ensure a non-zero exit code.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags for all subcommands
	defaultTelemetry := false
	if p := loadProfile(); p != nil {
		defaultTelemetry = p.Telemetry
	}
	rootCmd.PersistentFlags().BoolVarP(
		&flagVerbose,
		"verbose",
		"V",
		false,
		"Enable verbose logging",
	)
	rootCmd.PersistentFlags().BoolVar(
		&flagTelemetry,
		"telemetry",
		defaultTelemetry,
		"Enable anonymous telemetry (usage metrics). Currently a stub/no-op unless enabled.",
	)
}
