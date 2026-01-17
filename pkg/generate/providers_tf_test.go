package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestHelmKubeConfigTokensUsesAttribute(t *testing.T) {
	cluster := &clusterRef{
		host: hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: "eks"},
			hcl.TraverseAttr{Name: "endpoint"},
		},
		caData: hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: "eks"},
			hcl.TraverseAttr{Name: "ca_data"},
		},
		token: hcl.Traversal{
			hcl.TraverseRoot{Name: "data"},
			hcl.TraverseAttr{Name: "aws_eks_cluster_auth"},
			hcl.TraverseAttr{Name: "this"},
			hcl.TraverseAttr{Name: "token"},
		},
	}

	file := hclwrite.NewEmptyFile()
	body := file.Body()
	helm := body.AppendNewBlock("provider", []string{"helm"})
	helm.Body().SetAttributeRaw("kubernetes", helmKubeConfigTokens(cluster))

	out := string(file.Bytes())
	if strings.Contains(out, "kubernetes {\n") {
		t.Fatalf("expected kubernetes to be set as attribute, got block:\n%s", out)
	}
	if !strings.Contains(out, `kubernetes = {`) {
		t.Fatalf("expected kubernetes attribute in helm provider:\n%s", out)
	}
	if !strings.Contains(out, "cluster_ca_certificate = base64decode(module.eks.ca_data)") {
		t.Fatalf("expected CA decode expression, got:\n%s", out)
	}
}

func TestProvidersTFIncludesKubernetesAndKustomize(t *testing.T) {
	outDir := t.TempDir()
	cluster := &clusterRef{
		host: hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: "eks"},
			hcl.TraverseAttr{Name: "endpoint"},
		},
		caData: hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: "eks"},
			hcl.TraverseAttr{Name: "ca_data"},
		},
		token: hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: "eks"},
			hcl.TraverseAttr{Name: "token"},
		},
	}

	if err := writeProvidersTF(outDir, "aws", "us-east-1", "111111111111", true, false, true, cluster, false); err != nil {
		t.Fatalf("writeProvidersTF error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "providers.tf"))
	if err != nil {
		t.Fatalf("read providers.tf: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, `provider "kubernetes"`) {
		t.Fatalf("expected kubernetes provider block, got:\n%s", out)
	}
	if !strings.Contains(out, `provider "kustomization"`) {
		t.Fatalf("expected kustomization provider block, got:\n%s", out)
	}
	if !strings.Contains(out, "host") || !strings.Contains(out, "module.eks.endpoint") {
		t.Fatalf("expected host traversal, got:\n%s", out)
	}
	if !strings.Contains(out, "cluster_ca_certificate = base64decode(module.eks.ca_data)") {
		t.Fatalf("expected CA decode expression, got:\n%s", out)
	}
}
