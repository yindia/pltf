package runner

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Cmd struct {
	Name   string
	Args   []string
	Dir    string
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
}

type Runner interface {
	Run(cmd Cmd) error
	RunOutput(cmd Cmd) (string, error)
	RunExit(cmd Cmd) (int, error)
}

type LocalRunner struct{}

func (LocalRunner) Run(cmd Cmd) error {
	c := buildCmd(cmd)
	return c.Run()
}

func (LocalRunner) RunOutput(cmd Cmd) (string, error) {
	c := buildCmd(cmd)
	var buf bytes.Buffer
	c.Stdout = &buf
	if err := c.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (LocalRunner) RunExit(cmd Cmd) (int, error) {
	c := buildCmd(cmd)
	err := c.Run()
	if err == nil {
		return 0, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), err
	}
	return 1, err
}

var Default Runner = LocalRunner{}

func buildCmd(cmd Cmd) *exec.Cmd {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		name = "sh"
	}
	c := exec.Command(name, cmd.Args...)
	if cmd.Dir != "" {
		c.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		c.Env = cmd.Env
	}
	if cmd.Stdout != nil {
		c.Stdout = cmd.Stdout
	}
	if cmd.Stderr != nil {
		c.Stderr = cmd.Stderr
	}
	return c
}

func Describe(cmd Cmd) string {
	parts := append([]string{cmd.Name}, cmd.Args...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func WrapError(cmd Cmd, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s failed: %w", Describe(cmd), err)
}
