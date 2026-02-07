package neat

import (
	"testing"
)

func TestCrossoverFitterParentDominatesDisjoint(t *testing.T) {
	rng := NewRand(42)

	parentA := Genome{
		Fitness: 2.0,
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 3, Kind: NodeHidden, Activation: ActivationLinear},
			{ID: 4, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 3, Weight: 1.0, Enabled: true},
			{Innovation: 2, In: 2, Out: 3, Weight: 1.0, Enabled: true},
			{Innovation: 3, In: 3, Out: 4, Weight: 1.0, Enabled: true},
		},
	}

	parentB := Genome{
		Fitness: 1.0,
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 3, Kind: NodeHidden, Activation: ActivationLinear},
			{ID: 4, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 3, Weight: 2.0, Enabled: true},
			{Innovation: 2, In: 2, Out: 3, Weight: 2.0, Enabled: true},
			{Innovation: 4, In: 2, Out: 4, Weight: 2.0, Enabled: true},
		},
	}

	child, err := Crossover(rng, parentA, parentB)
	if err != nil {
		t.Fatalf("Crossover error: %v", err)
	}

	innovs := make(map[InnovID]struct{})
	for _, c := range child.Connections {
		innovs[c.Innovation] = struct{}{}
	}

	if _, ok := innovs[3]; !ok {
		t.Fatalf("expected child to include innov 3 from fitter parent")
	}
	if _, ok := innovs[4]; ok {
		t.Fatalf("did not expect child to include innov 4 from less fit parent")
	}

	nodeIDs := make(map[NodeID]struct{})
	for _, n := range child.Nodes {
		nodeIDs[n.ID] = struct{}{}
	}
	for _, c := range child.Connections {
		if _, ok := nodeIDs[c.In]; !ok {
			t.Fatalf("missing node %d for connection", c.In)
		}
		if _, ok := nodeIDs[c.Out]; !ok {
			t.Fatalf("missing node %d for connection", c.Out)
		}
	}
}
