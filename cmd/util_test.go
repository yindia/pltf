package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"pltf/pkg/backend"
	"pltf/pkg/config"
)

func TestResolveModulesRootEmbedded(t *testing.T) {
	root, err := resolveModulesRoot("")
	if err != nil {
		t.Fatalf("resolveModulesRoot(\"\") error: %v", err)
	}
	if fi, err := os.Stat(root); err != nil {
		t.Fatalf("embedded root missing: %v", err)
	} else if !fi.IsDir() {
		t.Fatalf("embedded root is not a directory: %s", root)
	}

	meta := filepath.Join(root, "aws_eks", "module.yaml")
	if _, err := os.Stat(meta); err != nil {
		t.Fatalf("expected aws_eks/module.yaml in embedded root: %v", err)
	}

	root2, err := resolveModulesRoot("")
	if err != nil {
		t.Fatalf("second resolveModulesRoot(\"\") error: %v", err)
	}
	if root != root2 {
		t.Fatalf("embedded modules root changed between calls: %s vs %s", root, root2)
	}
}

func TestResolveModulesRootCustom(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "example"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	root, err := resolveModulesRoot(tmp)
	if err != nil {
		t.Fatalf("resolveModulesRoot(custom) error: %v", err)
	}
	if root != filepath.Clean(tmp) {
		t.Fatalf("expected cleaned custom root %s, got %s", filepath.Clean(tmp), root)
	}
}

func TestBackendResolveDefaultsS3(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "aws"},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Region: "us-east-1"},
		},
	}
	envEntry := cfg.Environments["dev"]
	b, err := backend.Resolve(cfg.Metadata.Provider, cfg, envEntry)
	if err != nil {
		t.Fatalf("backend.Resolve error: %v", err)
	}
	if b.Type != "s3" {
		t.Fatalf("unexpected backend type: %s", b.Type)
	}
	expected := "org-example-tfstate"
	if b.Bucket != expected {
		t.Fatalf("expected bucket %s, got %s", expected, b.Bucket)
	}
}

func TestBackendResolveCrossCloud(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "aws"},
		Backend:  config.Backend{Type: "s3", Bucket: "custom-bkt", Region: "eu-west-1"},
		Environments: map[string]config.EnvironmentEntry{
			"prod": {Region: "us-central1"},
		},
	}
	envEntry := cfg.Environments["prod"]
	b, err := backend.Resolve(cfg.Metadata.Provider, cfg, envEntry)
	if err != nil {
		t.Fatalf("backend.Resolve error: %v", err)
	}
	if b.Bucket != "custom-bkt" || b.Region != "eu-west-1" || b.Type != "s3" {
		t.Fatalf("unexpected backend %+v", b)
	}
}

func TestBackendResolveAzureDefaults(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "azure"},
		Backend:  config.Backend{Type: "azurerm"},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Region: "eastus"},
		},
	}
	envEntry := cfg.Environments["dev"]
	b, err := backend.Resolve(cfg.Metadata.Provider, cfg, envEntry)
	if err != nil {
		t.Fatalf("unexpected error for azure defaults: %v", err)
	}
	if b.Bucket == "" || b.Type != "azurerm" {
		t.Fatalf("unexpected backend %+v", b)
	}
}
