package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
}

type ghEvent struct {
	Action      string `json:"action"`
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
}

type ghComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

func maybeUpsertPRComment(run tfRunSummary) error {
	token := os.Getenv("GITHUB_TOKEN")
	repoFull := os.Getenv("GITHUB_REPOSITORY")
	if strings.TrimSpace(token) == "" || strings.TrimSpace(repoFull) == "" {
		return nil
	}

	owner, repo, err := splitRepo(repoFull)
	if err != nil {
		return err
	}

	prNumber, err := detectPRNumber()
	if err != nil {
		// Not a PR event; just skip quietly.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	body := buildPRCommentBody(run)

	existingID, found, err := findExistingPRComment(ctx, client, token, owner, repo, prNumber)
	if err != nil {
		return err
	}

	if found {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/comments/%d", owner, repo, existingID)
		return doGitHubRequest(ctx, client, token, http.MethodPatch, url, map[string]string{"body": body}, nil)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
	return doGitHubRequest(ctx, client, token, http.MethodPost, url, map[string]string{"body": body}, nil)
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
	sb.WriteString(fmt.Sprintf("### Terraform %s %s\n\n", titleCase(run.Action), statusEmoji))
	sb.WriteString(fmt.Sprintf("- spec: `%s`\n", run.Spec))
	if strings.TrimSpace(run.Env) != "" {
		sb.WriteString(fmt.Sprintf("- env: `%s`\n", run.Env))
	}
	if strings.TrimSpace(run.OutDir) != "" {
		sb.WriteString(fmt.Sprintf("- out: `%s`\n", run.OutDir))
	}
	sb.WriteString(fmt.Sprintf("- status: %s\n", statusText))
	if strings.TrimSpace(run.Err) != "" {
		sb.WriteString(fmt.Sprintf("- error: `%s`\n", truncateForComment(run.Err)))
	}
	if run.Plan != nil {
		sb.WriteString("\n<details><summary>Plan summary</summary>\n\n")
		sb.WriteString(fmt.Sprintf("- add: %d\n- change: %d\n- destroy: %d\n\n", run.Plan.Added, run.Plan.Changed, run.Plan.Destroyed))
		if len(run.Plan.Adds) > 0 {
			sb.WriteString("**Add**\n")
			for _, a := range run.Plan.Adds {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
			sb.WriteString("\n")
		}
		if len(run.Plan.Changes) > 0 {
			sb.WriteString("**Change**\n")
			for _, a := range run.Plan.Changes {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
			sb.WriteString("\n")
		}
		if len(run.Plan.Deletes) > 0 {
			sb.WriteString("**Destroy**\n")
			for _, a := range run.Plan.Deletes {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("</details>\n")
	}
	if strings.TrimSpace(run.AI) != "" {
		sb.WriteString("\n<details><summary>AI risk review</summary>\n\n")
		sb.WriteString(run.AI)
		if !strings.HasSuffix(run.AI, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("</details>\n")
	}
	sb.WriteString("\n```\n")
	sb.WriteString(fmt.Sprintf("pltf terraform %s -f %s", run.Action, run.Spec))
	if strings.TrimSpace(run.Env) != "" {
		sb.WriteString(fmt.Sprintf(" --env %s", run.Env))
	}
	sb.WriteString("\n```\n")
	sb.WriteString("\n_This comment updates automatically on pushes to the PR._\n")
	return sb.String()
}

func truncateForComment(s string) string {
	const max = 500
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func detectPRNumber() (int, error) {
	if eventPath := os.Getenv("GITHUB_EVENT_PATH"); eventPath != "" {
		if b, err := os.ReadFile(eventPath); err == nil {
			var ev ghEvent
			if json.Unmarshal(b, &ev) == nil && ev.PullRequest != nil && ev.PullRequest.Number > 0 {
				return ev.PullRequest.Number, nil
			}
		}
	}

	if v := os.Getenv("PR_NUMBER"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n, nil
		}
	}

	return 0, errors.New("pull request number not found")
}

func findExistingPRComment(ctx context.Context, client *http.Client, token, owner, repo string, prNumber int) (int64, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments?per_page=100", owner, repo, prNumber)
	var comments []ghComment
	if err := doGitHubRequest(ctx, client, token, http.MethodGet, url, nil, &comments); err != nil {
		return 0, false, err
	}
	for _, c := range comments {
		if strings.Contains(c.Body, prCommentMarker) {
			return c.ID, true, nil
		}
	}
	return 0, false, nil
}

func doGitHubRequest(ctx context.Context, client *http.Client, token, method, url string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "pltf-cli")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github API %s %s: %s - %s", method, url, resp.Status, string(b))
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func splitRepo(full string) (owner, repo string, err error) {
	parts := strings.Split(full, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid GITHUB_REPOSITORY: %q", full)
	}
	return parts[0], parts[1], nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
