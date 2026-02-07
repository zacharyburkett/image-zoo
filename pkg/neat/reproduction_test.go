package neat

import "testing"

func TestNextGenerationMaintainsSizeAndResetsFitness(t *testing.T) {
	rng := NewRand(5)
	pcfg := DefaultPopulationConfig()
	pcfg.CompatibilityThreshold = 10.0

	base := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{{Innovation: 1, In: 1, Out: 2, Weight: 1.0, Enabled: true}},
	}

	genomes := []Genome{
		cloneGenome(base),
		cloneGenome(base),
		cloneGenome(base),
		cloneGenome(base),
	}
	for i := range genomes {
		genomes[i].Fitness = float64(i + 1)
	}

	pop, err := NewPopulation(rng, pcfg, genomes)
	if err != nil {
		t.Fatalf("NewPopulation error: %v", err)
	}

	mcfg := DefaultMutationConfig()
	mcfg.AddConnectionProb = 0
	mcfg.AddNodeProb = 0
	mcfg.WeightMutateProb = 0
	mcfg.BiasMutateProb = 0
	mcfg.ToggleEnableProb = 0
	mcfg.ActivationMutateProb = 0

	rcfg := DefaultReproductionConfig()
	rcfg.Elitism = 1

	if err := pop.NextGeneration(mcfg, rcfg); err != nil {
		t.Fatalf("NextGeneration error: %v", err)
	}
	if len(pop.Genomes) != 4 {
		t.Fatalf("expected population size 4, got %d", len(pop.Genomes))
	}
	for i, g := range pop.Genomes {
		if g.Fitness != 0 {
			t.Fatalf("expected fitness reset at %d", i)
		}
		if _, err := BuildAcyclicPlan(g, nil, nil); err != nil {
			t.Fatalf("expected acyclic genome at %d: %v", i, err)
		}
	}
}
