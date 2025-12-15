package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"pltf/pkg/config"
)

func TestPrintModulesTable(t *testing.T) {
	metas := map[string]*config.ModuleMetadata{
		"aws_s3":  {Type: "aws_s3", Name: "aws_s3", Provider: "aws", Version: "1.0.0", Description: "S3 bucket"},
		"aws_eks": {Type: "aws_eks", Name: "aws_eks", Provider: "aws", Version: "1.0.0", Description: "EKS cluster"},
	}
	var buf bytes.Buffer
	swap := osStdoutSwap(&buf)
	if err := printModules(metas, "table"); err != nil {
		t.Fatalf("printModules error: %v", err)
	}
	swap()
	out := buf.String()
	if !strings.Contains(out, "aws_s3") || !strings.Contains(out, "aws_eks") {
		t.Fatalf("missing module names in output: %s", out)
	}
}

func TestPrintModuleDetailJSON(t *testing.T) {
	meta := &config.ModuleMetadata{
		Type:     "aws_s3",
		Name:     "aws_s3",
		Provider: "aws",
		Version:  "1.0.0",
		Inputs:   []config.InputSpec{{Name: "bucket_name", Type: "string", Required: true}},
		Outputs:  []config.OutputSpec{{Name: "bucket_arn", Type: "string"}},
	}
	var buf bytes.Buffer
	swap := osStdoutSwap(&buf)
	if err := printModuleDetail(meta, "json"); err != nil {
		t.Fatalf("printModuleDetail error: %v", err)
	}
	swap()
	if !strings.Contains(buf.String(), "bucket_name") {
		t.Fatalf("expected input name in JSON output, got: %s", buf.String())
	}
}

// osStdoutSwap redirects stdout for test capture.
func osStdoutSwap(dst *bytes.Buffer) func() {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan struct{})
	go func() {
		_, _ = dst.ReadFrom(r)
		close(done)
	}()
	return func() {
		w.Close()
		<-done
		os.Stdout = old
	}
}
