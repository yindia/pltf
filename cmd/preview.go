package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pltf/pkg/config"
)

var (
	previewFile string
	previewEnv  string
	previewOut  string
)

// report/preview command: summarize what would be generated/applied.
var previewCmd = &cobra.Command{
	Use:   "preview",
	Args:  cobra.NoArgs,
	Short: "Preview a spec: provider, backend, modules, labels (no Terraform run)",
	Long:  "Parse a spec (Environment or Service) and show a concise summary: provider, backend type, environment, labels, and modules to be rendered.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPreview(previewFile, previewEnv)
	},
}

func runPreview(file, env string) error {
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
		printPreviewEnv(envCfg, envName)
	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		printPreviewService(svcCfg, envCfg, envName)
	default:
		return fmt.Errorf("unknown kind %q", kind)
	}
	return nil
}

func printPreviewEnv(cfg *config.EnvironmentConfig, envName string) {
	envEntry := cfg.Environments[envName]
	bk, _ := previewBackend(cfg.Metadata.Provider, cfg, envEntry)

	summary := map[string]interface{}{
		"kind":           "Environment",
		"name":           cfg.Metadata.Name,
		"env":            envName,
		"provider":       cfg.Metadata.Provider,
		"labels":         cfg.Metadata.Labels,
		"backend_type":   bk.backendType,
		"backend_bucket": bk.bucket,
		"modules":        cfg.Modules,
	}
	renderPreview(summary)
}

func printPreviewService(svc *config.ServiceConfig, envCfg *config.EnvironmentConfig, envName string) {
	envEntry := envCfg.Environments[envName]
	bk, _ := previewBackend(envCfg.Metadata.Provider, envCfg, envEntry)
	summary := map[string]interface{}{
		"kind":           "Service",
		"name":           svc.Metadata.Name,
		"env":            envName,
		"provider":       envCfg.Metadata.Provider,
		"labels_env":     envCfg.Metadata.Labels,
		"labels_service": svc.Metadata.Labels,
		"backend_type":   bk.backendType,
		"backend_bucket": bk.bucket,
		"modules":        svc.Modules,
	}
	renderPreview(summary)
}

type previewBackendDetails struct {
	backendType   string
	bucket        string
	container     string
	resourceGroup string
}

func previewBackend(provider string, envCfg *config.EnvironmentConfig, envEntry config.EnvironmentEntry) (previewBackendDetails, error) {
	bType := strings.ToLower(strings.TrimSpace(envCfg.Backend.Type))
	if bType == "" {
		bType = strings.ToLower(provider)
	}

	bucket := envCfg.Backend.Bucket
	container := envCfg.Backend.Container
	resourceGroup := envCfg.Backend.ResourceGroup

	switch bType {
	case "aws", "s3", "":
		if bucket == "" {
			bucket = fmt.Sprintf("%s-%s-%s", envCfg.Metadata.Name, envCfg.Metadata.Org, envEntry.Region)
		}
	case "gcp", "google", "gcs":
		if bucket == "" {
			bucket = fmt.Sprintf("%s-%s-%s", envCfg.Metadata.Name, envCfg.Metadata.Org, envEntry.Region)
		}
	case "azure", "azurerm":
		if bucket == "" {
			return previewBackendDetails{}, fmt.Errorf("backend.bucket required for azure backend")
		}
		if container == "" {
			container = "tfstate"
		}
		if resourceGroup == "" {
			resourceGroup = fmt.Sprintf("%s-tfstate-rg", envCfg.Metadata.Name)
		}
	default:
		return previewBackendDetails{}, fmt.Errorf("unsupported backend type %q", bType)
	}

	return previewBackendDetails{
		backendType:   bType,
		bucket:        bucket,
		container:     container,
		resourceGroup: resourceGroup,
	}, nil
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
		fmt.Fprintf(tw, "Env\t%s\n", summary["env"])
		fmt.Fprintf(tw, "Provider\t%s\n", summary["provider"])
		fmt.Fprintf(tw, "Backend\t%s (%s)\n", summary["backend_bucket"], summary["backend_type"])
		if labels, ok := summary["labels"].(map[string]string); ok && len(labels) > 0 {
			fmt.Fprintf(tw, "Labels\t%v\n", labels)
		}
		if labelsSvc, ok := summary["labels_service"].(map[string]string); ok && len(labelsSvc) > 0 {
			fmt.Fprintf(tw, "Service Labels\t%v\n", labelsSvc)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Modules:")
		fmt.Fprintln(tw, "ID\tTYPE")
		if mods, ok := summary["modules"].([]config.Module); ok {
			for _, m := range mods {
				fmt.Fprintf(tw, "%s\t%s\n", m.ID, m.Type)
			}
		}
		tw.Flush()
	}
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().StringVarP(&previewFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	previewCmd.Flags().StringVarP(&previewEnv, "env", "e", "", "Environment key to use for preview")
}
