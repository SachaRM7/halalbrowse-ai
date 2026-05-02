package ml

import (
	"path/filepath"
	"strings"
)

// ImageScorer uses file names/URLs as offline-friendly image hints.
type ImageScorer struct {
	haramHints map[string]float64
	halalHints map[string]float64
}

func NewImageScorer() ImageScorer {
	return ImageScorer{
		haramHints: map[string]float64{
			"adult":    0.30,
			"casino":   0.20,
			"gambling": 0.22,
			"lingerie": 0.30,
			"beer":     0.18,
		},
		halalHints: map[string]float64{
			"quran":    0.20,
			"masjid":   0.16,
			"family":   0.08,
			"children": 0.06,
		},
	}
}

func (s ImageScorer) Score(identifier string) Score {
	normalized := normalizeText(filepath.Base(identifier))
	confidence := 0.76
	reasons := []string{}
	signals := map[string]float64{}

	for keyword, weight := range s.haramHints {
		if strings.Contains(normalized, keyword) {
			confidence -= weight
			reasons = append(reasons, "matched risky image hint: "+keyword)
			signals[keyword] = -weight
		}
	}
	for keyword, weight := range s.halalHints {
		if strings.Contains(normalized, keyword) {
			confidence += weight
			reasons = append(reasons, "matched positive image hint: "+keyword)
			signals[keyword] = weight
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "neutral filename heuristics")
	}
	return newScore(confidence, reasons, signals)
}
