package config

import "strings"

// GitProvider enumerates supported git providers.
type GitProvider string

const (
	GitProviderGitHub    GitProvider = "github"
	GitProviderGitLab    GitProvider = "gitlab"
	GitProviderBitbucket GitProvider = "bitbucket"
)

func normalizeGitProvider(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
