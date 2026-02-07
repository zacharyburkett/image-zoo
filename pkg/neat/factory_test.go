package neat

import "testing"

func TestNewMinimalGenome(t *testing.T) {
	rng := NewRand(7)
	tracker, err := NewInnovationTracker(nil)
	if err != nil {
		t.Fatalf("NewInnovationTracker error: %v", err)
	}
	g, err := NewMinimalGenome(2, 1, ActivationSigmoid, rng, tracker, 1.0)
	if err != nil {
		t.Fatalf("NewMinimalGenome error: %v", err)
	}
	if len(g.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(g.Connections))
	}
	if g.Nodes[2].Kind != NodeOutput {
		t.Fatalf("expected output node")
	}
}
