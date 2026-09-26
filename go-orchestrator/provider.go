package main

// Output represents a single response from a provider for a single prompt.
type Output struct {
	Provider string
	Text     string
	Latency  float64
	Error    error
}

// Provider is the interface every LLM adapter (Claude, GPT, Gemini, Ollama)
// must implement. This lets the orchestrator call any provider the same way.
type Provider interface {
	Complete(prompt string) Output
	Name() string
}