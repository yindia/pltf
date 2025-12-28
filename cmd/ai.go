package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"pltf/pkg/ai"
)

func maybeAICritique(run tfRunSummary) string {
	if run.Plan == nil || strings.ToLower(run.Action) != "plan" {
		return ""
	}

	provider, err := ai.FromConfig()
	if err != nil {
		// This can happen if the provider is unsupported, or if the OpenAI key is missing.
		// In either case, we don't want to block the user, so we just return an empty string.
		// The error is already handled by the provider constructor.
		return ""
	}
	if provider == nil {
		// No provider configured
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary := fmt.Sprintf("adds: %d, changes: %d, destroys: %d", run.Plan.Added, run.Plan.Changed, run.Plan.Destroyed)
	var b strings.Builder
	if len(run.Plan.Adds) > 0 {
		b.WriteString("\nadds:\n")
		for _, a := range run.Plan.Adds {
			b.WriteString("- " + a + "\n")
		}
	}
	if len(run.Plan.Changes) > 0 {
		b.WriteString("\nchanges:\n")
		for _, a := range run.Plan.Changes {
			b.WriteString("- " + a + "\n")
		}
	}
	if len(run.Plan.Deletes) > 0 {
		b.WriteString("\ndestroys:\n")
		for _, a := range run.Plan.Deletes {
			b.WriteString("- " + a + "\n")
		}
	}
	if strings.TrimSpace(run.Plan.Text) != "" {
		b.WriteString("\nplan text:\n")
		b.WriteString(run.Plan.Text)
	}

	critique, err := provider.Critique(ctx, summary, b.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: AI critique failed: %v\n", err)
		return ""
	}
	return critique
}
