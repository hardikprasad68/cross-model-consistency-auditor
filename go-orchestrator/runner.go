package main

import "fmt"

var LocalModels = []string{
	"llama3.1",
	"llama3.2",
	"mistral",
	"gemma2",
	"qwen2.5",
}

func RunAll(providers []Provider, prompt string) []Output {
	results := make(chan Output, len(providers))

	for _, p := range providers {
		go func(p Provider) {
			defer func() {
				if r := recover(); r != nil {
					results <- Output{Provider: p.Name(), Error: fmt.Errorf("panic: %v", r)}
				}
			}()
			results <- p.Complete(prompt)
		}(p)
	}

	outputs := make([]Output, 0, len(providers))
	for i := 0; i < len(providers); i++ {
		outputs = append(outputs, <-results)
	}
	return outputs
}

func BuildLocalProviders() []Provider {
	providers := make([]Provider, 0, len(LocalModels))
	for _, model := range LocalModels {
		providers = append(providers, NewOllamaProviderWithModel(model))
	}
	return providers
}