package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Masterminds/semver"
	"pltf/pkg/provider"
)

type Runner struct {
	workdir string
	stdout  io.Writer
	stderr  io.Writer
	env     []string
}

func NewRunner(workdir string, stdout, stderr io.Writer, extraEnv ...string) (*Runner, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if err := ensureTerraformVersion(provider.RequiredTfVersion); err != nil {
		return nil, err
	}
	env := os.Environ()
	if len(extraEnv) > 0 {
		env = append(env, extraEnv...)
	}
	return &Runner{
		workdir: workdir,
		stdout:  stdout,
		stderr:  stderr,
		env:     env,
	}, nil
}

func (r *Runner) EngineCmd() string {
	return "terraform"
}

func (r *Runner) Exec(args []string) (string, int, error) {
	cmd := exec.Command(r.EngineCmd(), args...)
	cmd.Dir = r.workdir
	cmd.Env = r.env

	var stdoutBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdoutBuf, r.stdout)

	cmd.Stderr = r.stderr

	err := cmd.Run()
	exit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = exitErr.ExitCode()
		} else {
			return stdoutBuf.String(), exit, err
		}
	}

	return stdoutBuf.String(), exit, err
}

func ensureTerraformVersion(constraintStr string) error {
	out, err := runCommand("", "terraform", "version", "-json")
	if err != nil {
		return fmt.Errorf("failed to determine terraform version: %w", err)
	}
	var resp struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return fmt.Errorf("unable to parse terraform version: %w", err)
	}
	version, err := semver.NewVersion(resp.TerraformVersion)
	if err != nil {
		return fmt.Errorf("unable to parse terraform version %q: %w", resp.TerraformVersion, err)
	}
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return fmt.Errorf("invalid terraform version constraint %q: %w", constraintStr, err)
	}
	if !constraint.Check(version) {
		return fmt.Errorf("terraform %s is required, but current terraform is %s; please install a matching version", constraintStr, resp.TerraformVersion)
	}
	return nil
}

func runCommand(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
