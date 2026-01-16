package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvironmentConfig_WithStack(t *testing.T) {
	tmpDir := t.TempDir()

	stackPath := filepath.Join(tmpDir, "stack.yaml")
	stackYAML := `apiVersion: platform.io/v1
kind: Stack
metadata:
  name: eks-observability
  labels:
    team: platform
    owner: stack
variables:
  cluster_name: default-cluster
  region: us-west-2
modules:
  - id: base_env
    type: aws_base
  - id: obs
    type: aws_observability
`
	if err := os.WriteFile(stackPath, []byte(stackYAML), 0o600); err != nil {
		t.Fatalf("write stack: %v", err)
	}

	envPath := filepath.Join(tmpDir, "env.yaml")
	envYAML := `apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example
  org: acme
  provider: aws
  labels:
    owner: env
  stacks:
    - ./stack.yaml
backend:
  type: s3
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
modules:
  - id: dns
    type: aws_dns
`
	if err := os.WriteFile(envPath, []byte(envYAML), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := LoadEnvironmentConfig(envPath)
	if err != nil {
		t.Fatalf("load environment: %v", err)
	}

	if got := cfg.Metadata.Labels["team"]; got != "platform" {
		t.Fatalf("labels.team = %q, want %q", got, "platform")
	}
	if got := cfg.Metadata.Labels["owner"]; got != "env" {
		t.Fatalf("labels.owner = %q, want %q", got, "env")
	}

	if len(cfg.Modules) != 3 {
		t.Fatalf("modules len = %d, want 3", len(cfg.Modules))
	}
	if cfg.Modules[0].ID != "base_env" {
		t.Fatalf("modules[0].id = %q, want base_env", cfg.Modules[0].ID)
	}
	if cfg.Modules[1].ID != "obs" {
		t.Fatalf("modules[1].id = %q, want obs", cfg.Modules[1].ID)
	}
	if cfg.Modules[2].ID != "dns" {
		t.Fatalf("modules[2].id = %q, want dns", cfg.Modules[2].ID)
	}
	if got := cfg.Environments["dev"].Variables["cluster_name"]; got != "default-cluster" {
		t.Fatalf("env variables.cluster_name = %q, want %q", got, "default-cluster")
	}
	if got := cfg.Environments["dev"].Variables["region"]; got != "us-west-2" {
		t.Fatalf("env variables.region = %q, want %q", got, "us-west-2")
	}
}

func TestLoadServiceConfig_WithStackVariables(t *testing.T) {
	tmpDir := t.TempDir()

	stackPath := filepath.Join(tmpDir, "stack.yaml")
	stackYAML := `apiVersion: platform.io/v1
kind: Stack
metadata:
  name: eks-observability
variables:
  cluster_name: default-cluster
  region: us-west-2
modules:
  - id: base_env
    type: aws_base
`
	if err := os.WriteFile(stackPath, []byte(stackYAML), 0o600); err != nil {
		t.Fatalf("write stack: %v", err)
	}

	envPath := filepath.Join(tmpDir, "env.yaml")
	envYAML := `apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example
  org: acme
  provider: aws
backend:
  type: s3
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
modules:
  - id: base
    type: aws_base
`
	if err := os.WriteFile(envPath, []byte(envYAML), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	servicePath := filepath.Join(tmpDir, "service.yaml")
	serviceYAML := `apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml
  stacks:
    - ./stack.yaml
  envRef:
    dev: {}
modules:
  - id: app
    type: aws_k8s_service
    inputs:
      cluster_name: var.cluster_name
`
	if err := os.WriteFile(servicePath, []byte(serviceYAML), 0o600); err != nil {
		t.Fatalf("write service: %v", err)
	}

	svcCfg, _, err := LoadService(servicePath)
	if err != nil {
		t.Fatalf("load service: %v", err)
	}

	envEntry := svcCfg.Metadata.EnvRef["dev"]
	if got := envEntry.Variables["cluster_name"]; got != "default-cluster" {
		t.Fatalf("service variables.cluster_name = %q, want %q", got, "default-cluster")
	}
	if got := envEntry.Variables["region"]; got != "us-west-2" {
		t.Fatalf("service variables.region = %q, want %q", got, "us-west-2")
	}
}

func TestLoadEnvironmentConfig_RejectsVariableOverride(t *testing.T) {
	tmpDir := t.TempDir()

	stackPath := filepath.Join(tmpDir, "stack.yaml")
	stackYAML := `apiVersion: platform.io/v1
kind: Stack
metadata:
  name: example-stack
variables:
  cluster_name: default-cluster
modules:
  - id: base
    type: aws_base
`
	if err := os.WriteFile(stackPath, []byte(stackYAML), 0o600); err != nil {
		t.Fatalf("write stack: %v", err)
	}

	envPath := filepath.Join(tmpDir, "env.yaml")
	envYAML := `apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example
  org: acme
  provider: aws
  stacks:
    - ./stack.yaml
variables:
  cluster_name: override-cluster
backend:
  type: s3
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
modules:
  - id: base
    type: aws_base
`
	if err := os.WriteFile(envPath, []byte(envYAML), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	if _, err := LoadEnvironmentConfig(envPath); err == nil {
		t.Fatalf("expected error for stack variable override, got nil")
	}
}
