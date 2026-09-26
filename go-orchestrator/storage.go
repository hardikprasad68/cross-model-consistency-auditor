package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func generateRunID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("run_%d", time.Now().UnixNano())
	}
	return "run_" + hex.EncodeToString(b)
}

type RunRecord struct {
	RunID           string  `json:"run_id"`
	PromptID        string  `json:"prompt_id"`
	Category        string  `json:"category"`
	Provider        string  `json:"provider"`
	Output          string  `json:"output"`
	Latency         float64 `json:"latency"`
	SemanticScore   float64 `json:"semantic_score"`
	StructuralScore float64 `json:"structural_score"`
	BehavioralScore float64 `json:"behavioral_score"`
	Classification  string  `json:"classification"`
	Timestamp       string  `json:"timestamp"`
}

func SaveRun(path string, promptID string, category string, outputs []Output, scores DivergenceScores, classification string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open storage file %s: %v", path, err)
	}
	defer f.Close()

	runID := generateRunID()
	timestamp := time.Now().Format(time.RFC3339)

	for _, out := range outputs {
		if out.Error != nil {
			continue
		}
		record := RunRecord{
			RunID:           runID,
			PromptID:        promptID,
			Category:        category,
			Provider:        out.Provider,
			Output:          out.Text,
			Latency:         out.Latency,
			SemanticScore:   scores.Semantic,
			StructuralScore: scores.Structural,
			BehavioralScore: scores.Behavioral,
			Classification:  classification,
			Timestamp:       timestamp,
		}
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}