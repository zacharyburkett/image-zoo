package neat

import "testing"

func TestBuildAcyclicPlanEval(t *testing.T) {
	g := Genome{
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

	plan, err := BuildAcyclicPlan(g, nil, nil)
	if err != nil {
		t.Fatalf("BuildAcyclicPlan error: %v", err)
	}

	out, err := plan.Eval([]float64{2, 3})
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 output, got %d", len(out))
	}
	if out[0] != 5 {
		t.Fatalf("expected output 5, got %v", out[0])
	}
}

func TestBuildAcyclicPlanCycle(t *testing.T) {
	g := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeHidden, Activation: ActivationLinear},
			{ID: 3, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 1.0, Enabled: true},
			{Innovation: 2, In: 2, Out: 3, Weight: 1.0, Enabled: true},
			{Innovation: 3, In: 3, Out: 2, Weight: 1.0, Enabled: true},
		},
	}

	_, err := BuildAcyclicPlan(g, nil, nil)
	if err == nil {
		t.Fatalf("expected cycle error, got nil")
	}
}

func TestBuildAcyclicPlanDisabledConn(t *testing.T) {
	g := Genome{
		Nodes: []NodeGene{
			{ID: 1, Kind: NodeInput, Activation: ActivationLinear},
			{ID: 2, Kind: NodeOutput, Activation: ActivationLinear},
		},
		Connections: []ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 2.0, Enabled: false},
		},
	}

	plan, err := BuildAcyclicPlan(g, nil, nil)
	if err != nil {
		t.Fatalf("BuildAcyclicPlan error: %v", err)
	}

	out, err := plan.Eval([]float64{7})
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if out[0] != 0 {
		t.Fatalf("expected output 0 for disabled conn, got %v", out[0])
	}
}
