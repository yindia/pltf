package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModule(t *testing.T, root, name, provider string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	meta := []byte("name: " + name + "\ntype: " + name + "\nprovider: " + provider + "\nversion: 1.0.0\n")
	if err := os.WriteFile(filepath.Join(dir, "module.yaml"), meta, 0o644); err != nil {
		t.Fatalf("write module.yaml: %v", err)
	}
}

func TestScanModuleRootsPrecedence(t *testing.T) {
	embedded := t.TempDir()
	custom := t.TempDir()
	writeModule(t, embedded, "aws_s3", "aws")
	writeModule(t, custom, "aws_s3", "aws") // should win because custom listed first

	roots := []string{custom, embedded}
	recs, err := ScanModuleRoots(roots, []string{"aws_s3"})
	if err != nil {
		t.Fatalf("ScanModuleRoots error: %v", err)
	}
	rec, ok := recs["aws_s3"]
	if !ok {
		t.Fatalf("missing aws_s3 record")
	}
	if rec.Root != custom {
		t.Fatalf("expected custom root precedence, got %s", rec.Root)
	}
}

func TestScanModuleRootsMissingRequired(t *testing.T) {
	embedded := t.TempDir()
	writeModule(t, embedded, "aws_s3", "aws")
	_, err := ScanModuleRoots([]string{embedded}, []string{"aws_s3", "aws_eks"})
	if err == nil {
		t.Fatalf("expected error for missing required module")
	}
}
