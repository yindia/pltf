package clihelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquasecurity/defsec/pkg/extrafs"
	tfscanner "github.com/aquasecurity/defsec/pkg/scanners/terraform"
)

type TfsecFinding struct {
	Severity    string
	Rule        string
	Location    string
	Description string
	Impact      string
	Resolution  string
	Links       []string
	Snippet     string
}

type TfsecSummary struct {
	ExitCode int
	Failed   int
	Low      int
	Medium   int
	High     int
	Critical int
	Findings []TfsecFinding
	Report   string
	Timings  struct {
		DiskIO     time.Duration
		Parsing    time.Duration
		Adaptation time.Duration
		Checks     time.Duration
		Total      time.Duration
	}
	Counts struct {
		ModulesDownloaded int
		ModulesProcessed  int
		BlocksProcessed   int
		FilesRead         int
		Passed            int
		Ignored           int
	}
}

type CostSummary struct {
	TotalMonthly string
	Breakdown    string
	Raw          string
}

func FormatTfsecInsights(summary *TfsecSummary) string {
	var b strings.Builder
	b.WriteString("  timings\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  disk i/o             %s\n", formatDurationMs(summary.Timings.DiskIO))
	fmt.Fprintf(&b, "  parsing              %s\n", formatDurationMs(summary.Timings.Parsing))
	fmt.Fprintf(&b, "  adaptation           %s\n", formatDurationMs(summary.Timings.Adaptation))
	fmt.Fprintf(&b, "  checks               %s\n", formatDurationMs(summary.Timings.Checks))
	fmt.Fprintf(&b, "  total                %s\n\n", formatDurationMs(summary.Timings.Total))

	b.WriteString("  counts\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  modules downloaded   %d\n", summary.Counts.ModulesDownloaded)
	fmt.Fprintf(&b, "  modules processed    %d\n", summary.Counts.ModulesProcessed)
	fmt.Fprintf(&b, "  blocks processed     %d\n", summary.Counts.BlocksProcessed)
	fmt.Fprintf(&b, "  files read           %d\n\n", summary.Counts.FilesRead)

	b.WriteString("  results\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  passed               %d\n", summary.Counts.Passed)
	fmt.Fprintf(&b, "  ignored              %d\n", summary.Counts.Ignored)
	fmt.Fprintf(&b, "  critical             %d\n", summary.Critical)
	fmt.Fprintf(&b, "  high                 %d\n", summary.High)
	fmt.Fprintf(&b, "  medium               %d\n", summary.Medium)
	fmt.Fprintf(&b, "  low                  %d\n\n", summary.Low)

	totalProblems := summary.Failed
	fmt.Fprintf(&b, "  %d passed, %d ignored, %d potential problem(s) detected.\n", summary.Counts.Passed, summary.Counts.Ignored, totalProblems)
	return b.String()
}

func FormatTfsecReport(summary *TfsecSummary) string {
	var b strings.Builder
	for i, f := range summary.Findings {
		fmt.Fprintf(&b, "Result #%d %s %s\n", i+1, strings.ToUpper(f.Severity), f.Description)
		b.WriteString("────────────────────────────────────────────────────────────────────────────────\n")
		if f.Location != "" {
			fmt.Fprintf(&b, "  %s\n", f.Location)
			b.WriteString("────────────────────────────────────────────────────────────────────────────────\n")
		}
		if f.Snippet != "" {
			b.WriteString(f.Snippet)
			if !strings.HasSuffix(f.Snippet, "\n") {
				b.WriteString("\n")
			}
		}
		if f.Rule != "" {
			fmt.Fprintf(&b, "          ID %s\n", f.Rule)
		}
		if f.Impact != "" {
			fmt.Fprintf(&b, "      Impact %s\n", f.Impact)
		}
		if f.Resolution != "" {
			fmt.Fprintf(&b, "  Resolution %s\n", f.Resolution)
		}
		if len(f.Links) > 0 {
			b.WriteString("\n  More Information\n")
			for _, link := range f.Links {
				fmt.Fprintf(&b, "  - %s\n", link)
			}
		}
		b.WriteString("────────────────────────────────────────────────────────────────────────────────\n\n")
	}
	b.WriteString(FormatTfsecInsights(summary))
	return b.String()
}

func FormatTfsecInsightsForComment(s *TfsecSummary) string {
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

	if s.Report != "" && !strings.HasSuffix(s.Report, "\n") {
		b.WriteString(s.Report)
		if !strings.HasSuffix(s.Report, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func RunTfsecScan(dir string) (*TfsecSummary, error) {
	root := filepath.Clean(dir)
	scnr := tfscanner.New()
	results, metrics, err := scnr.ScanFSWithMetrics(context.Background(), extrafs.OSDir(root), ".")
	if err != nil {
		return nil, err
	}
	exit := tfsecExitCode(metrics)
	summary := &TfsecSummary{
		ExitCode: exit,
		Failed:   metrics.Executor.Counts.Failed,
		Low:      metrics.Executor.Counts.Low,
		Medium:   metrics.Executor.Counts.Medium,
		High:     metrics.Executor.Counts.High,
		Critical: metrics.Executor.Counts.Critical,
	}
	summary.Counts.ModulesDownloaded = metrics.Parser.Counts.ModuleDownloads
	summary.Counts.ModulesProcessed = metrics.Parser.Counts.Modules
	summary.Counts.BlocksProcessed = metrics.Parser.Counts.Blocks
	summary.Counts.FilesRead = metrics.Parser.Counts.Files
	summary.Counts.Passed = metrics.Executor.Counts.Passed
	summary.Counts.Ignored = metrics.Executor.Counts.Ignored
	summary.Timings.DiskIO = metrics.Parser.Timings.DiskIODuration
	summary.Timings.Parsing = metrics.Parser.Timings.ParseDuration
	summary.Timings.Adaptation = metrics.Executor.Timings.Adaptation
	summary.Timings.Checks = metrics.Executor.Timings.RunningChecks
	summary.Timings.Total = metrics.Timings.Total
	for _, res := range results {
		status := strings.ToLower(fmt.Sprint(res.Status()))
		if !strings.Contains(status, "fail") {
			continue
		}
		f := TfsecFinding{
			Severity:    string(res.Severity()),
			Rule:        res.Rule().LongID(),
			Location:    res.Range().String(),
			Description: res.Description(),
			Impact:      res.Rule().Impact,
			Resolution:  res.Rule().Resolution,
			Links:       res.Rule().Links,
		}
		summary.Findings = append(summary.Findings, f)
	}
	summary.Report = FormatTfsecReport(summary)

	if exit != 0 {
		fmt.Fprintf(os.Stderr, "warn: tfsec reported issues (exit=%d, failed=%d low=%d medium=%d high=%d critical=%d)\n",
			exit, summary.Failed, summary.Low, summary.Medium, summary.High, summary.Critical)
	}
	fmt.Fprint(os.Stderr, summary.Report)
	return summary, nil
}

func tfsecExitCode(metrics tfscanner.Metrics) int {
	if metrics.Executor.Counts.Failed == 0 {
		return 0
	}
	if metrics.Executor.Counts.Failed == metrics.Executor.Counts.Low {
		return 2
	}
	return 1
}

func formatDurationMs(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)
	return fmt.Sprintf("%.6fms", ms)
}
