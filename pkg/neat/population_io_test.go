package neat

import (
	"bytes"
	"testing"
)

func TestPopulationSaveLoad(t *testing.T) {
	g := Genome{
		Fitness: 1.0,
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationSigmoid},
		},
		Connections: []ConnectionGene{{Innovation: 1, In: 1, Out: 2, Weight: 0.5, Enabled: true}},
	}
	buf := &bytes.Buffer{}
	if err := SavePopulation(buf, []Genome{g}); err != nil {
		t.Fatalf("SavePopulation error: %v", err)
	}

	loaded, err := LoadPopulation(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("LoadPopulation error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 genome, got %d", len(loaded))
	}
	if loaded[0].Fitness != g.Fitness {
		t.Fatalf("fitness mismatch")
	}
}
