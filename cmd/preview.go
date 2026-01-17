package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pltf/pkg/backend"
	"pltf/pkg/clihelper"
	"pltf/pkg/config"
)

var (
	previewFile string
	previewEnv  string
	previewOut  string
	previewMods string
)

// report/preview command: summarize what would be generated/applied.
var previewCmd = &cobra.Command{
	Use:   "preview",
	Args:  cobra.NoArgs,
	Short: "Preview a spec: provider, backend, modules, labels (no Terraform run)",
	Long:  "Parse a spec (Environment, Service, or Stack) and show a concise summary: provider, backend type, environment, labels, and modules to be rendered.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPreview(previewFile, previewEnv)
	},
}

func runPreview(file, env string) error {
	file = clihelper.DefaultString(file, "env.yaml")
	if err := clihelper.EnsureFile(file, "spec file"); err != nil {
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
		envName, err := clihelper.SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		embeddedRoot, customRoot, err := clihelper.ResolveModuleRoots(previewMods)
		if err != nil {
			return err
		}
		stacks, err := loadPreviewStacks(file, envCfg.Metadata.Stacks, embeddedRoot, customRoot)
		if err != nil {
			return err
		}
		mergedVars := mergeStringMap(mergeStackVariables(stacks), envCfg.Variables)
		outputs, err := buildPreviewOutputs(envCfg.Modules, embeddedRoot, customRoot)
		if err != nil {
			return err
		}
		printPreviewEnv(envCfg, envName, stacks, outputs, mergedVars)
	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := clihelper.SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		embeddedRoot, customRoot, err := clihelper.ResolveModuleRoots(previewMods)
		if err != nil {
			return err
		}
		stacks, err := loadPreviewStacks(file, svcCfg.Metadata.Stacks, embeddedRoot, customRoot)
		if err != nil {
			return err
		}
		mergedVars := mergeStringMap(mergeStackVariables(stacks), svcCfg.Variables)
		outputs, err := buildPreviewOutputs(svcCfg.Modules, embeddedRoot, customRoot)
		if err != nil {
			return err
		}
		printPreviewService(svcCfg, envCfg, envName, stacks, outputs, mergedVars)
	case "Stack":
		stackCfg, err := config.LoadStack(file)
		if err != nil {
			return err
		}
		embeddedRoot, customRoot, err := clihelper.ResolveModuleRoots(previewMods)
		if err != nil {
			return err
		}
		outputs, err := buildPreviewOutputs(stackCfg.Modules, embeddedRoot, customRoot)
		if err != nil {
			return err
		}
		printPreviewStack(stackCfg, outputs)
	default:
		return fmt.Errorf("unknown kind %q", kind)
	}
	return nil
}

func printPreviewEnv(cfg *config.EnvironmentConfig, envName string, stacks []previewStack, outputs []previewModuleOutputs, mergedVars map[string]string) {
	envEntry := cfg.Environments[envName]
	bk, _ := backend.Resolve(cfg.Metadata.Provider, cfg, envEntry)

	summary := map[string]interface{}{
		"kind":           "Environment",
		"name":           cfg.Metadata.Name,
		"env":            envName,
		"provider":       cfg.Metadata.Provider,
		"labels":         cfg.Metadata.Labels,
		"variables":      mergedVars,
		"backend_type":   bk.Type,
		"backend_bucket": bk.Bucket,
		"stacks":         stacks,
		"modules":        cfg.Modules,
		"outputs":        outputs,
	}
	renderPreview(summary)
}

func printPreviewService(svc *config.ServiceConfig, envCfg *config.EnvironmentConfig, envName string, stacks []previewStack, outputs []previewModuleOutputs, mergedVars map[string]string) {
	envEntry := envCfg.Environments[envName]
	bk, _ := backend.Resolve(envCfg.Metadata.Provider, envCfg, envEntry)
	summary := map[string]interface{}{
		"kind":           "Service",
		"name":           svc.Metadata.Name,
		"env":            envName,
		"provider":       envCfg.Metadata.Provider,
		"labels_env":     envCfg.Metadata.Labels,
		"labels_service": svc.Metadata.Labels,
		"variables":      mergedVars,
		"backend_type":   bk.Type,
		"backend_bucket": bk.Bucket,
		"stacks":         stacks,
		"modules":        svc.Modules,
		"outputs":        outputs,
	}
	renderPreview(summary)
}

func printPreviewStack(cfg *config.StackConfig, outputs []previewModuleOutputs) {
	summary := map[string]interface{}{
		"kind":      "Stack",
		"name":      cfg.Metadata.Name,
		"labels":    cfg.Metadata.Labels,
		"variables": cfg.Variables,
		"modules":   cfg.Modules,
		"outputs":   outputs,
	}
	renderPreview(summary)
}

type previewStack struct {
	Name      string            `json:"name"`
	Path      string            `json:"path"`
	Labels    map[string]string `json:"labels,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	Modules   []config.Module   `json:"modules,omitempty"`
}

func loadPreviewStacks(specPath string, stackRefs []string, embeddedRoot, customRoot string) ([]previewStack, error) {
	if len(stackRefs) == 0 {
		return nil, nil
	}
	baseDir := filepath.Dir(specPath)
	stacks := make([]previewStack, 0, len(stackRefs))
	for _, ref := range stackRefs {
		if strings.TrimSpace(ref) == "" {
			return nil, fmt.Errorf("stack reference is empty in %s", specPath)
		}
		stackPath, err := config.ResolveStackRef(ref, baseDir)
		if err != nil {
			return nil, err
		}
		stackCfg, err := config.LoadStack(stackPath)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, previewStack{
			Name:      stackCfg.Metadata.Name,
			Path:      stackPath,
			Labels:    stackCfg.Metadata.Labels,
			Variables: stackCfg.Variables,
			Modules:   stackCfg.Modules,
		})
	}
	return stacks, nil
}

func renderPreview(summary map[string]interface{}) {
	format := "table"
	if flagVerbose {
		format = "yaml"
	}
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(summary)
	case "yaml":
		out, _ := yaml.Marshal(summary)
		fmt.Print(string(out))
	default:
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(tw, "Kind\t%s\n", summary["kind"])
		fmt.Fprintf(tw, "Name\t%s\n", summary["name"])
		if envVal, ok := summary["env"]; ok && envVal != nil {
			fmt.Fprintf(tw, "Env\t%s\n", envVal)
		}
		if providerVal, ok := summary["provider"]; ok && providerVal != nil {
			fmt.Fprintf(tw, "Provider\t%s\n", providerVal)
		}
		if backendBucket, ok := summary["backend_bucket"]; ok && backendBucket != nil {
			fmt.Fprintf(tw, "Backend\t%s (%s)\n", backendBucket, summary["backend_type"])
		}
		if labels, ok := summary["labels"].(map[string]string); ok && len(labels) > 0 {
			fmt.Fprintln(tw)
			printKeyValues(tw, "Labels", labels)
		}
		if labelsSvc, ok := summary["labels_service"].(map[string]string); ok && len(labelsSvc) > 0 {
			fmt.Fprintln(tw)
			printKeyValues(tw, "Service Labels", labelsSvc)
		}
		if vars, ok := summary["variables"].(map[string]string); ok && len(vars) > 0 {
			fmt.Fprintln(tw)
			printKeyValues(tw, "Variables", vars)
		}
		fmt.Fprintln(tw)
		if stacks, ok := summary["stacks"].([]previewStack); ok && len(stacks) > 0 {
			fmt.Fprintln(tw, "Stacks:")
			fmt.Fprintln(tw, "NAME\tPATH")
			for _, stack := range stacks {
				fmt.Fprintf(tw, "%s\t%s\n", stack.Name, stack.Path)
			}
			fmt.Fprintln(tw)
			fmt.Fprintln(tw, "Stack Modules:")
			fmt.Fprintln(tw, "STACK\tID\tTYPE")
			for _, stack := range stacks {
				if len(stack.Modules) == 0 {
					continue
				}
				for _, m := range stack.Modules {
					fmt.Fprintf(tw, "%s\t%s\t%s\n", stack.Name, m.ID, m.Type)
				}
			}
			fmt.Fprintln(tw)
		}
		fmt.Fprintln(tw, "Modules:")
		fmt.Fprintln(tw, "ID\tTYPE\tOUTPUTS")
		outputMap := map[string][]string{}
		if outs, ok := summary["outputs"].([]previewModuleOutputs); ok {
			for _, out := range outs {
				if len(out.Outputs) == 0 {
					continue
				}
				outputMap[out.ID] = out.Outputs
			}
		}
		if mods, ok := summary["modules"].([]config.Module); ok {
			for _, m := range mods {
				outputs := strings.Join(outputMap[m.ID], ", ")
				fmt.Fprintf(tw, "%s\t%s\t%s\n", m.ID, m.Type, outputs)
			}
		}
		tw.Flush()
	}
}

func printKeyValues(w io.Writer, title string, values map[string]string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(w, "%s:\n", title)
	fmt.Fprintln(w, "KEY\tVALUE")
	keys := sortedMapKeys(values)
	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%s\n", k, values[k])
	}
}

func sortedMapKeys(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mergeStringMap(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

func mergeStackVariables(stacks []previewStack) map[string]string {
	if len(stacks) == 0 {
		return nil
	}
	merged := map[string]string{}
	for _, stack := range stacks {
		merged = mergeStringMap(merged, stack.Variables)
	}
	return merged
}

type previewModuleOutputs struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Outputs []string `json:"outputs,omitempty"`
}

func buildPreviewOutputs(mods []config.Module, embeddedRoot, customRoot string) ([]previewModuleOutputs, error) {
	if len(mods) == 0 {
		return nil, nil
	}
	embeddedMetas, err := config.ScanModuleMetas(embeddedRoot)
	if err != nil {
		return nil, err
	}
	var customMetas map[string]*config.ModuleMetadata
	if strings.TrimSpace(customRoot) != "" {
		customMetas, err = config.ScanModuleMetas(customRoot)
		if err != nil {
			return nil, err
		}
	}

	results := make([]previewModuleOutputs, 0, len(mods))
	for _, m := range mods {
		meta, err := clihelper.SelectModuleMeta(m, embeddedMetas, customMetas, embeddedRoot)
		if err != nil {
			return nil, err
		}
		if len(meta.Outputs) == 0 {
			continue
		}
		names := make([]string, 0, len(meta.Outputs))
		for _, out := range meta.Outputs {
			names = append(names, out.Name)
		}
		sort.Strings(names)
		results = append(results, previewModuleOutputs{
			ID:      m.ID,
			Type:    m.Type,
			Outputs: names,
		})
	}
	return results, nil
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().StringVarP(&previewFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	previewCmd.Flags().StringVarP(&previewEnv, "env", "e", "", "Environment key to use for preview")
}
