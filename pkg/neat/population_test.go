package neat

import "testing"

func TestPopulationSpeciate(t *testing.T) {
	rng := NewRand(10)
	cfg := DefaultPopulationConfig()
	cfg.CompatibilityThreshold = 1.0

	g1 := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{{Innovation: 1, In: 1, Out: 2, Weight: 1.0, Enabled: true}},
	}
	g2 := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{{Innovation: 1, In: 1, Out: 2, Weight: 1.1, Enabled: true}},
	}
	g3 := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
			{ID: 3, Kind: NodeHidden, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 1.0, Enabled: true},
			{Innovation: 2, In: 1, Out: 3, Weight: -1.0, Enabled: true},
			{Innovation: 3, In: 3, Out: 2, Weight: 2.0, Enabled: true},
		},
	}

	pop, err := NewPopulation(rng, cfg, []Genome{g1, g2, g3})
	if err != nil {
		t.Fatalf("NewPopulation error: %v", err)
	}

	if err := pop.Speciate(); err != nil {
		t.Fatalf("Speciate error: %v", err)
	}
	if len(pop.Species) != 2 {
		t.Fatalf("expected 2 species, got %d", len(pop.Species))
	}
}
