package cppn

import (
	"testing"

	"github.com/zacharyburkett/image-zoo/pkg/neat"
)

func TestRenderGrayscale(t *testing.T) {
	spec := InputSpec{UseX: true}

	g := neat.Genome{
		Nodes: []neat.NodeGene{
			{ID: 1, Kind: neat.NodeInput, Activation: neat.ActivationLinear},
			{ID: 2, Kind: neat.NodeOutput, Activation: neat.ActivationLinear},
		},
		Connections: []neat.ConnectionGene{
			{Innovation: 1, In: 1, Out: 2, Weight: 1, Enabled: true},
		},
	}
	plan, err := neat.BuildAcyclicPlan(g, nil, nil)
	if err != nil {
		t.Fatalf("BuildAcyclicPlan error: %v", err)
	}

	pixels, err := RenderGrayscale(plan, 2, 2, spec)
	if err != nil {
		t.Fatalf("RenderGrayscale error: %v", err)
	}
	if len(pixels) != 2*2*4 {
		t.Fatalf("unexpected pixel length %d", len(pixels))
	}
}
