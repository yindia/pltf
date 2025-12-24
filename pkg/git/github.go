package git

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrNoPRNumber = errors.New("pull request number not found")

type githubCommenter struct {
	token string
	owner string
	repo  string
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

type ghReview struct {
	State             string `json:"state"`
	AuthorAssociation string `json:"author_association"`
	SubmittedAt       string `json:"submitted_at"`
	User              struct {
		Login string `json:"login"`
	} `json:"user"`
}

func newGitHubCommenterFromEnv() (*githubCommenter, error) {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	repoFull := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	if token == "" || repoFull == "" {
		return nil, ErrNoProvider
	}
	owner, repo, err := splitRepo(repoFull)
	if err != nil {
		return nil, err
	}
	return &githubCommenter{token: token, owner: owner, repo: repo}, nil
}

func (c *githubCommenter) UpsertPRComment(comment PRComment) error {
	prNumber, err := detectPRNumber()
	if err != nil {
		return ErrNoPRNumber
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	existingID, found, err := findExistingPRComment(ctx, client, c.token, c.owner, c.repo, prNumber, comment.Marker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: list PR comments failed: %v\n", err)
		return err
	}

	if found {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/comments/%d", c.owner, c.repo, existingID)
		if err := doGitHubRequest(ctx, client, c.token, http.MethodPatch, url, map[string]string{"body": comment.Body}, nil); err != nil {
			fmt.Fprintf(os.Stderr, "warn: edit PR comment failed: %v\n", err)
			return err
		}
		fmt.Fprintf(os.Stderr, "info: updated PR comment %d\n", existingID)
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", c.owner, c.repo, prNumber)
	if err := doGitHubRequest(ctx, client, c.token, http.MethodPost, url, map[string]string{"body": comment.Body}, nil); err != nil {
		fmt.Fprintf(os.Stderr, "warn: create PR comment failed: %v\n", err)
		return err
	}
	fmt.Fprintln(os.Stderr, "info: created PR comment")
	return nil
}

func (c *githubCommenter) GetReviewSummary() (ReviewSummary, error) {
	prNumber, err := detectPRNumber()
	if err != nil {
		return ReviewSummary{}, ErrNoPRNumber
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews?per_page=100", c.owner, c.repo, prNumber)
	var reviews []ghReview
	if err := doGitHubRequest(ctx, client, c.token, http.MethodGet, url, nil, &reviews); err != nil {
		return ReviewSummary{}, err
	}

	type reviewEntry struct {
		review ghReview
		at     time.Time
	}
	latest := make(map[string]reviewEntry)
	for _, r := range reviews {
		login := strings.TrimSpace(r.User.Login)
		if login == "" {
			continue
		}
		t := parseReviewTime(r.SubmittedAt)
		if prev, ok := latest[login]; ok {
			if prev.at.After(t) {
				continue
			}
		}
		latest[login] = reviewEntry{review: r, at: t}
	}

	names := make([]string, 0, len(latest))
	for name := range latest {
		names = append(names, name)
	}
	sort.Strings(names)

	summary := ReviewSummary{}
	for _, name := range names {
		r := latest[name].review
		state := strings.ToUpper(strings.TrimSpace(r.State))
		if state == "APPROVED" {
			summary.Approvals++
		}
		summary.Reviews = append(summary.Reviews, ReviewStatus{
			Name:   name,
			Team:   strings.TrimSpace(r.AuthorAssociation),
			Status: state,
		})
	}
	return summary, nil
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

	return 0, ErrNoPRNumber
}

func findExistingPRComment(ctx context.Context, client *http.Client, token, owner, repo string, prNumber int, marker string) (int64, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments?per_page=100", owner, repo, prNumber)
	var comments []ghComment
	if err := doGitHubRequest(ctx, client, token, http.MethodGet, url, nil, &comments); err != nil {
		return 0, false, err
	}
	if marker == "" {
		return 0, false, nil
	}
	for _, c := range comments {
		if strings.Contains(c.Body, marker) {
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
		fmt.Fprintf(os.Stderr, "warn: github request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		fmt.Fprintf(os.Stderr, "warn: github API %s %s failed: %s - %s\n", method, url, resp.Status, string(b))
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

func parseReviewTime(s string) time.Time {
	if strings.TrimSpace(s) == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
