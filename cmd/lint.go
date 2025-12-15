package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/spf13/cobra"

	"pltf/pkg/config"
)

var (
	lintFile string
	lintEnv  string
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Args:  cobra.NoArgs,
	Short: "Lint an Environment or Service spec and suggest fixes",
	Long:  "Perform structural validation plus lightweight linting (unused variables, missing labels) for Environment or Service specs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLintWithOutput(cmd.OutOrStdout(), lintFile, lintEnv)
	},
}

func runLint(file, env string) error {
	return runLintWithOutput(os.Stdout, file, env)
}

func runLintWithOutput(out io.Writer, file, env string) error {
	file = defaultString(file, "env.yaml")
	if err := ensureFile(file, "spec file"); err != nil {
		return err
	}

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		suggestions := lintEnvironment(envCfg, envName)
		printLintSuggestions(out, kind, envName, suggestions)
	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		suggestions := lintService(svcCfg, envCfg, envName)
		printLintSuggestions(out, kind, envName, suggestions)
	default:
		return fmt.Errorf("unknown kind %q", kind)
	}
	return nil
}

func lintEnvironment(cfg *config.EnvironmentConfig, envName string) []string {
	var suggestions []string
	envEntry := cfg.Environments[envName]
	if len(cfg.Metadata.Labels) == 0 {
		suggestions = append(suggestions, "Add metadata.labels for tagging (team, cost_center, etc.)")
	}

	unused := findUnusedVars(envEntry.Variables, nil, cfg.Modules)
	for _, v := range unused {
		suggestions = append(suggestions, fmt.Sprintf("Variable %q is defined but not referenced in modules for env %q", v, envName))
	}
	return suggestions
}

func lintService(svc *config.ServiceConfig, envCfg *config.EnvironmentConfig, envName string) []string {
	var suggestions []string
	if len(svc.Metadata.Labels) == 0 && len(envCfg.Metadata.Labels) == 0 {
		suggestions = append(suggestions, "Add labels on service or environment for tagging (team, cost_center, etc.)")
	}

	envRef := svc.Metadata.EnvRef[envName]
	unused := findUnusedVars(envCfg.Environments[envName].Variables, envRef.Variables, svc.Modules)
	for _, v := range unused {
		suggestions = append(suggestions, fmt.Sprintf("Variable %q is defined but not referenced in service modules for env %q", v, envName))
	}
	return suggestions
}

var varPattern = regexp.MustCompile(`var\\.([A-Za-z0-9_]+)`)

func findUnusedVars(envVars map[string]string, svcVars map[string]string, mods []config.Module) []string {
	defined := map[string]struct{}{}
	for k := range envVars {
		defined[k] = struct{}{}
	}
	for k := range svcVars {
		defined[k] = struct{}{}
	}

	used := map[string]struct{}{}
	for _, m := range mods {
		for _, v := range m.Inputs {
			markVarsInValue(v, used)
		}
	}

	var unused []string
	for k := range defined {
		if _, ok := used[k]; !ok {
			unused = append(unused, k)
		}
	}
	sort.Strings(unused)
	return unused
}

func markVarsInValue(val interface{}, used map[string]struct{}) {
	switch t := val.(type) {
	case string:
		for _, m := range varPattern.FindAllStringSubmatch(t, -1) {
			if len(m) > 1 {
				used[m[1]] = struct{}{}
			}
		}
	case []interface{}:
		for _, v := range t {
			markVarsInValue(v, used)
		}
	case map[string]interface{}:
		for _, v := range t {
			markVarsInValue(v, used)
		}
	}
}

func printLintSuggestions(out io.Writer, kind, env string, suggestions []string) {
	fmt.Fprintf(out, "%s lint for env %s:\n", kind, env)
	if len(suggestions) == 0 {
		fmt.Fprintln(out, "  No lint issues found.")
		return
	}
	for _, s := range suggestions {
		fmt.Fprintf(out, "  - %s\n", s)
	}
}

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().StringVarP(&lintFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	lintCmd.Flags().StringVarP(&lintEnv, "env", "e", "", "Environment key to lint (dev, prod, etc.)")
}
