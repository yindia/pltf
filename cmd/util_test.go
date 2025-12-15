package cmd

import (
	"os"
	"path/filepath"
	"testing"

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

func TestComputeBackendDefaultsS3(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "aws"},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Region: "us-east-1"},
		},
	}
	b, err := computeBackend(cfg, "dev")
	if err != nil {
		t.Fatalf("computeBackend error: %v", err)
	}
	if b.typeName != "aws" && b.typeName != "s3" && b.typeName != "" {
		t.Fatalf("unexpected backend type: %s", b.typeName)
	}
	expected := "example-org-us-east-1"
	if b.bucket != expected {
		t.Fatalf("expected bucket %s, got %s", expected, b.bucket)
	}
}

func TestComputeBackendCrossCloud(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "gcp"},
		Backend:  config.Backend{Type: "s3", Bucket: "custom-bkt", Region: "eu-west-1"},
		Environments: map[string]config.EnvironmentEntry{
			"prod": {Region: "us-central1"},
		},
	}
	b, err := computeBackend(cfg, "prod")
	if err != nil {
		t.Fatalf("computeBackend error: %v", err)
	}
	if b.bucket != "custom-bkt" || b.region != "eu-west-1" || b.typeName != "s3" {
		t.Fatalf("unexpected backend %+v", b)
	}
}

func TestComputeBackendAzureRequiresBucket(t *testing.T) {
	cfg := &config.EnvironmentConfig{
		Metadata: config.EnvironmentMetadata{Name: "example", Org: "org", Provider: "azure"},
		Backend:  config.Backend{Type: "azurerm"},
		Environments: map[string]config.EnvironmentEntry{
			"dev": {Region: "eastus"},
		},
	}
	if _, err := computeBackend(cfg, "dev"); err == nil {
		t.Fatalf("expected error for missing azure bucket")
	}
}

func TestIsS3BackendSupportsCrossProvider(t *testing.T) {
	cases := []struct {
		name string
		b    backendDetails
		want bool
	}{
		{"empty defaults to s3", backendDetails{typeName: ""}, true},
		{"aws explicit", backendDetails{typeName: "aws"}, true},
		{"s3 explicit", backendDetails{typeName: "s3"}, true},
		{"gcs", backendDetails{typeName: "gcs"}, false},
		{"azure", backendDetails{typeName: "azurerm"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isS3Backend(tc.b); got != tc.want {
				t.Fatalf("isS3Backend(%+v) = %v, want %v", tc.b, got, tc.want)
			}
		})
	}
}
