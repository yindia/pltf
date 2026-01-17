package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"pltf/pkg/backend"
	"pltf/pkg/clihelper"
	"pltf/pkg/config"
	terraform "pltf/pkg/terraform"
)

func runGraph(mode, file, env, modules, out string, vars []string, outFile string, planFile string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "terraform", "":
		return runTerraformGraph(file, env, modules, out, vars, outFile, planFile)
	case "spec":
		dot, err := buildSpecGraphFromFile(file, env)
		if err != nil {
			return err
		}
		return writeGraphOutput([]byte(dot), outFile)
	default:
		return fmt.Errorf("unknown graph mode %q (expected terraform or spec)", mode)
	}
}

func runTerraformGraph(file, env, modules, out string, vars []string, outFile string, planFile string) (retErr error) {
	if err := clihelper.AutoGenerateQuiet(file, env, modules, out, vars); err != nil {
		return err
	}

	ctx, err := prepareWorkspaceContext(file, env, out)
	if err != nil {
		return err
	}

	_, finishRun := clihelper.StartLocalRun("graph", file, ctx.env, ctx.outDir)
	defer func() {
		finishRun(retErr)
	}()

	envEntry := ctx.envCfg.Environments[ctx.env]
	bk, err := backend.Resolve(ctx.envCfg.Metadata.Provider, ctx.envCfg, envEntry)
	if err != nil {
		return err
	}
	if err := backend.Ensure(context.Background(), bk); err != nil {
		return fmt.Errorf("failed to ensure backend: %w", err)
	}

	tfRunner, err := terraform.NewRunner(ctx.outDir, io.Discard, os.Stderr)
	if err != nil {
		return err
	}

	if err := runTerraformInitWithRetry(tfRunner); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	args := []string{"graph"}
	tfvarsName := fmt.Sprintf("%s.tfvars", ctx.env)
	args = append(args, "-var-file="+tfvarsName)
	if strings.TrimSpace(planFile) != "" {
		args = append(args, "-plan="+planFile)
	}

	output, _, err := tfRunner.Exec(args)
	if err != nil {
		return fmt.Errorf("terraform graph failed: %w", err)
	}
	return writeGraphOutput([]byte(output), outFile)
}

func writeGraphOutput(data []byte, outFile string) error {
	if strings.TrimSpace(outFile) == "" {
		fmt.Print(string(data))
		return nil
	}
	outFile = filepath.Clean(outFile)
	if err := os.WriteFile(outFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write graph to %s: %w", outFile, err)
	}
	fmt.Printf("Wrote graph to %s\n", outFile)
	return nil
}

func buildSpecGraphFromFile(file, env string) (string, error) {
	file = clihelper.DefaultString(file, "env.yaml")
	kind, err := config.DetectKind(file)
	if err != nil {
		return "", err
	}

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return "", err
		}
		envName, err := clihelper.SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return "", err
		}
		_ = envName
		return buildSpecGraph(envCfg.Modules), nil
	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return "", err
		}
		envName, err := clihelper.SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return "", err
		}
		_ = envName
		mods := append([]config.Module{}, envCfg.Modules...)
		mods = append(mods, svcCfg.Modules...)
		return buildSpecGraph(mods), nil
	case "Stack":
		return "", fmt.Errorf("stack specs cannot be graphed directly; reference them from Environment or Service specs")
	default:
		return "", fmt.Errorf("unknown kind %q", kind)
	}
}

func buildSpecGraph(mods []config.Module) string {
	deps := collectSpecDeps(mods)
	nodes := make([]string, 0, len(mods))
	for _, m := range mods {
		nodes = append(nodes, m.ID)
		if _, ok := deps[m.ID]; !ok {
			deps[m.ID] = map[string]struct{}{}
		}
	}
	sort.Strings(nodes)

	var b strings.Builder
	b.WriteString("digraph modules {\n  rankdir=LR;\n")
	for _, n := range nodes {
		b.WriteString(fmt.Sprintf("  \"%s\";\n", n))
	}
	for _, src := range nodes {
		var targets []string
		for tgt := range deps[src] {
			targets = append(targets, tgt)
		}
		sort.Strings(targets)
		for _, tgt := range targets {
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", src, tgt))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

var moduleRefScan = regexp.MustCompile(`module\.([A-Za-z0-9_.-]+)\.[A-Za-z0-9_]+`)

func collectSpecDeps(mods []config.Module) map[string]map[string]struct{} {
	deps := map[string]map[string]struct{}{}
	add := func(from, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		if _, ok := deps[from]; !ok {
			deps[from] = map[string]struct{}{}
		}
		deps[from][to] = struct{}{}
	}

	index := map[string]struct{}{}
	for _, m := range mods {
		index[m.ID] = struct{}{}
	}

	for _, m := range mods {
		for _, targets := range m.Links {
			for _, t := range targets {
				if _, ok := index[t]; ok {
					add(m.ID, t)
				}
			}
		}
		for _, v := range m.Inputs {
			collectSpecDepsFromValue(m.ID, v, add, index)
		}
	}
	return deps
}

func collectSpecDepsFromValue(modID string, v interface{}, add func(string, string), index map[string]struct{}) {
	switch val := v.(type) {
	case string:
		for _, match := range moduleRefScan.FindAllStringSubmatch(val, -1) {
			if len(match) > 1 {
				if _, ok := index[match[1]]; ok {
					add(modID, match[1])
				}
			}
		}
	case []interface{}:
		for _, item := range val {
			collectSpecDepsFromValue(modID, item, add, index)
		}
	case []string:
		for _, item := range val {
			collectSpecDepsFromValue(modID, item, add, index)
		}
	case map[string]interface{}:
		for _, item := range val {
			collectSpecDepsFromValue(modID, item, add, index)
		}
	case map[string]string:
		for _, item := range val {
			collectSpecDepsFromValue(modID, item, add, index)
		}
	}
}
