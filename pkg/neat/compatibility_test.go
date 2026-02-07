package neat

import (
	"math"
	"testing"
)

func TestCompatibilityDistance(t *testing.T) {
	cfg := DefaultDistanceConfig()

	a := Genome{
		Connections: []ConnectionGene{
			{Innovation: 1, Weight: 1.0},
			{Innovation: 2, Weight: 1.0},
			{Innovation: 3, Weight: 1.0},
		},
	}
	b := Genome{
		Connections: []ConnectionGene{
			{Innovation: 1, Weight: 1.0},
			{Innovation: 2, Weight: 2.0},
			{Innovation: 4, Weight: 3.0},
			{Innovation: 5, Weight: 4.0},
		},
	}

	distance := CompatibilityDistance(a, b, cfg)
	want := 3.2
	if math.Abs(distance-want) > 1e-9 {
		t.Fatalf("expected distance %.2f, got %.5f", want, distance)
	}
}
