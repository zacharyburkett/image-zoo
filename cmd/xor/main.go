package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/zacharyburkett/image-zoo/pkg/neat"
)

func main() {
	seed := flag.Int64("seed", 42, "random seed")
	popSize := flag.Int("pop", 150, "population size")
	maxGen := flag.Int("gen", 200, "max generations")
	target := flag.Float64("target", 3.9, "target fitness")
	flag.Parse()

	rng := neat.NewRand(*seed)
	initTracker, err := neat.NewInnovationTracker(nil)
	if err != nil {
		panic(err)
	}

	genomes := make([]neat.Genome, 0, *popSize)
	for i := 0; i < *popSize; i++ {
		g, err := neat.NewMinimalGenome(3, 1, neat.ActivationSigmoid, rng, initTracker, 1.0)
		if err != nil {
			panic(err)
		}
		genomes = append(genomes, g)
	}

	pcfg := neat.DefaultPopulationConfig()
	pcfg.CompatibilityThreshold = 3.0

	pop, err := neat.NewPopulation(rng, pcfg, genomes)
	if err != nil {
		panic(err)
	}

	mcfg := neat.DefaultMutationConfig()
	mcfg.AllowedActivations = []neat.ActivationType{
		neat.ActivationSigmoid,
	}
	// Match the original NEAT XOR setup more closely (bias node only).
	mcfg.BiasMutateProb = 0

	rcfg := neat.DefaultReproductionConfig()

	runner := neat.Runner{
		Population:  pop,
		Mutation:    mcfg,
		Reproduction: rcfg,
		Fitness:     xorFitness,
	}

	best := neat.Genome{}
	for gen := 0; gen < *maxGen; gen++ {
		currentBest, err := runner.Evaluate()
		if err != nil {
			panic(err)
		}
		best = currentBest
		if err := pop.Speciate(); err != nil {
			panic(err)
		}
		fmt.Printf("gen %d best=%.4f species=%d\n", gen, best.Fitness, len(pop.Species))
		if best.Fitness >= *target {
			break
		}
		if gen < *maxGen-1 {
			if err := pop.NextGeneration(mcfg, rcfg); err != nil {
				panic(err)
			}
		}
	}

	fmt.Printf("best fitness=%.4f\n", best.Fitness)
}

func xorFitness(g *neat.Genome) (float64, error) {
	plan, err := neat.BuildAcyclicPlan(*g, nil, nil)
	if err != nil {
		return 0, nil
	}

	tests := []struct {
		x1, x2 float64
		expect float64
	}{
		{0, 0, 0},
		{0, 1, 1},
		{1, 0, 1},
		{1, 1, 0},
	}

	sum := 0.0
	for _, tt := range tests {
		out, err := plan.Eval([]float64{tt.x1, tt.x2, 1})
		if err != nil {
			return 0, err
		}
		diff := tt.expect - out[0]
		sum += diff * diff
	}

	fitness := 4.0 - sum
	return math.Max(0, fitness), nil
}
