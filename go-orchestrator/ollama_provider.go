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

type OllamaProvider struct {
	Host  string
	Model string
}

func NewOllamaProvider() *OllamaProvider {
	return NewOllamaProviderWithModel(os.Getenv("OLLAMA_MODEL"))
}

func NewOllamaProviderWithModel(model string) *OllamaProvider {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1"
	}
	return &OllamaProvider{Host: host, Model: model}
}

func (o *OllamaProvider) Name() string {
	return "ollama:" + o.Model
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (o *OllamaProvider) Complete(prompt string) Output {
	start := time.Now()

	reqBody := ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Output{Provider: o.Name(), Error: err}
	}

	url := o.Host + "/api/generate"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return Output{Provider: o.Name(), Error: err}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Output{Provider: o.Name(), Error: fmt.Errorf("could not reach Ollama at %s (is it running?): %v", o.Host, err)}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Output{Provider: o.Name(), Error: err}
	}

	var parsed ollamaResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return Output{Provider: o.Name(), Error: fmt.Errorf("failed to parse Ollama response: %v", err)}
	}

	if parsed.Error != "" {
		return Output{Provider: o.Name(), Error: fmt.Errorf("ollama error: %s", parsed.Error)}
	}

	latency := time.Since(start).Seconds()

	return Output{
		Provider: o.Name(),
		Text:     parsed.Response,
		Latency:  latency,
	}
}