package cmd

import (
	"testing"

	"pltf/pkg/config"
)

func TestFindUnusedVars(t *testing.T) {
	envVars := map[string]string{"unused": "x"}
	svcVars := map[string]string{"svc_only": "y"}
	mods := []config.Module{
		{
			ID:   "eks",
			Type: "aws_eks",
			Inputs: map[string]interface{}{
				"cluster_name": "var.cluster_name",
			},
		},
	}
	unused := findUnusedVars(envVars, svcVars, mods)
	if len(unused) != 2 {
		t.Fatalf("expected 2 unused vars, got %v", unused)
	}
	want := map[string]struct{}{"unused": {}, "svc_only": {}}
	for _, v := range unused {
		if _, ok := want[v]; !ok {
			t.Fatalf("unexpected unused var %s", v)
		}
	}
}
