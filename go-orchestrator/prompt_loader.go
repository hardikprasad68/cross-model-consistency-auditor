package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PromptCase struct {
	ID             string            `yaml:"id"`
	Category       string            `yaml:"category"`
	PromptTemplate string            `yaml:"prompt_template"`
	Variables      map[string]string `yaml:"variables"`
	ExpectedFormat string            `yaml:"expected_format"`
	Tags           []string          `yaml:"tags"`
}

func LoadPromptCase(path string) (PromptCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PromptCase{}, fmt.Errorf("could not read %s: %v", path, err)
	}

	var pc PromptCase
	if err := yaml.Unmarshal(data, &pc); err != nil {
		return PromptCase{}, fmt.Errorf("could not parse YAML in %s: %v", path, err)
	}

	if pc.ID == "" {
		return PromptCase{}, fmt.Errorf("%s is missing required field 'id'", path)
	}
	if pc.PromptTemplate == "" {
		return PromptCase{}, fmt.Errorf("%s is missing required field 'prompt_template'", path)
	}

	return pc, nil
}

func LoadAllPromptCases(dir string) ([]PromptCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read prompts directory %s: %v", dir, err)
	}

	var cases []PromptCase
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		pc, err := LoadPromptCase(path)
		if err != nil {
			fmt.Println("WARNING: skipping", path, "-", err)
			continue
		}
		cases = append(cases, pc)
	}
	return cases, nil
}

func (pc PromptCase) Render() string {
	result := pc.PromptTemplate
	for key, value := range pc.Variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return strings.TrimSpace(result)
}