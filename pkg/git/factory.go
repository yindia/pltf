package git

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Provider is a git hosting provider for PR comments.
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderBitbucket Provider = "bitbucket"
	ProviderGitLab    Provider = "gitlab"
)

var (
	// ErrNoProvider indicates no supported provider could be detected from env.
	ErrNoProvider = errors.New("no supported git provider detected")
	// ErrProviderNotImplemented indicates the provider is recognized but not implemented yet.
	ErrProviderNotImplemented = errors.New("git provider not implemented")
)

// PRComment identifies the comment body and optional marker to find existing comments.
type PRComment struct {
	Body   string
	Marker string
}

// ReviewStatus describes a single reviewer's latest status.
type ReviewStatus struct {
	Name   string
	Team   string
	Status string
}

// ReviewSummary aggregates review status and approvals.
type ReviewSummary struct {
	Approvals int
	Reviews   []ReviewStatus
}

// Commenter posts or updates a pull/merge request comment.
type Commenter interface {
	UpsertPRComment(comment PRComment) error
	GetReviewSummary() (ReviewSummary, error)
}

// NewCommenter creates a commenter for the specified provider or auto-detected provider if empty.
func NewCommenter(provider string) (Commenter, error) {
	selected := strings.TrimSpace(provider)
	if selected == "" {
		if v := strings.TrimSpace(os.Getenv("GIT_PROVIDER")); v != "" {
			selected = v
		} else if hasGitHubEnv() {
			selected = string(ProviderGitHub)
		} else {
			return nil, ErrNoProvider
		}
	}

	switch strings.ToLower(selected) {
	case string(ProviderGitHub):
		return newGitHubCommenterFromEnv()
	case string(ProviderBitbucket), string(ProviderGitLab):
		return nil, fmt.Errorf("%w: %s", ErrProviderNotImplemented, selected)
	default:
		return nil, fmt.Errorf("unknown git provider %q", selected)
	}
}

func hasGitHubEnv() bool {
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN")) != "" && strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY")) != ""
}
