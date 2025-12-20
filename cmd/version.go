package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"pltf/modules"
	"pltf/pkg/generate"
	"pltf/pkg/version"
)

type tfVersionJSON struct {
	TerraformVersion   string            `json:"terraform_version"`
	ProviderSelections map[string]string `json:"provider_selections"`
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Args:  cobra.NoArgs,
	Short: "Show pltf version plus Terraform and key provider versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		printVersions()
		return nil
	},
}

func printVersions() {
	fmt.Printf("pltf version: %s\n", cliVersion())

	tfVer, provs, err := terraformVersions()
	if err != nil {
		fmt.Printf("terraform: not available (%v)\n", err)
	} else {
		fmt.Printf("terraform: %s\n", tfVer)
	}

	// Normalize key providers
	fmt.Println("providers:")
	for _, key := range []string{
		"registry.terraform.io/hashicorp/aws",
		"registry.terraform.io/hashicorp/google",
		"registry.terraform.io/hashicorp/azurerm",
	} {
		name := strings.TrimPrefix(key, "registry.terraform.io/hashicorp/")
		val := provs[key]
		if val == "" {
			val = "n/a"
		}
		fmt.Printf("  - %s: %s\n", name, val)
	}

	// Show any other providers if present
	for k, v := range provs {
		if strings.Contains(k, "hashicorp/aws") || strings.Contains(k, "hashicorp/google") || strings.Contains(k, "hashicorp/azurerm") {
			continue
		}
		fmt.Printf("  - %s: %s\n", k, v)
	}
}

func cliVersion() string {
	if v := strings.TrimSpace(os.Getenv("PLTF_VERSION")); v != "" {
		return v
	}
	if strings.TrimSpace(version.Version) != "" && version.Version != "dev" {
		return version.Version
	}
	return "dev"
}

func terraformVersions() (string, map[string]string, error) {
	out, err := runCmdOutput(".", "terraform", "version", "-json")
	if err != nil {
		// Still return defaults if terraform is missing
		return generate.RequiredTfVersion, providerDefaults(), err
	}
	var parsed tfVersionJSON
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return generate.RequiredTfVersion, providerDefaults(), err
	}

	// Start with defaults.
	provs := providerDefaults()
	// Overlay embedded versions.
	for k, v := range embeddedProviderVersions() {
		if v != "" {
			provs[k] = v
		}
	}
	// Overlay actual terraform selections (most authoritative).
	for k, v := range parsed.ProviderSelections {
		if v != "" {
			provs[k] = v
		}
	}

	return parsed.TerraformVersion, provs, nil
}

func providerDefaults() map[string]string {
	return map[string]string{
		"registry.terraform.io/hashicorp/aws":     generate.AWSProviderVersion,
		"registry.terraform.io/hashicorp/google":  generate.GCPProviderVersion,
		"registry.terraform.io/hashicorp/azurerm": generate.AzureProviderVersion,
	}
}

func embeddedProviderVersions() map[string]string {
	out := map[string]string{}
	root, err := modules.Materialize()
	if err != nil {
		return out
	}
	targets := map[string]struct{}{
		"aws":     {},
		"google":  {},
		"azurerm": {},
	}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "versions.tf" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		inProviders := false
		brace := 0
		current := ""
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(t, "required_providers") {
				inProviders = true
				brace += strings.Count(t, "{") - strings.Count(t, "}")
				continue
			}
			if inProviders {
				brace += strings.Count(t, "{") - strings.Count(t, "}")
				if brace <= 0 {
					inProviders = false
					current = ""
					continue
				}
				for name := range targets {
					if strings.HasPrefix(t, name) {
						current = name
						break
					}
				}
				if strings.HasPrefix(t, "version") && current != "" {
					parts := strings.Split(t, "=")
					if len(parts) == 2 {
						v := strings.Trim(strings.TrimSpace(parts[1]), "\"")
						key := "registry.terraform.io/hashicorp/" + current
						if out[key] == "" {
							out[key] = v
						}
					}
				}
			}
		}
		return nil
	})
	return out
}
