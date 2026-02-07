package neat

import "fmt"

// FitnessFunc evaluates a genome and returns its fitness.
type FitnessFunc func(*Genome) (float64, error)

// Runner executes the NEAT evolution loop.
type Runner struct {
	Population  *Population
	Mutation    MutationConfig
	Reproduction ReproductionConfig
	Fitness     FitnessFunc
}

// Evaluate computes fitness for the current population and returns the best genome.
func (r *Runner) Evaluate() (Genome, error) {
	if r == nil {
		return Genome{}, fmt.Errorf("runner is nil")
	}
	if r.Population == nil {
		return Genome{}, fmt.Errorf("population is nil")
	}
	if r.Fitness == nil {
		return Genome{}, fmt.Errorf("fitness function is nil")
	}
	if len(r.Population.Genomes) == 0 {
		return Genome{}, fmt.Errorf("population has no genomes")
	}

	var best Genome
	bestSet := false
	for i := range r.Population.Genomes {
		fitness, err := r.Fitness(&r.Population.Genomes[i])
		if err != nil {
			return Genome{}, err
		}
		r.Population.Genomes[i].Fitness = fitness
		if !bestSet || fitness > best.Fitness {
			best = cloneGenome(r.Population.Genomes[i])
			bestSet = true
		}
	}
	return best, nil
}

// Run evolves for up to maxGenerations and stops early at targetFitness.
// It returns the best genome and the generation it was found.
func (r *Runner) Run(maxGenerations int, targetFitness float64) (Genome, int, error) {
	if maxGenerations <= 0 {
		return Genome{}, 0, fmt.Errorf("maxGenerations must be > 0")
	}
	best := Genome{}
	for gen := 0; gen < maxGenerations; gen++ {
		currentBest, err := r.Evaluate()
		if err != nil {
			return Genome{}, gen, err
		}
		best = currentBest
		if best.Fitness >= targetFitness {
			return best, gen, nil
		}
		if gen == maxGenerations-1 {
			break
		}
		if err := r.Population.NextGeneration(r.Mutation, r.Reproduction); err != nil {
			return Genome{}, gen, err
		}
	}
	return best, maxGenerations - 1, nil
}
