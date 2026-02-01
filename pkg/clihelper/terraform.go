package clihelper

import (
	"fmt"
	"io"
	"path/filepath"
)

// TerraformExecOptions encapsulates Terraform flag helpers used by commands.
type TerraformExecOptions struct {
	Targets      []string
	Parallelism  int
	Lock         bool
	LockTimeout  string
	NoColor      bool
	Input        bool
	Refresh      *bool
	PlanFile     string
	DetailedExit bool
	JSONOutput   bool
	AutoApprove  bool
	Rover        bool
	Scan         bool
	Cost         bool
	Stdout       io.Writer
	Stderr       io.Writer
}

// AppendTerraformArgs appends the standard Terraform flags onto the provided args slice.
func AppendTerraformArgs(args []string, opts TerraformExecOptions) []string {
	if opts.NoColor {
		args = append(args, "-no-color")
	}
	if !opts.Input {
		args = append(args, "-input=false")
	}
	if !opts.Lock {
		args = append(args, "-lock=false")
	}
	if opts.LockTimeout != "" {
		args = append(args, "-lock-timeout="+opts.LockTimeout)
	}
	if opts.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", opts.Parallelism))
	}
	for _, t := range opts.Targets {
		args = append(args, "-target="+t)
	}
	if opts.Refresh != nil {
		args = append(args, fmt.Sprintf("-refresh=%t", *opts.Refresh))
	}
	if opts.AutoApprove {
		args = append(args, "-auto-approve")
	}
	return args
}

// SanitizePlanArgs replaces absolute tfvars paths with workspace-relative paths when terraforming the command log.
func SanitizePlanArgs(planArgs []string, outDir, tfvarsArg, rootDir string) []string {
	sanitized := append([]string(nil), planArgs...)
	if tfvarsArg == "" {
		return sanitized
	}
	absVar := "-var-file=" + filepath.Join(outDir, tfvarsArg)
	relPath := tfvarsArg
	if rootDir != "" {
		if rel, err := filepath.Rel(rootDir, outDir); err == nil && rel != "." && rel != "" {
			relPath = filepath.Join(rel, tfvarsArg)
		}
	}
	relPath = filepath.ToSlash(filepath.Clean(relPath))
	if relPath == "." {
		relPath = tfvarsArg
	}
	relVar := "-var-file=" + relPath
	for i, arg := range sanitized {
		if arg == absVar {
			sanitized[i] = relVar
		}
	}
	return sanitized
}
