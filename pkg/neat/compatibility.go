package neat

import "sort"

// DistanceConfig controls compatibility distance calculation.
type DistanceConfig struct {
	ExcessCoeff            float64
	DisjointCoeff          float64
	WeightCoeff            float64
	NormalizationThreshold int
}

// DefaultDistanceConfig returns common NEAT coefficients.
func DefaultDistanceConfig() DistanceConfig {
	return DistanceConfig{
		ExcessCoeff:            1.0,
		DisjointCoeff:          1.0,
		WeightCoeff:            0.4,
		NormalizationThreshold: 20,
	}
}

// CompatibilityDistance computes the NEAT distance between two genomes.
func CompatibilityDistance(a, b Genome, cfg DistanceConfig) float64 {
	aGenes := sortedConnections(a.Connections)
	bGenes := sortedConnections(b.Connections)

	var disjoint, excess int
	var weightDiff float64
	var matching int

	i, j := 0, 0
	for i < len(aGenes) && j < len(bGenes) {
		ai := aGenes[i]
		bj := bGenes[j]
		switch {
		case ai.Innovation == bj.Innovation:
			matching++
			if ai.Weight > bj.Weight {
				weightDiff += ai.Weight - bj.Weight
			} else {
				weightDiff += bj.Weight - ai.Weight
			}
			i++
			j++
		case ai.Innovation < bj.Innovation:
			disjoint++
			i++
		default:
			disjoint++
			j++
		}
	}

	excess += len(aGenes) - i
	excess += len(bGenes) - j

	avgWeightDiff := 0.0
	if matching > 0 {
		avgWeightDiff = weightDiff / float64(matching)
	}

	N := maxInt(len(aGenes), len(bGenes))
	if N < cfg.NormalizationThreshold {
		N = 1
	}

	return cfg.ExcessCoeff*float64(excess)/float64(N) +
		cfg.DisjointCoeff*float64(disjoint)/float64(N) +
		cfg.WeightCoeff*avgWeightDiff
}

func sortedConnections(conns []ConnectionGene) []ConnectionGene {
	if len(conns) == 0 {
		return nil
	}
	cpy := make([]ConnectionGene, len(conns))
	copy(cpy, conns)
	sort.Slice(cpy, func(i, j int) bool { return cpy[i].Innovation < cpy[j].Innovation })
	return cpy
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
