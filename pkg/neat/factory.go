package neat

import (
	"fmt"
	"sort"
)

// NewMinimalGenome creates a fully connected input->output genome.
// Inputs are assigned IDs starting at 1, outputs follow sequentially.
func NewMinimalGenome(inputCount, outputCount int, outputActivation ActivationType, rng RNG, tracker *InnovationTracker, weightRange float64) (Genome, error) {
	if inputCount <= 0 {
		return Genome{}, fmt.Errorf("inputCount must be > 0")
	}
	if outputCount <= 0 {
		return Genome{}, fmt.Errorf("outputCount must be > 0")
	}
	if rng == nil {
		return Genome{}, fmt.Errorf("rng is nil")
	}
	if tracker == nil {
		return Genome{}, fmt.Errorf("innovation tracker is nil")
	}

	nodes := make([]NodeGene, 0, inputCount+outputCount)
	for i := 0; i < inputCount; i++ {
		nodes = append(nodes, NodeGene{
			ID:         NodeID(i + 1),
			Kind:       NodeInput,
			Activation: ActivationLinear,
			Bias:       0,
		})
	}
	for i := 0; i < outputCount; i++ {
		nodes = append(nodes, NodeGene{
			ID:         NodeID(inputCount + i + 1),
			Kind:       NodeOutput,
			Activation: outputActivation,
			Bias:       0,
		})
	}

	conns := make([]ConnectionGene, 0, inputCount*outputCount)
	for in := 1; in <= inputCount; in++ {
		for out := 1; out <= outputCount; out++ {
			inID := NodeID(in)
			outID := NodeID(inputCount + out)
			innov := tracker.Innovation(inID, outID)
			weight := randRange(rng, -weightRange, weightRange)
			conns = append(conns, ConnectionGene{
				Innovation: innov,
				In:         inID,
				Out:        outID,
				Weight:     weight,
				Enabled:    true,
			})
		}
	}

	sort.Slice(conns, func(i, j int) bool { return conns[i].Innovation < conns[j].Innovation })
	return Genome{Nodes: nodes, Connections: conns}, nil
}
