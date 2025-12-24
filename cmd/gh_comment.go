package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"pltf/pkg/git"
)

const prCommentMarker = "<!-- pltf:terraform-run -->"

type tfRunSummary struct {
	Action string
	Status string
	Spec   string
	Env    string
	OutDir string
	Err    string
	Plan   *planSummary
	AI     string
	Scan   *tfsecSummary
	Cost   *costSummary
}

func maybeUpsertPRComment(run tfRunSummary) error {
	body := buildPRCommentBody(run)
	commenter, err := git.NewCommenter("")
	if err != nil {
		if errors.Is(err, git.ErrNoProvider) {
			fmt.Fprintln(os.Stderr, "info: skipping PR comment, missing git provider credentials")
			return nil
		}
		if errors.Is(err, git.ErrProviderNotImplemented) {
			fmt.Fprintf(os.Stderr, "info: skipping PR comment, provider not implemented: %v\n", err)
			return nil
		}
		return err
	}

	if err := commenter.UpsertPRComment(git.PRComment{Body: body, Marker: prCommentMarker}); err != nil {
		if errors.Is(err, git.ErrNoPRNumber) {
			return nil
		}
		return err
	}
	return nil
}

func buildPRCommentBody(run tfRunSummary) string {
	statusEmoji := "✅"
	statusText := run.Status
	if strings.ToLower(run.Status) != "succeeded" {
		statusEmoji = "❌"
		if strings.TrimSpace(statusText) == "" {
			statusText = "failed"
		}
	}

	var sb strings.Builder
	sb.WriteString(prCommentMarker)
	sb.WriteString("\n\n")
	sb.WriteString("### Terrateam Plan Output\n\n")
	sb.WriteString(fmt.Sprintf("**%s** %s\n\n", run.Spec, statusEmoji+" "+titleCase(statusText)))

	sb.WriteString("<details><summary>Expand for plan output details</summary>\n\n")
	sb.WriteString("```\n")
	if run.Plan != nil && strings.TrimSpace(run.Plan.Text) != "" {
		sb.WriteString(run.Plan.Text)
		if !strings.HasSuffix(run.Plan.Text, "\n") {
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("Plan output unavailable.\n")
	}
	sb.WriteString("```\n")
	if run.Plan != nil {
		sb.WriteString(fmt.Sprintf("\nPlan: %d to add, %d to change, %d to destroy\n",
			run.Plan.Added, run.Plan.Changed, run.Plan.Destroyed))
	}
	if strings.TrimSpace(run.Err) != "" {
		sb.WriteString(fmt.Sprintf("\nError: %s\n", truncateForComment(run.Err)))
	}
	sb.WriteString("\n</details>\n")

	if run.Cost != nil {
		sb.WriteString("\n---\n\n")
		sb.WriteString("**Cost Estimation**\n\n")
		if run.Cost.TotalMonthly != "" {
			sb.WriteString(fmt.Sprintf("Total Monthly Difference: %s\n\n", run.Cost.TotalMonthly))
		}
		if strings.TrimSpace(run.Cost.Breakdown) != "" {
			sb.WriteString("<details><summary>Expand for cost estimation details</summary>\n\n")
			sb.WriteString("```\n")
			sb.WriteString(run.Cost.Breakdown)
			if !strings.HasSuffix(run.Cost.Breakdown, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("```\n</details>\n")
		}
		if strings.TrimSpace(run.Cost.Raw) != "" {
			raw := truncateForComment(run.Cost.Raw)
			sb.WriteString("\n<details><summary>Raw cost data (json)</summary>\n\n")
			sb.WriteString("```\n")
			sb.WriteString(raw)
			if !strings.HasSuffix(raw, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("```\n</details>\n")
		}
	}

	if strings.TrimSpace(run.AI) != "" || run.Scan != nil {
		sb.WriteString("\n---\n\n")
		sb.WriteString("**Approval Requirements**\n\n")
		if strings.TrimSpace(run.AI) != "" {
			sb.WriteString("AI risk review:\n")
			sb.WriteString(run.AI)
			if !strings.HasSuffix(run.AI, "\n") {
				sb.WriteString("\n")
			}
		}
		if run.Scan != nil {
			sb.WriteString("\n<details><summary>Security scan (tfsec)</summary>\n\n")
			sb.WriteString("```\n")
			sb.WriteString(formatTfsecInsightsForComment(run.Scan))
			sb.WriteString("```\n</details>\n")
		}
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("To apply all these changes, comment:\n\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("pltf terraform apply -f %s --auto-approve", run.Spec))
	if strings.TrimSpace(run.Env) != "" {
		sb.WriteString(fmt.Sprintf(" --env %s", run.Env))
	}
	if run.Plan != nil && len(run.Plan.RawPlanArgs) > 0 {
		for _, a := range run.Plan.RawPlanArgs {
			sb.WriteString(" ")
			sb.WriteString(a)
		}
	}
	sb.WriteString("\n```\n")
	sb.WriteString("\n_This comment updates automatically on pushes to the PR._\n")
	return sb.String()
}

func truncateForComment(s string) string {
	const max = 4000
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func formatTfsecInsightsForComment(s *tfsecSummary) string {
	if strings.TrimSpace(s.Report) != "" {
		return s.Report
	}
	var b strings.Builder
	b.WriteString("timings\n")
	b.WriteString("──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "disk i/o             %s\n", formatDurationMs(s.Timings.DiskIO))
	fmt.Fprintf(&b, "parsing              %s\n", formatDurationMs(s.Timings.Parsing))
	fmt.Fprintf(&b, "adaptation           %s\n", formatDurationMs(s.Timings.Adaptation))
	fmt.Fprintf(&b, "checks               %s\n", formatDurationMs(s.Timings.Checks))
	fmt.Fprintf(&b, "total                %s\n\n", formatDurationMs(s.Timings.Total))

	b.WriteString("counts\n")
	b.WriteString("──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "modules downloaded   %d\n", s.Counts.ModulesDownloaded)
	fmt.Fprintf(&b, "modules processed    %d\n", s.Counts.ModulesProcessed)
	fmt.Fprintf(&b, "blocks processed     %d\n", s.Counts.BlocksProcessed)
	fmt.Fprintf(&b, "files read           %d\n\n", s.Counts.FilesRead)

	b.WriteString("results\n")
	b.WriteString("──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "passed               %d\n", s.Counts.Passed)
	fmt.Fprintf(&b, "ignored              %d\n", s.Counts.Ignored)
	fmt.Fprintf(&b, "critical             %d\n", s.Critical)
	fmt.Fprintf(&b, "high                 %d\n", s.High)
	fmt.Fprintf(&b, "medium               %d\n", s.Medium)
	fmt.Fprintf(&b, "low                  %d\n\n", s.Low)

	totalProblems := s.Failed
	fmt.Fprintf(&b, "%d passed, %d ignored, %d potential problem(s) detected.\n", s.Counts.Passed, s.Counts.Ignored, totalProblems)
	return b.String()
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
