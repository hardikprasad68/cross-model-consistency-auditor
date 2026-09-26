package main

import "fmt"

func main() {
	cases, err := LoadAllPromptCases("../prompts")
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("Loaded", len(cases), "prompt cases from ../prompts")
	fmt.Println()

	providers := BuildLocalProviders()

	for _, pc := range cases {
		fmt.Println("############################################")
		fmt.Println("PROMPT:", pc.ID, "| category:", pc.Category)
		fmt.Println("############################################")

		renderedPrompt := pc.Render()
		outputs := RunAll(providers, renderedPrompt)

		for _, out := range outputs {
			fmt.Println("--- " + out.Provider + " ---")
			if out.Error != nil {
				fmt.Println("ERROR:", out.Error)
			} else {
				fmt.Printf("Latency: %.2fs | Text: %s\n", out.Latency, out.Text)
			}
		}

		scores, err := ScoreOutputs("http://localhost:8000/score", outputs, pc.ExpectedFormat)
		if err != nil {
			fmt.Println("SCORING ERROR:", err)
			continue
		}

		classification := Classify(scores)
		fmt.Printf("\nSemantic: %.3f | Structural: %.3f | Behavioral: %.3f\n", scores.Semantic, scores.Structural, scores.Behavioral)
		fmt.Println("Classification:", classification)

		err = SaveRun("../data/runs.jsonl", pc.ID, pc.Category, outputs, scores, classification)
		if err != nil {
			fmt.Println("SAVE ERROR:", err)
		} else {
			fmt.Println("Saved.")
		}
		fmt.Println()
	}

	fmt.Println("Full suite complete.")
}