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

type GeminiProvider struct {
	APIKey string
	Model  string
}

func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
		Model:  "gemini-3.1-flash-lite",
	}
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *GeminiProvider) Complete(prompt string) Output {
	start := time.Now()

	if g.APIKey == "" {
		return Output{Provider: g.Name(), Error: fmt.Errorf("GOOGLE_API_KEY not set")}
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.Model, g.APIKey,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return Output{Provider: g.Name(), Error: err}
	}
	req.Header.Set("Content-Type", "application/json")

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

	var parsed geminiResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return Output{Provider: g.Name(), Error: fmt.Errorf("failed to parse response: %v", err)}
	}

	if parsed.Error != nil {
		return Output{Provider: g.Name(), Error: fmt.Errorf("gemini API error: %s", parsed.Error.Message)}
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return Output{Provider: g.Name(), Error: fmt.Errorf("empty response from gemini")}
	}

	latency := time.Since(start).Seconds()

	return Output{
		Provider: g.Name(),
		Text:     parsed.Candidates[0].Content.Parts[0].Text,
		Latency:  latency,
	}
}