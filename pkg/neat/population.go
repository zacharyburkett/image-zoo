package neat

import "fmt"

// PopulationConfig controls speciation behavior.
type PopulationConfig struct {
	DistanceConfig
	CompatibilityThreshold float64
}

// DefaultPopulationConfig returns default speciation settings.
func DefaultPopulationConfig() PopulationConfig {
	cfg := DefaultDistanceConfig()
	return PopulationConfig{
		DistanceConfig:         cfg,
		CompatibilityThreshold: 3.0,
	}
}

// Species groups similar genomes.
type Species struct {
	ID            int
	Representative int
	Members       []int
}

// Population tracks genomes and species.
type Population struct {
	Config  PopulationConfig
	RNG     RNG
	Tracker *InnovationTracker
	Genomes []Genome
	Species []Species
}

// NewPopulation creates a population from genomes.
func NewPopulation(rng RNG, cfg PopulationConfig, genomes []Genome) (*Population, error) {
	if rng == nil {
		return nil, fmt.Errorf("rng is nil")
	}
	if len(genomes) == 0 {
		return nil, fmt.Errorf("no genomes provided")
	}
	tracker := NewInnovationTracker(genomes)
	return &Population{
		Config:  cfg,
		RNG:     rng,
		Tracker: tracker,
		Genomes: genomes,
	}, nil
}

// Speciate assigns genomes to species based on compatibility distance.
func (p *Population) Speciate() error {
	if p == nil {
		return fmt.Errorf("population is nil")
	}
	if len(p.Genomes) == 0 {
		return fmt.Errorf("population has no genomes")
	}

	p.Species = p.Species[:0]
	speciesID := 1

	for idx := range p.Genomes {
		placed := false
		for s := range p.Species {
			rep := p.Species[s].Representative
			distance := CompatibilityDistance(p.Genomes[idx], p.Genomes[rep], p.Config.DistanceConfig)
			if distance <= p.Config.CompatibilityThreshold {
				p.Species[s].Members = append(p.Species[s].Members, idx)
				placed = true
				break
			}
		}
		if !placed {
			p.Species = append(p.Species, Species{
				ID:            speciesID,
				Representative: idx,
				Members:       []int{idx},
			})
			speciesID++
		}
	}
	return nil
}
