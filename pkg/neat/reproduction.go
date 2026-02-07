package neat

import (
	"fmt"
	"math"
	"sort"
)

// ReproductionConfig controls selection and mating behavior.
type ReproductionConfig struct {
	SurvivalThreshold     float64
	Elitism               int
	CrossoverProb         float64
	InterspeciesMateProb  float64
}

// DefaultReproductionConfig returns common NEAT defaults.
func DefaultReproductionConfig() ReproductionConfig {
	return ReproductionConfig{
		SurvivalThreshold:    0.2,
		Elitism:              1,
		CrossoverProb:        0.75,
		InterspeciesMateProb: 0.001,
	}
}

// NextGeneration produces the next population by speciation, selection, and mutation.
func (p *Population) NextGeneration(mcfg MutationConfig, rcfg ReproductionConfig) error {
	if p == nil {
		return fmt.Errorf("population is nil")
	}
	if err := p.Speciate(); err != nil {
		return err
	}

	next, err := p.reproduce(mcfg, rcfg)
	if err != nil {
		return err
	}
	p.Genomes = next
	return nil
}

func (p *Population) reproduce(mcfg MutationConfig, rcfg ReproductionConfig) ([]Genome, error) {
	if p.RNG == nil {
		return nil, fmt.Errorf("rng is nil")
	}

	popSize := len(p.Genomes)
	if popSize == 0 {
		return nil, fmt.Errorf("population has no genomes")
	}

	speciesInfos := buildSpeciesInfo(p.Genomes, p.Species)
	offspringCounts := allocateOffspring(speciesInfos, popSize)

	next := make([]Genome, 0, popSize)
	for i, info := range speciesInfos {
		count := offspringCounts[i]
		if count <= 0 {
			continue
		}

		members := info.sortedMembers
		if len(members) == 0 {
			continue
		}

		elitism := minInt(rcfg.Elitism, len(members))
		if elitism > count {
			elitism = count
		}
		for e := 0; e < elitism; e++ {
			child := cloneGenome(p.Genomes[members[e]])
			child.Fitness = 0
			next = append(next, child)
		}

		remaining := count - elitism
		if remaining <= 0 {
			continue
		}

		survivors := survivorPool(members, rcfg.SurvivalThreshold)
		for k := 0; k < remaining; k++ {
			child, err := p.makeOffspring(survivors, i, rcfg)
			if err != nil {
				return nil, err
			}
			if err := mcfg.Mutate(p.RNG, &child, p.Tracker); err != nil {
				return nil, err
			}
			child.Fitness = 0
			next = append(next, child)
		}
	}

	if len(next) != popSize {
		return nil, fmt.Errorf("reproduction size mismatch: got %d want %d", len(next), popSize)
	}
	return next, nil
}

func (p *Population) makeOffspring(survivors []int, speciesIndex int, rcfg ReproductionConfig) (Genome, error) {
	if len(survivors) == 0 {
		return Genome{}, fmt.Errorf("no survivors available")
	}

	if randBool(p.RNG, rcfg.CrossoverProb) && len(survivors) > 1 {
		p1 := p.selectParent(survivors)
		p2 := p.selectMate(speciesIndex, survivors, rcfg)
		child, err := Crossover(p.RNG, p1, p2)
		if err != nil {
			return Genome{}, err
		}
		return child, nil
	}

	parent := p.selectParent(survivors)
	child := cloneGenome(parent)
	return child, nil
}

func (p *Population) selectMate(speciesIndex int, survivors []int, rcfg ReproductionConfig) Genome {
	if randBool(p.RNG, rcfg.InterspeciesMateProb) && len(p.Species) > 1 {
		other := p.pickOtherSpecies(speciesIndex)
		if other >= 0 {
			otherMembers := p.Species[other].Members
			if len(otherMembers) > 0 {
				return p.selectParent(otherMembers)
			}
		}
	}
	return p.selectParent(survivors)
}

func (p *Population) selectParent(indices []int) Genome {
	if len(indices) == 0 {
		return Genome{}
	}
	weights := make([]float64, len(indices))
	total := 0.0
	for i, idx := range indices {
		w := p.Genomes[idx].Fitness
		if w < 0 {
			w = 0
		}
		weights[i] = w
		total += w
	}
	if total == 0 {
		pick := indices[p.RNG.Intn(len(indices))]
		return p.Genomes[pick]
	}
	target := p.RNG.Float64() * total
	acc := 0.0
	for i, w := range weights {
		acc += w
		if acc >= target {
			return p.Genomes[indices[i]]
		}
	}
	return p.Genomes[indices[len(indices)-1]]
}

func (p *Population) pickOtherSpecies(current int) int {
	if len(p.Species) <= 1 {
		return -1
	}
	choices := make([]int, 0, len(p.Species)-1)
	for i := range p.Species {
		if i == current {
			continue
		}
		choices = append(choices, i)
	}
	return choices[p.RNG.Intn(len(choices))]
}

func survivorPool(sortedMembers []int, threshold float64) []int {
	if len(sortedMembers) == 0 {
		return nil
	}
	if threshold <= 0 {
		return []int{sortedMembers[0]}
	}
	count := int(math.Ceil(float64(len(sortedMembers)) * threshold))
	if count < 1 {
		count = 1
	}
	if count > len(sortedMembers) {
		count = len(sortedMembers)
	}
	return sortedMembers[:count]
}

type speciesInfo struct {
	index         int
	sortedMembers []int
	adjustedSum   float64
}

func buildSpeciesInfo(genomes []Genome, species []Species) []speciesInfo {
	infos := make([]speciesInfo, len(species))
	for i, s := range species {
		sorted := sortMembersByFitness(genomes, s.Members)
		adjusted := 0.0
		for _, idx := range sorted {
			adjusted += genomes[idx].Fitness / float64(len(s.Members))
		}
		infos[i] = speciesInfo{
			index:         i,
			sortedMembers: sorted,
			adjustedSum:   adjusted,
		}
	}
	return infos
}

func allocateOffspring(infos []speciesInfo, popSize int) []int {
	counts := make([]int, len(infos))
	if popSize == 0 {
		return counts
	}

	totalAdjusted := 0.0
	for _, info := range infos {
		totalAdjusted += info.adjustedSum
	}

	raw := make([]float64, len(infos))
	if totalAdjusted == 0 {
		share := float64(popSize) / float64(maxInt(1, len(infos)))
		for i := range infos {
			raw[i] = share
		}
	} else {
		for i, info := range infos {
			raw[i] = float64(popSize) * (info.adjustedSum / totalAdjusted)
		}
	}

	sum := 0
	fracs := make([]struct {
		idx  int
		frac float64
	}, len(raw))
	for i, val := range raw {
		counts[i] = int(math.Floor(val))
		sum += counts[i]
		fracs[i] = struct {
			idx  int
			frac float64
		}{idx: i, frac: val - float64(counts[i])}
	}

	remaining := popSize - sum
	sort.Slice(fracs, func(i, j int) bool { return fracs[i].frac > fracs[j].frac })
	for i := 0; i < remaining && i < len(fracs); i++ {
		counts[fracs[i].idx]++
	}
	return counts
}

func enforceAcyclic(nodes []NodeGene, conns []ConnectionGene) []ConnectionGene {
	if len(conns) == 0 {
		return conns
	}
	nodeMap := nodeGeneMap(nodes)
	for {
		if _, err := topoOrder(nodeMap, conns); err == nil {
			return conns
		}
		idx := -1
		var max InnovID = -1
		for i, c := range conns {
			if !c.Enabled {
				continue
			}
			if c.Innovation > max {
				max = c.Innovation
				idx = i
			}
		}
		if idx < 0 {
			return conns
		}
		conns[idx].Enabled = false
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
