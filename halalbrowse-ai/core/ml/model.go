package ml

import "strings"

// Score captures the halal confidence produced by a local heuristic scorer.
type Score struct {
	Confidence float64            `json:"confidence"`
	Label      string             `json:"label"`
	Reasons    []string           `json:"reasons,omitempty"`
	Signals    map[string]float64 `json:"signals,omitempty"`
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func normalizeText(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func newScore(confidence float64, reasons []string, signals map[string]float64) Score {
	label := "halal"
	if confidence < 0.5 {
		label = "non-halal"
	}
	return Score{Confidence: clamp(confidence), Label: label, Reasons: reasons, Signals: signals}
}
