package cppn

import "math"

// FitnessWeights controls how metrics are combined into a single score.
type FitnessWeights struct {
	Entropy         float64
	EdgeDensity     float64
	FineEdges       float64
	Variance        float64
	Symmetry        float64
	ColorVar        float64
	HighFreqPenalty float64
}

// DefaultFitnessWeights provides a balanced exploratory mix.
func DefaultFitnessWeights() FitnessWeights {
	return FitnessWeights{
		Entropy:         0.2,
		EdgeDensity:     0.35,
		FineEdges:       0.15,
		Variance:        0.2,
		Symmetry:        0.1,
		ColorVar:        0.15,
		HighFreqPenalty: 0.25,
	}
}

// ScoreFromMetrics converts metrics into a fitness score.
func ScoreFromMetrics(m Metrics, weights FitnessWeights, color bool) float64 {
	entropyScore := clamp01(m.Entropy / 8.0)
	varianceScore := clamp01(m.Variance / 0.25)
	edgeScore := targetScore(m.EdgeDensity, 0.18, 0.18)
	fineEdgeScore := targetScore(m.FineEdges, 0.35, 0.25)
	symScore := clamp01((m.SymmetryX + m.SymmetryY) * 0.5)
	colorScore := clamp01(m.ColorVar / 0.25)

	hfNorm := clamp01(m.HighFreq / 1.0)
	hfPenalty := clamp01((hfNorm - 0.35) / 0.65)

	score := weights.Entropy*entropyScore +
		weights.EdgeDensity*edgeScore +
		weights.FineEdges*fineEdgeScore +
		weights.Variance*varianceScore +
		weights.Symmetry*symScore
	if color {
		score += weights.ColorVar * colorScore
	}
	score -= weights.HighFreqPenalty * hfPenalty

	if score < 0 {
		return 0
	}
	return score
}

func targetScore(value, target, tolerance float64) float64 {
	if tolerance <= 0 {
		return 0
	}
	diff := math.Abs(value - target)
	score := 1 - diff/tolerance
	return clamp01(score)
}
