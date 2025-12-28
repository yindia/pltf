package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// OpenAI is an AI provider that uses OpenAI's API.
type OpenAI struct {
	APIKey   string
	Model    string
	Endpoint string
}

// New creates a new OpenAI provider.
func New() (*OpenAI, error) {
	key := getAPIKey()
	if key == "" {
		return nil, fmt.Errorf("OpenAI API key not found. Set OPENAI_API_KEY or PLTF_AI_API_KEY")
	}

	model := os.Getenv("PLTF_AI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	endpoint := os.Getenv("PLTF_AI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	return &OpenAI{
		APIKey:   key,
		Model:    model,
		Endpoint: endpoint,
	},
}

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

// Critique sends the plan to OpenAI for a critique.
func (o *OpenAI) Critique(ctx context.Context, planSummary string, planDetails string) (string, error) {
	reqBody := aiRequest{
		Model: o.Model,
		Messages: []aiMsg{
			{
				Role:    "system",
				Content: "You are a senior cloud/Terraform reviewer. Analyze the plan diff and produce a concise risk assessment for a PR comment. Focus on: destructive changes + blast radius; IAM, network exposure, and data plane changes; data loss/downtime risks; state/backend/provider drift; unsafe defaults. Output format: 3 sections with bullets: `Blockers`, `Cautions`, `Notes`. If none for a section, write `- none`. Keep to <120 words. Do not restate the plan; only risks and required actions.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Plan summary: %s\nDetails:%s", planSummary, planDetails),
			},
		},
		MaxTokens:   300,
		Temperature: 0.2,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal AI request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to create AI request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	var aiResp aiResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return "", fmt.Errorf("failed to decode AI response: %w", err)
	}
	if aiResp.Error != nil {
		return "", fmt.Errorf("AI API error: %s", aiResp.Error.Message)
	}
	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI")
	}
	return strings.TrimSpace(aiResp.Choices[0].Message.Content), nil
}

func getAPIKey() string {
	if v := strings.TrimSpace(os.Getenv("PLTF_AI_API_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("OPEN_AI_KEY"))
}
