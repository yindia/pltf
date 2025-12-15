package generate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyUsedModulesSkipsTerraformArtifacts(t *testing.T) {
	src := t.TempDir()
	moduleType := "aws_base"
	moduleDir := filepath.Join(src, moduleType)
	if err := os.MkdirAll(filepath.Join(moduleDir, ".terraform"), 0o755); err != nil {
		t.Fatalf("mkdir .terraform: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "terraform.tfstate"), []byte("state"), 0o644); err != nil {
		t.Fatalf("write tfstate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte("# module"), 0o644); err != nil {
		t.Fatalf("write main.tf: %v", err)
	}

	dst := t.TempDir()
	used := map[string]bool{moduleType: true}
	rootByType := map[string]string{moduleType: src}

	if err := copyUsedModules(dst, used, rootByType); err != nil {
		t.Fatalf("copyUsedModules: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "modules", moduleType, ".terraform")); err == nil {
		t.Fatalf("expected .terraform to be skipped")
	}
	if _, err := os.Stat(filepath.Join(dst, "modules", moduleType, "terraform.tfstate")); err == nil {
		t.Fatalf("expected tfstate to be skipped")
	}
	if _, err := os.Stat(filepath.Join(dst, "modules", moduleType, "main.tf")); err != nil {
		t.Fatalf("expected main.tf to be copied: %v", err)
	}
}
