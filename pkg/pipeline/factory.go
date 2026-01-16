package pipeline

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Provider is a CI/CD provider for workflow generation.
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderBitbucket Provider = "bitbucket"
	ProviderGitLab    Provider = "gitlab"
)

var (
	// ErrProviderNotImplemented indicates the provider is recognized but not implemented yet.
	ErrProviderNotImplemented = errors.New("pipeline provider not implemented")
)

// Workflow represents a generated pipeline workflow file.
type Workflow struct {
	Name     string
	FileName string
	Content  string
}

// Generator builds CI/CD workflows from specs.
type Generator interface {
	Generate(specPath string) (Workflow, error)
}

// NewGenerator creates a workflow generator for the specified provider.
// If provider is empty, it defaults to GitHub.
func NewGenerator(provider string) (Generator, error) {
	selected := strings.TrimSpace(provider)
	if selected == "" {
		if v := strings.TrimSpace(os.Getenv("CICD_PROVIDER")); v != "" {
			selected = v
		} else if strings.TrimSpace(os.Getenv("GITHUB_ACTIONS")) != "" {
			selected = string(ProviderGitHub)
		} else {
			selected = string(ProviderGitHub)
		}
	}

	switch strings.ToLower(selected) {
	case string(ProviderGitHub):
		return newGitHubGenerator(), nil
	case string(ProviderBitbucket), string(ProviderGitLab):
		return nil, fmt.Errorf("%w: %s", ErrProviderNotImplemented, selected)
	default:
		return nil, fmt.Errorf("unknown pipeline provider %q", selected)
	}
}
