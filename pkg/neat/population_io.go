package neat

import (
	"encoding/json"
	"io"
)

// SavePopulation writes genomes as JSON to the writer.
func SavePopulation(w io.Writer, genomes []Genome) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(genomes)
}

// LoadPopulation reads genomes from JSON.
func LoadPopulation(r io.Reader) ([]Genome, error) {
	dec := json.NewDecoder(r)
	var genomes []Genome
	if err := dec.Decode(&genomes); err != nil {
		return nil, err
	}
	return genomes, nil
}

// Save writes the population genomes as JSON.
func (p *Population) Save(w io.Writer) error {
	return SavePopulation(w, p.Genomes)
}

// LoadPopulationWithConfig loads genomes and constructs a population.
func LoadPopulationWithConfig(r io.Reader, rng RNG, cfg PopulationConfig) (*Population, error) {
	genomes, err := LoadPopulation(r)
	if err != nil {
		return nil, err
	}
	return NewPopulation(rng, cfg, genomes)
}
