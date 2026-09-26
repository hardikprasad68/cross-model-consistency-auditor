package main

func Classify(scores DivergenceScores) string {
	switch {
	case scores.Structural > 0.3:
		return "structurally_fragile"
	case scores.Behavioral > 0.25:
		return "behaviorally_fragile"
	case scores.Semantic > 0.4:
		return "cosmetically_fragile"
	default:
		return "portable"
	}
}