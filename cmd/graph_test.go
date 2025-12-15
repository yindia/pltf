package cmd

import (
	"strings"
	"testing"

	"pltf/pkg/config"
)

func TestBuildSpecGraphIncludesLinksAndRefs(t *testing.T) {
	mods := []config.Module{
		{ID: "bucket", Type: "aws_s3"},
		{ID: "role", Type: "aws_iam_role"},
		{
			ID:   "svc",
			Type: "helm_chart",
			Inputs: map[string]interface{}{
				"values": map[string]interface{}{
					"arn": "${module.bucket.bucket_arn}",
					"list": []interface{}{
						"module.role.role_arn",
					},
				},
			},
			Links: config.AccessLinks{
				"read": []string{"role"},
			},
		},
	}

	out := buildSpecGraph(mods)
	if !strings.Contains(out, "\"svc\" -> \"bucket\"") {
		t.Fatalf("expected dependency from svc to bucket, got:\n%s", out)
	}
	if !strings.Contains(out, "\"svc\" -> \"role\"") {
		t.Fatalf("expected dependency from svc to role via link, got:\n%s", out)
	}
}
