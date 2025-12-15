package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"pltf/pkg/config"
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

func runTerraformGraph(file, env, modules, out string, vars []string, outFile string, planFile string) error {
	if err := autoGenerateQuiet(file, env, modules, out, vars); err != nil {
		return err
	}

	ctx, err := prepareStackContext(file, env, out)
	if err != nil {
		return err
	}

	bk, err := computeBackend(ctx.envCfg, ctx.env)
	if err != nil {
		return err
	}
	if isS3Backend(bk) {
		if err := ensureS3Bucket(bk.bucket, bk.region); err != nil {
			return fmt.Errorf("failed to ensure backend bucket: %w", err)
		}
	}

	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = ctx.outDir
	initCmd.Stdout = io.Discard
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	args := []string{"graph"}
	if strings.TrimSpace(planFile) != "" {
		args = append(args, "-plan="+planFile)
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = ctx.outDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform graph failed: %w", err)
	}
	return writeGraphOutput(buf.Bytes(), outFile)
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
	file = defaultString(file, "env.yaml")
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
		envName, err := selectEnvName(kind, env, envCfg, nil)
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
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return "", err
		}
		_ = envName
		mods := append([]config.Module{}, envCfg.Modules...)
		mods = append(mods, svcCfg.Modules...)
		return buildSpecGraph(mods), nil
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
