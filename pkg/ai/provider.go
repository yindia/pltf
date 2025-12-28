package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"pltf/pkg/ai/openai"
)

// Provider is the interface for AI providers that can critique Terraform plans.
type Provider interface {
	Critique(ctx context.Context, planSummary string, planDetails string) (string, error)
}

// FromConfig returns an AI provider based on environment configuration.
// It defaults to OpenAI if no provider is specified.
func FromConfig() (Provider, error) {
	provider := strings.ToLower(os.Getenv("PLTF_AI_PROVIDER"))
	if provider == "" || provider == "openai" {
		return openai.New()
	}
	return nil, fmt.Errorf("unsupported AI provider: %q", provider)
}

