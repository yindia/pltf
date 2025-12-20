package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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
		return "", nil, err
	}
	var parsed tfVersionJSON
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return "", nil, err
	}
	return parsed.TerraformVersion, parsed.ProviderSelections, nil
}
