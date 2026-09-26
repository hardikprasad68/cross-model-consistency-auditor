package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DivergenceScores struct {
	Semantic   float64 `json:"semantic"`
	Structural float64 `json:"structural"`
	Behavioral float64 `json:"behavioral"`
}

type scoreRequest struct {
	Outputs        []string `json:"outputs"`
	ExpectedFormat string   `json:"expected_format"`
}

func ScoreOutputs(scorerURL string, outputs []Output, expectedFormat string) (DivergenceScores, error) {
	texts := make([]string, 0, len(outputs))
	for _, o := range outputs {
		if o.Error == nil {
			texts = append(texts, o.Text)
		}
	}

	if len(texts) < 2 {
		return DivergenceScores{}, fmt.Errorf("need at least 2 successful outputs to score divergence, got %d", len(texts))
	}

	if expectedFormat == "" {
		expectedFormat = "text"
	}

	reqBody := scoreRequest{Outputs: texts, ExpectedFormat: expectedFormat}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return DivergenceScores{}, err
	}

	req, err := http.NewRequest("POST", scorerURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return DivergenceScores{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return DivergenceScores{}, fmt.Errorf("could not reach scoring service at %s (is 'uvicorn main:app --port 8000' running?): %v", scorerURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return DivergenceScores{}, err
	}

	var scores DivergenceScores
	if err := json.Unmarshal(bodyBytes, &scores); err != nil {
		return DivergenceScores{}, fmt.Errorf("failed to parse scorer response: %v (raw: %s)", err, string(bodyBytes))
	}

	return scores, nil
}