package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type aiRequest struct {
	Model       string  `json:"model"`
	Messages    []aiMsg `json:"messages"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type aiMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiResponse struct {
	Choices []struct {
		Message aiMsg `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func maybeAICritique(run tfRunSummary) string {
	key := getAIKey()
	if strings.TrimSpace(key) == "" {
		return ""
	}
	if run.Plan == nil || strings.ToLower(run.Action) != "plan" {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

	reqBody := aiRequest{
		Model: "gpt-4o-mini",
		Messages: []aiMsg{
			{
				Role:    "system",
				Content: "You are a senior cloud/Terraform reviewer. Analyze the plan diff and produce a concise risk assessment for a PR comment. Focus on: destructive changes + blast radius; IAM, network exposure, and data plane changes; data loss/downtime risks; state/backend/provider drift; unsafe defaults. Output format: 3 sections with bullets: `Blockers`, `Cautions`, `Notes`. If none for a section, write `- none`. Keep to <120 words. Do not restate the plan; only risks and required actions.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Plan summary: %s\nDetails:%s", summary, b.String()),
			},
		},
		MaxTokens:   300,
		Temperature: 0.2,
	}

	buf, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return ""
	}
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var aiResp aiResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return ""
	}
	if aiResp.Error != nil {
		return ""
	}
	if len(aiResp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(aiResp.Choices[0].Message.Content)
}

func getAIKey() string {
	if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("OPEN_AI_KEY"))
}
