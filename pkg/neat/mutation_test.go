package neat

import "testing"

func TestMutateAddNodeSplitsConnection(t *testing.T) {
	rng := NewRand(1)
	g := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 0.5, Enabled: true},
		},
	}
	tracker, err := NewInnovationTracker([]Genome{g})
	if err != nil {
		t.Fatalf("NewInnovationTracker error: %v", err)
	}

	if err := MutateAddNode(rng, &g, tracker, []ActivationType{ActivationSigmoid}); err != nil {
		t.Fatalf("MutateAddNode error: %v", err)
	}

	if len(g.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Connections) != 3 {
		t.Fatalf("expected 3 connections, got %d", len(g.Connections))
	}

	foundDisabled := false
	for _, c := range g.Connections {
		if c.Innovation == 1 && !c.Enabled {
			foundDisabled = true
			break
		}
	}
	if !foundDisabled {
		t.Fatalf("expected original connection to be disabled")
	}

	var newNode NodeGene
	foundNew := false
	for _, n := range g.Nodes {
		if n.Kind == NodeHidden {
			newNode = n
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Fatalf("expected a new hidden node")
	}

	var inToNew, newToOut bool
	for _, c := range g.Connections {
		if c.In == 1 && c.Out == newNode.ID && c.Enabled {
			inToNew = true
		}
		if c.In == newNode.ID && c.Out == 2 && c.Enabled {
			newToOut = true
		}
	}
	if !inToNew || !newToOut {
		t.Fatalf("expected split connections to be present")
	}
}

func TestMutateAddConnection(t *testing.T) {
	rng := NewRand(2)
	g := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 3, Kind: NodeOutput, Activation: ActivationLinear},
		},
	}
	tracker, err := NewInnovationTracker([]Genome{g})
	if err != nil {
		t.Fatalf("NewInnovationTracker error: %v", err)
	}

	if err := MutateAddConnection(rng, &g, tracker, 1.0, 10); err != nil {
		t.Fatalf("MutateAddConnection error: %v", err)
	}

	if len(g.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(g.Connections))
	}
	c := g.Connections[0]
	if c.In == c.Out {
		t.Fatalf("unexpected self connection")
	}
	if nodeByID := map[NodeID]NodeKind{1: NodeInput, 2: NodeInput, 3: NodeOutput}; nodeByID[c.Out] == NodeInput {
		t.Fatalf("connection targets input node")
	}
}

func TestMutateAddConnectionNoCandidates(t *testing.T) {
	rng := NewRand(3)
	g := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 3, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 3, Weight: 0.1, Enabled: true},
			{Innovation: 2, In: 2, Out: 3, Weight: 0.2, Enabled: true},
		},
	}
	tracker, err := NewInnovationTracker([]Genome{g})
	if err != nil {
		t.Fatalf("NewInnovationTracker error: %v", err)
	}

	if err := MutateAddConnection(rng, &g, tracker, 1.0, 5); err == nil {
		t.Fatalf("expected ErrNoConnectionCandidates")
	} else if err != ErrNoConnectionCandidates {
		t.Fatalf("expected ErrNoConnectionCandidates, got %v", err)
	}
}
