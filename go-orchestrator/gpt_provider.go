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

type GPTProvider struct {
	APIKey string
	Model  string
}

func NewGPTProvider() *GPTProvider {
	return &GPTProvider{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	}
}

func (g *GPTProvider) Name() string {
	return "gpt"
}

type gptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type gptRequest struct {
	Model    string       `json:"model"`
	Messages []gptMessage `json:"messages"`
}

type gptChoice struct {
	Message gptMessage `json:"message"`
}

type gptResponse struct {
	Choices []gptChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *GPTProvider) Complete(prompt string) Output {
	start := time.Now()

	if g.APIKey == "" {
		return Output{Provider: g.Name(), Error: fmt.Errorf("OPENAI_API_KEY not set")}
	}

	reqBody := gptRequest{
		Model: g.Model,
		Messages: []gptMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}

	var parsed gptResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return Output{Provider: g.Name(), Error: fmt.Errorf("failed to parse response: %v", err)}
	}

	if parsed.Error != nil {
		return Output{Provider: g.Name(), Error: fmt.Errorf("gpt API error: %s", parsed.Error.Message)}
	}

	if len(parsed.Choices) == 0 {
		return Output{Provider: g.Name(), Error: fmt.Errorf("empty response from gpt")}
	}

	latency := time.Since(start).Seconds()

	return Output{
		Provider: g.Name(),
		Text:     parsed.Choices[0].Message.Content,
		Latency:  latency,
	}
}