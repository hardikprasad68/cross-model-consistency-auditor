package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ClaudeProvider struct {
	APIKey string
	Model  string
}

func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-sonnet-4-6",
	}
}

func (c *ClaudeProvider) Name() string {
	return "claude"
}

type claudeRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string                  `json:"model"`
	MaxTokens int                     `json:"max_tokens"`
	Messages  []claudeRequestMessage  `json:"messages"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeResponse struct {
	Content []claudeContentBlock `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *ClaudeProvider) Complete(prompt string) Output {
	start := time.Now()

	if c.APIKey == "" {
		return Output{Provider: c.Name(), Error: fmt.Errorf("ANTHROPIC_API_KEY not set")}
	}

	reqBody := claudeRequest{
		Model:     c.Model,
		MaxTokens: 1024,
		Messages: []claudeRequestMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Output{Provider: c.Name(), Error: err}
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return Output{Provider: c.Name(), Error: err}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Output{Provider: c.Name(), Error: err}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Output{Provider: c.Name(), Error: err}
	}

	var parsed claudeResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return Output{Provider: c.Name(), Error: fmt.Errorf("failed to parse response: %v", err)}
	}

	if parsed.Error != nil {
		return Output{Provider: c.Name(), Error: fmt.Errorf("claude API error: %s", parsed.Error.Message)}
	}

	if len(parsed.Content) == 0 {
		return Output{Provider: c.Name(), Error: fmt.Errorf("empty response from claude")}
	}

	latency := time.Since(start).Seconds()

	return Output{
		Provider: c.Name(),
		Text:     parsed.Content[0].Text,
		Latency:  latency,
	}
}