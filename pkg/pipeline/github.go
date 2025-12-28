package pipeline

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"pltf/pkg/config"
)

type githubGenerator struct{}

func newGitHubGenerator() *githubGenerator {
	return &githubGenerator{}
}

func (g *githubGenerator) Generate(specPath string) (Workflow, error) {
	kind, err := config.DetectKind(specPath)
	if err != nil {
		return Workflow{}, err
	}

	var envs []string
	switch kind {
	case "Environment":
		cfg, err := config.LoadEnvironmentConfig(specPath)
		if err != nil {
			return Workflow{}, err
		}
		envs = sortedKeys(cfg.Environments)
	case "Service":
		svc, _, err := config.LoadService(specPath)
		if err != nil {
			return Workflow{}, err
		}
		envs = sortedKeysEnvRef(svc.Metadata.EnvRef)
	default:
		return Workflow{}, fmt.Errorf("unsupported kind %q for pipeline generation", kind)
	}

	if len(envs) == 0 {
		return Workflow{}, fmt.Errorf("no environments found in %s", specPath)
	}

	specBase := strings.TrimSuffix(filepath.Base(specPath), filepath.Ext(specPath))
	name := fmt.Sprintf("pltf: %s", specBase)
	fileName := fmt.Sprintf("pltf-%s.yml", specBase)

	content := buildGitHubWorkflow(name, specPath, envs)
	return Workflow{Name: name, FileName: fileName, Content: content}, nil
}

func buildGitHubWorkflow(name, specPath string, envs []string) string {
	envList := yamlInlineList(envs)
	var b strings.Builder
	b.WriteString("name: ")
	b.WriteString(name)
	b.WriteString("\n\n")
	b.WriteString("on:\n")
	b.WriteString("  pull_request:\n")
	b.WriteString("    branches: [ \"**\" ]\n")
	b.WriteString("  push:\n")
	b.WriteString("    branches: [ \"main\" ]\n\n")
	b.WriteString("jobs:\n")
	b.WriteString("  plan:\n")
	b.WriteString("    if: github.event_name == 'pull_request'\n")
	b.WriteString("    runs-on: ubuntu-latest\n")
	b.WriteString("    strategy:\n")
	b.WriteString("      matrix:\n")
	b.WriteString("        env: ")
	b.WriteString(envList)
	b.WriteString("\n")
	b.WriteString("    steps:\n")
	b.WriteString("      - uses: actions/checkout@v4\n")
	b.WriteString("      - uses: actions/setup-go@v5\n")
	b.WriteString("        with:\n")
	b.WriteString("          go-version-file: go.mod\n")
	b.WriteString("      - uses: hashicorp/setup-terraform@v3\n")
	b.WriteString("        with:\n")
	b.WriteString("          terraform_version: 1.8.5\n")
	b.WriteString("      - name: Install pltf\n")
	b.WriteString("        run: |\n")
	b.WriteString("          go install -ldflags \"-X 'pltf/pkg/version.Version=${{ github.sha }}'\" ./...\n")
	b.WriteString("          echo \"$HOME/go/bin\" >> \"$GITHUB_PATH\"\n")
	b.WriteString("      - name: Validate + scan\n")
	b.WriteString("        env:\n")
	b.WriteString("          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
	b.WriteString("          # TODO: add cloud credentials here\n")
	b.WriteString("        run: |\n")
	b.WriteString("          pltf validate -f ")
	b.WriteString(specPath)
	b.WriteString(" --env ${{ matrix.env }} --scan\n")
	b.WriteString("      - name: Plan\n")
	b.WriteString("        env:\n")
	b.WriteString("          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
	b.WriteString("          # TODO: add cloud credentials here\n")
	b.WriteString("        run: |\n")
	b.WriteString("          pltf terraform plan -f ")
	b.WriteString(specPath)
	b.WriteString(" --env ${{ matrix.env }} --scan\n\n")
	b.WriteString("  apply:\n")
	b.WriteString("    if: github.event_name == 'push'\n")
	b.WriteString("    runs-on: ubuntu-latest\n")
	b.WriteString("    strategy:\n")
	b.WriteString("      matrix:\n")
	b.WriteString("        env: ")
	b.WriteString(envList)
	b.WriteString("\n")
	b.WriteString("    steps:\n")
	b.WriteString("      - uses: actions/checkout@v4\n")
	b.WriteString("      - uses: actions/setup-go@v5\n")
	b.WriteString("        with:\n")
	b.WriteString("          go-version-file: go.mod\n")
	b.WriteString("      - uses: hashicorp/setup-terraform@v3\n")
	b.WriteString("        with:\n")
	b.WriteString("          terraform_version: 1.8.5\n")
	b.WriteString("      - name: Install pltf\n")
	b.WriteString("        run: |\n")
	b.WriteString("          go install -ldflags \"-X 'pltf/pkg/version.Version=${{ github.sha }}'\" ./...\n")
	b.WriteString("          echo \"$HOME/go/bin\" >> \"$GITHUB_PATH\"\n")
	b.WriteString("      - name: Validate\n")
	b.WriteString("        env:\n")
	b.WriteString("          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
	b.WriteString("          # TODO: add cloud credentials here\n")
	b.WriteString("        run: |\n")
	b.WriteString("          pltf validate -f ")
	b.WriteString(specPath)
	b.WriteString(" --env ${{ matrix.env }}\n")
	b.WriteString("      - name: Apply\n")
	b.WriteString("        env:\n")
	b.WriteString("          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
	b.WriteString("          # TODO: add cloud credentials here\n")
	b.WriteString("        run: |\n")
	b.WriteString("          pltf terraform apply -f ")
	b.WriteString(specPath)
	b.WriteString(" --env ${{ matrix.env }} --auto-approve\n")
	return b.String()
}

func yamlInlineList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("%q", item))
	}
	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysEnvRef(m map[string]config.ServiceEnvRefEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
