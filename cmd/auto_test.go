package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"

	"pltf/pkg/config"
)

func writeYAML(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := yaml.Marshal(v)
	if err != nil {
		t.Fatalf("marshal yaml: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func resetProfileCache() {
	profileOnce = sync.Once{}
	profileData = nil
	profileErr = nil
}

func TestAutoValidateRequiresEnvSelection(t *testing.T) {
	t.Parallel()
	resetProfileCache()
	_ = os.Unsetenv("PLTF_DEFAULT_ENV")

	envCfg := config.EnvironmentConfig{
		APIVersion: "v1",
		Kind:       "Environment",
		Metadata: config.EnvironmentMetadata{
			Name:     "example",
			Org:      "acme",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev":  {Account: "111111111111", Region: "us-east-1"},
			"prod": {Account: "222222222222", Region: "us-west-2"},
		},
		Modules: []config.Module{
			{ID: "base", Type: "network"},
		},
	}

	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.yaml")
	writeYAML(t, envPath, envCfg)

	var buf bytes.Buffer
	err := autoValidateWithOutput(&buf, envPath, "")
	if err == nil || !strings.Contains(err.Error(), "--env is required") {
		t.Fatalf("expected env selection error, got %v (output=%s)", err, buf.String())
	}
}

func TestAutoValidateEmitsLintSuggestions(t *testing.T) {
	t.Parallel()
	resetProfileCache()
	_ = os.Unsetenv("PLTF_DEFAULT_ENV")

	envCfg := config.EnvironmentConfig{
		APIVersion: "v1",
		Kind:       "Environment",
		Metadata: config.EnvironmentMetadata{
			Name:     "demo",
			Org:      "acme",
			Provider: "aws",
		},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {
				Account:   "111111111111",
				Region:    "us-east-1",
				Variables: map[string]string{"unused": "value"},
			},
		},
		Modules: []config.Module{
			{ID: "eks", Type: "aws_eks"},
		},
	}

	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.yaml")
	writeYAML(t, envPath, envCfg)

	var buf bytes.Buffer
	if err := autoValidateWithOutput(&buf, envPath, "dev"); err != nil {
		t.Fatalf("autoValidateWithOutput returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Environment \"demo\" is valid") {
		t.Fatalf("expected validation output, got: %s", out)
	}
}
