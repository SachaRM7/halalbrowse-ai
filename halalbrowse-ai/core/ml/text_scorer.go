package ml

import "strings"

// TextScorer is a deterministic local scorer standing in for a quantized text model.
type TextScorer struct {
	haramWeights map[string]float64
	halalWeights map[string]float64
}

func NewTextScorer() TextScorer {
	return TextScorer{
		haramWeights: map[string]float64{
			"adult":    0.28,
			"casino":   0.25,
			"gambling": 0.22,
			"porn":     0.35,
			"explicit": 0.18,
			"nude":     0.30,
			"bet":      0.16,
			"alcohol":  0.16,
		},
		halalWeights: map[string]float64{
			"quran":   0.20,
			"tajweed": 0.18,
			"masjid":  0.12,
			"hadith":  0.16,
			"salah":   0.16,
			"islamic": 0.10,
		},
	}
}

func (s TextScorer) Score(text string) Score {
	normalized := normalizeText(text)
	if normalized == "" {
		return newScore(0.5, []string{"empty input"}, map[string]float64{"baseline": 0.5})
	}

	confidence := 0.78
	reasons := []string{}
	signals := map[string]float64{}

	for keyword, weight := range s.haramWeights {
		if strings.Contains(normalized, keyword) {
			confidence -= weight
			reasons = append(reasons, "matched haram keyword: "+keyword)
			signals[keyword] = -weight
		}
	}
	for keyword, weight := range s.halalWeights {
		if strings.Contains(normalized, keyword) {
			confidence += weight
			reasons = append(reasons, "matched halal keyword: "+keyword)
			signals[keyword] = weight
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "no strong content signals detected")
	}
	return newScore(confidence, reasons, signals)
}
