package neat

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGenomeJSONRoundTrip(t *testing.T) {
	g := Genome{
		Fitness: 1.25,
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear, Bias: 0},
			{ID: 2, Kind: NodeOutput, Activation: ActivationSigmoid, Bias: 0.5},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 0.75, Enabled: true},
		},
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(data), "sigmoid") {
		t.Fatalf("expected activation names to be encoded as strings, got %s", string(data))
	}

	var decoded Genome
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Fitness != g.Fitness {
		t.Fatalf("fitness mismatch: got %v want %v", decoded.Fitness, g.Fitness)
	}
	if len(decoded.Nodes) != len(g.Nodes) || len(decoded.Connections) != len(g.Connections) {
		t.Fatalf("decoded length mismatch")
	}
	if decoded.Nodes[1].Activation != g.Nodes[1].Activation {
		t.Fatalf("activation mismatch: got %v want %v", decoded.Nodes[1].Activation, g.Nodes[1].Activation)
	}
}

func TestActivationTypeJSONNumber(t *testing.T) {
	var a ActivationType
	if err := json.Unmarshal([]byte("2"), &a); err != nil {
		t.Fatalf("unmarshal numeric activation error: %v", err)
	}
	if a != ActivationTanh {
		t.Fatalf("expected ActivationTanh, got %v", a)
	}
}
