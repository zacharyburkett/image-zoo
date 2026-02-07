package neat

import "testing"

func TestRunnerRunStopsAtTarget(t *testing.T) {
	rng := NewRand(11)
	tracker, err := NewInnovationTracker(nil)
	if err != nil {
		t.Fatalf("NewInnovationTracker error: %v", err)
	}
	g, err := NewMinimalGenome(1, 1, ActivationSigmoid, rng, tracker, 1.0)
	if err != nil {
		t.Fatalf("NewMinimalGenome error: %v", err)
	}

	pop, err := NewPopulation(rng, DefaultPopulationConfig(), []Genome{g})
	if err != nil {
		t.Fatalf("NewPopulation error: %v", err)
	}

	runner := Runner{
		Population:  pop,
		Mutation:    DefaultMutationConfig(),
		Reproduction: DefaultReproductionConfig(),
		Fitness: func(*Genome) (float64, error) {
			return 1.0, nil
		},
	}

	best, gen, err := runner.Run(5, 1.0)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if gen != 0 {
		t.Fatalf("expected generation 0, got %d", gen)
	}
	if best.Fitness != 1.0 {
		t.Fatalf("expected fitness 1.0, got %v", best.Fitness)
	}
}
