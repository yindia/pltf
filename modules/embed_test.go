package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaterializeCopiesModules(t *testing.T) {
	root, err := Materialize()
	if err != nil {
		t.Fatalf("Materialize returned error: %v", err)
	}

	if fi, err := os.Stat(root); err != nil {
		t.Fatalf("materialized root missing: %v", err)
	} else if !fi.IsDir() {
		t.Fatalf("materialized root is not a directory: %s", root)
	}

	// Spot-check a known module metadata file to ensure embed + copy worked.
	metaPath := filepath.Join(root, "aws_eks", "module.yaml")
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("expected embedded module metadata at %s: %v", metaPath, err)
	}

	// Subsequent calls should return the same path (singleton).
	root2, err := Materialize()
	if err != nil {
		t.Fatalf("second Materialize returned error: %v", err)
	}
	if root2 != root {
		t.Fatalf("Materialize returned different paths across calls: %s vs %s", root, root2)
	}
}
